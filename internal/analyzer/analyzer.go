package analyzer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"commit-archaeologist/internal/ast"
	"commit-archaeologist/internal/db"
	"commit-archaeologist/internal/git"
)

type ProgressCallback func(stage string, current, total int, message string)

type Analyzer struct {
	store   *db.Store
	gitExec *git.Executor
	astEng  *ast.Engine
}

func NewAnalyzer(store *db.Store, gitExec *git.Executor) *Analyzer {
	return &Analyzer{
		store:   store,
		gitExec: gitExec,
		astEng:  ast.NewEngine(),
	}
}

// GenerateRepoID generates a stable deterministic ID for a repository path.
func GenerateRepoID(repoPath string) string {
	cleaned := strings.ToLower(filepath.ToSlash(filepath.Clean(repoPath)))
	h := sha256.Sum256([]byte(cleaned))
	return hex.EncodeToString(h[:8])
}

// AnalyzeRepository performs the full end-to-end analysis of a local git repository.
func (a *Analyzer) AnalyzeRepository(ctx context.Context, repoPath string, isDemo bool, onProgress ProgressCallback) (*db.OverviewStats, error) {
	cleanPath, err := git.ValidateRepoPath(repoPath)
	if err != nil {
		return nil, err
	}

	repoName := filepath.Base(cleanPath)
	repoID := GenerateRepoID(cleanPath)

	if onProgress != nil {
		onProgress("inspecting", 1, 5, "Inspecting repository and branches...")
	}

	defaultBranch, err := a.gitExec.GetDefaultBranch(ctx, cleanPath)
	if err != nil {
		defaultBranch = "main"
	}

	if onProgress != nil {
		onProgress("commits", 2, 5, "Extracting commit logs and change sets...")
	}

	commits, err := a.gitExec.ParseLog(ctx, cleanPath, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze commits: %w", err)
	}

	// Set repo ID on commits and files
	for i := range commits {
		commits[i].RepoID = repoID
		commits[i].Branch = defaultBranch
		for j := range commits[i].Files {
			commits[i].Files[j].RepoID = repoID
		}
	}

	if onProgress != nil {
		onProgress("files", 3, 5, "Scanning tracked files and AST dependencies...")
	}

	trackedFiles, err := a.gitExec.ListTrackedFiles(ctx, cleanPath)
	if err != nil {
		// Fallback: extract tracked files from commit files
		fileMap := make(map[string]bool)
		for _, c := range commits {
			for _, f := range c.Files {
				if f.ChangeType != "D" {
					fileMap[f.FilePath] = true
				}
			}
		}
		for f := range fileMap {
			trackedFiles = append(trackedFiles, f)
		}
	}

	// Parse AST dependencies and LOC
	dependencies, locMap, err := a.astEng.ParseRepoFiles(ctx, repoID, cleanPath, trackedFiles)
	if err != nil {
		// Non-fatal, proceed with empty dependencies
		dependencies = []db.DependencyEdge{}
		locMap = make(map[string]int)
	}

	if onProgress != nil {
		onProgress("calculating", 4, 5, "Calculating hotspots and contributor statistics...")
	}

	// Calculate Hotspots
	hotspots := a.calculateHotspots(repoID, commits, trackedFiles, locMap)

	// Calculate Contributors & Activity Punchcard
	contributors, activities := a.calculateContributors(repoID, commits)

	// Save all to SQLite in a transactional manner
	if onProgress != nil {
		onProgress("persisting", 5, 5, "Persisting repository intelligence to local SQLite...")
	}

	now := time.Now().UTC()
	repoMeta := &db.Repository{
		ID:             repoID,
		Name:           repoName,
		Path:           cleanPath,
		DefaultBranch:  defaultBranch,
		LastAnalyzedAt: &now,
		TotalCommits:   len(commits),
		TotalFiles:     len(trackedFiles),
		TotalAuthors:   len(contributors),
		IsDemo:         isDemo,
	}

	// Clear previous data for this repo to ensure clean update
	_ = a.store.ClearRepoData(ctx, repoID)

	if err := a.store.UpsertRepository(ctx, repoMeta); err != nil {
		return nil, fmt.Errorf("failed to save repository: %w", err)
	}
	if err := a.store.SaveCommits(ctx, commits); err != nil {
		return nil, fmt.Errorf("failed to save commits: %w", err)
	}
	if err := a.store.SaveHotspots(ctx, hotspots); err != nil {
		return nil, fmt.Errorf("failed to save hotspots: %w", err)
	}
	if err := a.store.SaveDependencies(ctx, dependencies); err != nil {
		return nil, fmt.Errorf("failed to save dependencies: %w", err)
	}
	if err := a.store.SaveContributors(ctx, contributors, activities); err != nil {
		return nil, fmt.Errorf("failed to save contributors: %w", err)
	}

	return a.store.GetOverviewStats(ctx, repoID)
}

type fileStatsAccumulator struct {
	commitCount int
	authors     map[string]bool
	insertions  int
	deletions   int
	lastMod     time.Time
}

func (a *Analyzer) calculateHotspots(repoID string, commits []db.Commit, trackedFiles []string, locMap map[string]int) []db.FileHotspot {
	stats := make(map[string]*fileStatsAccumulator)

	for _, c := range commits {
		for _, f := range c.Files {
			path := f.FilePath
			acc, exists := stats[path]
			if !exists {
				acc = &fileStatsAccumulator{
					authors: make(map[string]bool),
					lastMod: c.AuthorDate,
				}
				stats[path] = acc
			}

			acc.commitCount++
			acc.authors[c.AuthorEmail] = true
			acc.insertions += f.Insertions
			acc.deletions += f.Deletions
			if c.AuthorDate.After(acc.lastMod) {
				acc.lastMod = c.AuthorDate
			}
		}
	}

	var hotspots []db.FileHotspot
	for path, acc := range stats {
		authorCount := len(acc.authors)
		loc := locMap[path]
		ext := strings.ToLower(filepath.Ext(path))

		// Churn formula: weighted combination of commit frequency, author breadth, and line churn magnitude
		totalChurn := float64(acc.insertions + acc.deletions)
		churnScore := (float64(acc.commitCount) * 1.5) + (float64(authorCount) * 2.0) + (math.Log2(1.0+totalChurn) * 1.2)

		hotspots = append(hotspots, db.FileHotspot{
			RepoID:          repoID,
			FilePath:        path,
			Extension:       ext,
			CommitCount:     acc.commitCount,
			AuthorCount:     authorCount,
			TotalInsertions: acc.insertions,
			TotalDeletions:  acc.deletions,
			LinesOfCode:     loc,
			LastModified:    acc.lastMod,
			ChurnScore:      math.Round(churnScore*100) / 100,
		})
	}

	return hotspots
}

type contribAccumulator struct {
	name       string
	email      string
	commits    int
	firstDate  time.Time
	lastDate   time.Time
	insertions int
	deletions  int
}

func (a *Analyzer) calculateContributors(repoID string, commits []db.Commit) ([]db.Contributor, []db.ContributorActivity) {
	cMap := make(map[string]*contribAccumulator)
	actMap := make(map[string]int) // "email:day:hour" -> count

	for _, c := range commits {
		email := strings.ToLower(strings.TrimSpace(c.AuthorEmail))
		if email == "" {
			email = "unknown@local"
		}

		acc, exists := cMap[email]
		if !exists {
			acc = &contribAccumulator{
				name:      c.AuthorName,
				email:     email,
				firstDate: c.AuthorDate,
				lastDate:  c.AuthorDate,
			}
			cMap[email] = acc
		}

		acc.commits++
		acc.insertions += c.Insertions
		acc.deletions += c.Deletions

		if c.AuthorDate.Before(acc.firstDate) {
			acc.firstDate = c.AuthorDate
		}
		if c.AuthorDate.After(acc.lastDate) {
			acc.lastDate = c.AuthorDate
		}

		// Punch card activity: day of week (0=Sunday to 6=Saturday), hour (0-23)
		day := int(c.AuthorDate.Weekday())
		hour := c.AuthorDate.Hour()
		key := fmt.Sprintf("%s:%d:%d", email, day, hour)
		actMap[key]++
	}

	var contributors []db.Contributor
	for _, acc := range cMap {
		contributors = append(contributors, db.Contributor{
			RepoID:          repoID,
			Name:            acc.name,
			Email:           acc.email,
			TotalCommits:    acc.commits,
			FirstCommitDate: acc.firstDate,
			LastCommitDate:  acc.lastDate,
			Insertions:      acc.insertions,
			Deletions:       acc.deletions,
		})
	}

	var activities []db.ContributorActivity
	for key, count := range actMap {
		var email string
		var day, hour int
		parts := strings.Split(key, ":")
		if len(parts) == 3 {
			email = parts[0]
			fmt.Sscanf(parts[1], "%d", &day)
			fmt.Sscanf(parts[2], "%d", &hour)
			activities = append(activities, db.ContributorActivity{
				RepoID:      repoID,
				Email:       email,
				DayOfWeek:   day,
				HourOfDay:   hour,
				CommitCount: count,
			})
		}
	}

	return contributors, activities
}
