package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	db *sql.DB
}

type Repository struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Path           string     `json:"path"`
	DefaultBranch  string     `json:"defaultBranch"`
	LastAnalyzedAt *time.Time `json:"lastAnalyzedAt"`
	TotalCommits   int        `json:"totalCommits"`
	TotalFiles     int        `json:"totalFiles"`
	TotalAuthors   int        `json:"totalAuthors"`
	IsDemo         bool       `json:"isDemo"`
}

type Commit struct {
	Hash              string    `json:"hash"`
	RepoID            string    `json:"repoId"`
	ShortHash         string    `json:"shortHash"`
	ParentHashes      string    `json:"parentHashes"`
	AuthorName        string    `json:"authorName"`
	AuthorEmail       string    `json:"authorEmail"`
	AuthorDate        time.Time `json:"authorDate"`
	CommitterDate     time.Time `json:"committerDate"`
	Subject           string    `json:"subject"`
	Body              string    `json:"body"`
	Branch            string    `json:"branch"`
	Insertions        int       `json:"insertions"`
	Deletions         int       `json:"deletions"`
	FilesChangedCount int       `json:"filesChangedCount"`
	Files             []CommitFile `json:"files,omitempty"`
}

type CommitFile struct {
	ID         int64  `json:"id"`
	CommitHash string `json:"commitHash"`
	RepoID     string `json:"repoId"`
	FilePath   string `json:"filePath"`
	ChangeType string `json:"changeType"`
	OldPath    string `json:"oldPath,omitempty"`
	Insertions int    `json:"insertions"`
	Deletions  int    `json:"deletions"`
}

type FileHotspot struct {
	RepoID          string    `json:"repoId"`
	FilePath        string    `json:"filePath"`
	Extension       string    `json:"extension"`
	CommitCount     int       `json:"commitCount"`
	AuthorCount     int       `json:"authorCount"`
	TotalInsertions int       `json:"totalInsertions"`
	TotalDeletions  int       `json:"totalDeletions"`
	LinesOfCode     int       `json:"linesOfCode"`
	LastModified    time.Time `json:"lastModified"`
	ChurnScore      float64   `json:"churnScore"`
}

type DependencyEdge struct {
	ID              int64  `json:"id"`
	RepoID          string `json:"repoId"`
	SourceFile      string `json:"sourceFile"`
	TargetFile      string `json:"targetFile"`
	ImportStatement string `json:"importStatement"`
	Language        string `json:"language"`
	LineNumber      int    `json:"lineNumber"`
}

type DependencyGraph struct {
	Nodes []DependencyNode `json:"nodes"`
	Links []DependencyLink `json:"links"`
}

type DependencyNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Language string `json:"language"`
	InDegree int    `json:"inDegree"`
	OutDegree int   `json:"outDegree"`
}

type DependencyLink struct {
	Source    string `json:"source"`
	Target    string `json:"target"`
	Statement string `json:"statement"`
	Language  string `json:"language"`
}

type Contributor struct {
	RepoID           string    `json:"repoId"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	TotalCommits     int       `json:"totalCommits"`
	FirstCommitDate  time.Time `json:"firstCommitDate"`
	LastCommitDate   time.Time `json:"lastCommitDate"`
	Insertions       int       `json:"insertions"`
	Deletions        int       `json:"deletions"`
}

type ContributorActivity struct {
	RepoID      string `json:"repoId"`
	Email       string `json:"email"`
	DayOfWeek   int    `json:"dayOfWeek"`
	HourOfDay   int    `json:"hourOfDay"`
	CommitCount int    `json:"commitCount"`
}

type TimelinePoint struct {
	Date        string `json:"date"`
	CommitCount int    `json:"commitCount"`
	Insertions  int    `json:"insertions"`
	Deletions   int    `json:"deletions"`
	Authors     int    `json:"authors"`
}

type OverviewStats struct {
	Repository         Repository      `json:"repository"`
	TopHotspots        []FileHotspot   `json:"topHotspots"`
	TopContributors    []Contributor   `json:"topContributors"`
	RecentCommits      []Commit        `json:"recentCommits"`
	LanguageBreakdown  map[string]int  `json:"languageBreakdown"`
	Timeline           []TimelinePoint `json:"timeline"`
	TotalDependencies  int             `json:"totalDependencies"`
}

type CommitFilter struct {
	RepoID    string
	Author    string
	Search    string
	FilePath  string
	StartDate *time.Time
	EndDate   *time.Time
	Branch    string
	Limit     int
	Offset    int
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Apply schema migrations
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) UpsertRepository(ctx context.Context, repo *Repository) error {
	query := `
		INSERT INTO repositories (id, name, path, default_branch, last_analyzed_at, total_commits, total_files, total_authors, is_demo)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			path = excluded.path,
			default_branch = excluded.default_branch,
			last_analyzed_at = excluded.last_analyzed_at,
			total_commits = excluded.total_commits,
			total_files = excluded.total_files,
			total_authors = excluded.total_authors,
			is_demo = excluded.is_demo;
	`
	_, err := s.db.ExecContext(ctx, query,
		repo.ID, repo.Name, repo.Path, repo.DefaultBranch, repo.LastAnalyzedAt,
		repo.TotalCommits, repo.TotalFiles, repo.TotalAuthors, repo.IsDemo,
	)
	return err
}

func (s *Store) GetRepository(ctx context.Context, id string) (*Repository, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name, path, default_branch, last_analyzed_at, total_commits, total_files, total_authors, is_demo FROM repositories WHERE id = ?`, id)
	var repo Repository
	err := row.Scan(&repo.ID, &repo.Name, &repo.Path, &repo.DefaultBranch, &repo.LastAnalyzedAt, &repo.TotalCommits, &repo.TotalFiles, &repo.TotalAuthors, &repo.IsDemo)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (s *Store) ListRepositories(ctx context.Context) ([]Repository, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, path, default_branch, last_analyzed_at, total_commits, total_files, total_authors, is_demo FROM repositories ORDER BY last_analyzed_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []Repository
	for rows.Next() {
		var r Repository
		if err := rows.Scan(&r.ID, &r.Name, &r.Path, &r.DefaultBranch, &r.LastAnalyzedAt, &r.TotalCommits, &r.TotalFiles, &r.TotalAuthors, &r.IsDemo); err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, nil
}

func (s *Store) ClearRepoData(ctx context.Context, repoID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM commit_files WHERE repo_id = ?`, repoID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM commits WHERE repo_id = ?`, repoID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM file_hotspots WHERE repo_id = ?`, repoID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM dependencies WHERE repo_id = ?`, repoID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM contributors WHERE repo_id = ?`, repoID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM contributor_activity WHERE repo_id = ?`, repoID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) SaveCommits(ctx context.Context, commits []Commit) error {
	if len(commits) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmtCommit, err := tx.PrepareContext(ctx, `
		INSERT INTO commits (hash, repo_id, short_hash, parent_hashes, author_name, author_email, author_date, committer_date, subject, body, branch, insertions, deletions, files_changed_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(hash) DO UPDATE SET
			subject = excluded.subject,
			body = excluded.body,
			insertions = excluded.insertions,
			deletions = excluded.deletions,
			files_changed_count = excluded.files_changed_count;
	`)
	if err != nil {
		return err
	}
	defer stmtCommit.Close()

	stmtFile, err := tx.PrepareContext(ctx, `
		INSERT INTO commit_files (commit_hash, repo_id, file_path, change_type, old_path, insertions, deletions)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmtFile.Close()

	for _, c := range commits {
		_, err = stmtCommit.ExecContext(ctx,
			c.Hash, c.RepoID, c.ShortHash, c.ParentHashes, c.AuthorName, c.AuthorEmail,
			c.AuthorDate, c.CommitterDate, c.Subject, c.Body, c.Branch,
			c.Insertions, c.Deletions, c.FilesChangedCount,
		)
		if err != nil {
			return fmt.Errorf("failed to save commit %s: %w", c.Hash, err)
		}

		for _, f := range c.Files {
			_, err = stmtFile.ExecContext(ctx,
				c.Hash, c.RepoID, f.FilePath, f.ChangeType, f.OldPath, f.Insertions, f.Deletions,
			)
			if err != nil {
				return fmt.Errorf("failed to save commit file %s:%s: %w", c.Hash, f.FilePath, err)
			}
		}
	}

	return tx.Commit()
}

func (s *Store) SaveHotspots(ctx context.Context, hotspots []FileHotspot) error {
	if len(hotspots) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO file_hotspots (repo_id, file_path, extension, commit_count, author_count, total_insertions, total_deletions, lines_of_code, last_modified, churn_score)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(repo_id, file_path) DO UPDATE SET
			commit_count = excluded.commit_count,
			author_count = excluded.author_count,
			total_insertions = excluded.total_insertions,
			total_deletions = excluded.total_deletions,
			lines_of_code = excluded.lines_of_code,
			last_modified = excluded.last_modified,
			churn_score = excluded.churn_score;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, h := range hotspots {
		_, err = stmt.ExecContext(ctx,
			h.RepoID, h.FilePath, h.Extension, h.CommitCount, h.AuthorCount,
			h.TotalInsertions, h.TotalDeletions, h.LinesOfCode, h.LastModified, h.ChurnScore,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) SaveDependencies(ctx context.Context, edges []DependencyEdge) error {
	if len(edges) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO dependencies (repo_id, source_file, target_file, import_statement, language, line_number)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range edges {
		_, err = stmt.ExecContext(ctx, e.RepoID, e.SourceFile, e.TargetFile, e.ImportStatement, e.Language, e.LineNumber)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) SaveContributors(ctx context.Context, contributors []Contributor, activities []ContributorActivity) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmtContrib, err := tx.PrepareContext(ctx, `
		INSERT INTO contributors (repo_id, name, email, total_commits, first_commit_date, last_commit_date, insertions, deletions)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(repo_id, email) DO UPDATE SET
			name = excluded.name,
			total_commits = excluded.total_commits,
			first_commit_date = excluded.first_commit_date,
			last_commit_date = excluded.last_commit_date,
			insertions = excluded.insertions,
			deletions = excluded.deletions;
	`)
	if err != nil {
		return err
	}
	defer stmtContrib.Close()

	for _, c := range contributors {
		_, err = stmtContrib.ExecContext(ctx, c.RepoID, c.Name, c.Email, c.TotalCommits, c.FirstCommitDate, c.LastCommitDate, c.Insertions, c.Deletions)
		if err != nil {
			return err
		}
	}

	stmtAct, err := tx.PrepareContext(ctx, `
		INSERT INTO contributor_activity (repo_id, email, day_of_week, hour_of_day, commit_count)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(repo_id, email, day_of_week, hour_of_day) DO UPDATE SET
			commit_count = excluded.commit_count;
	`)
	if err != nil {
		return err
	}
	defer stmtAct.Close()

	for _, a := range activities {
		_, err = stmtAct.ExecContext(ctx, a.RepoID, a.Email, a.DayOfWeek, a.HourOfDay, a.CommitCount)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) QueryCommits(ctx context.Context, filter CommitFilter) ([]Commit, error) {
	var sb strings.Builder
	var args []interface{}

	sb.WriteString(`
		SELECT c.hash, c.repo_id, c.short_hash, c.parent_hashes, c.author_name, c.author_email,
		       c.author_date, c.committer_date, c.subject, c.body, c.branch, c.insertions, c.deletions, c.files_changed_count
		FROM commits c
	`)

	if filter.FilePath != "" {
		sb.WriteString(` INNER JOIN commit_files cf ON c.hash = cf.commit_hash `)
	}

	sb.WriteString(` WHERE c.repo_id = ? `)
	args = append(args, filter.RepoID)

	if filter.Author != "" {
		sb.WriteString(` AND (c.author_name LIKE ? OR c.author_email LIKE ?) `)
		pattern := "%" + filter.Author + "%"
		args = append(args, pattern, pattern)
	}
	if filter.Search != "" {
		sb.WriteString(` AND (c.subject LIKE ? OR c.body LIKE ? OR c.hash LIKE ?) `)
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern, pattern)
	}
	if filter.FilePath != "" {
		sb.WriteString(` AND cf.file_path LIKE ? `)
		args = append(args, "%"+filter.FilePath+"%")
	}
	if filter.StartDate != nil {
		sb.WriteString(` AND c.author_date >= ? `)
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		sb.WriteString(` AND c.author_date <= ? `)
		args = append(args, *filter.EndDate)
	}
	if filter.Branch != "" {
		sb.WriteString(` AND c.branch = ? `)
		args = append(args, filter.Branch)
	}

	sb.WriteString(` GROUP BY c.hash ORDER BY c.author_date DESC `)

	if filter.Limit > 0 {
		sb.WriteString(` LIMIT ? `)
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			sb.WriteString(` OFFSET ? `)
			args = append(args, filter.Offset)
		}
	} else {
		sb.WriteString(` LIMIT 100 `)
	}

	rows, err := s.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []Commit
	for rows.Next() {
		var c Commit
		if err := rows.Scan(
			&c.Hash, &c.RepoID, &c.ShortHash, &c.ParentHashes, &c.AuthorName, &c.AuthorEmail,
			&c.AuthorDate, &c.CommitterDate, &c.Subject, &c.Body, &c.Branch,
			&c.Insertions, &c.Deletions, &c.FilesChangedCount,
		); err != nil {
			return nil, err
		}
		commits = append(commits, c)
	}

	return commits, nil
}

func (s *Store) GetCommitDetails(ctx context.Context, hash string) (*Commit, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT hash, repo_id, short_hash, parent_hashes, author_name, author_email,
		       author_date, committer_date, subject, body, branch, insertions, deletions, files_changed_count
		FROM commits WHERE hash = ?
	`, hash)

	var c Commit
	if err := row.Scan(
		&c.Hash, &c.RepoID, &c.ShortHash, &c.ParentHashes, &c.AuthorName, &c.AuthorEmail,
		&c.AuthorDate, &c.CommitterDate, &c.Subject, &c.Body, &c.Branch,
		&c.Insertions, &c.Deletions, &c.FilesChangedCount,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	fRows, err := s.db.QueryContext(ctx, `
		SELECT id, commit_hash, repo_id, file_path, change_type, old_path, insertions, deletions
		FROM commit_files WHERE commit_hash = ? ORDER BY file_path ASC
	`, hash)
	if err != nil {
		return nil, err
	}
	defer fRows.Close()

	for fRows.Next() {
		var f CommitFile
		if err := fRows.Scan(&f.ID, &f.CommitHash, &f.RepoID, &f.FilePath, &f.ChangeType, &f.OldPath, &f.Insertions, &f.Deletions); err != nil {
			return nil, err
		}
		c.Files = append(c.Files, f)
	}

	return &c, nil
}

func (s *Store) GetHotspots(ctx context.Context, repoID string, limit int) ([]FileHotspot, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT repo_id, file_path, extension, commit_count, author_count,
		       total_insertions, total_deletions, lines_of_code, last_modified, churn_score
		FROM file_hotspots
		WHERE repo_id = ?
		ORDER BY churn_score DESC
		LIMIT ?
	`, repoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hotspots []FileHotspot
	for rows.Next() {
		var h FileHotspot
		if err := rows.Scan(
			&h.RepoID, &h.FilePath, &h.Extension, &h.CommitCount, &h.AuthorCount,
			&h.TotalInsertions, &h.TotalDeletions, &h.LinesOfCode, &h.LastModified, &h.ChurnScore,
		); err != nil {
			return nil, err
		}
		hotspots = append(hotspots, h)
	}
	return hotspots, nil
}

func (s *Store) GetDependencyGraph(ctx context.Context, repoID string) (*DependencyGraph, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, repo_id, source_file, target_file, import_statement, language, line_number
		FROM dependencies
		WHERE repo_id = ?
	`, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeMap := make(map[string]*DependencyNode)
	var links []DependencyLink

	for rows.Next() {
		var e DependencyEdge
		if err := rows.Scan(&e.ID, &e.RepoID, &e.SourceFile, &e.TargetFile, &e.ImportStatement, &e.Language, &e.LineNumber); err != nil {
			return nil, err
		}

		if _, exists := nodeMap[e.SourceFile]; !exists {
			nodeMap[e.SourceFile] = &DependencyNode{
				ID:       e.SourceFile,
				Name:     e.SourceFile,
				Language: e.Language,
			}
		}
		if _, exists := nodeMap[e.TargetFile]; !exists {
			nodeMap[e.TargetFile] = &DependencyNode{
				ID:       e.TargetFile,
				Name:     e.TargetFile,
				Language: e.Language,
			}
		}

		nodeMap[e.SourceFile].OutDegree++
		nodeMap[e.TargetFile].InDegree++

		links = append(links, DependencyLink{
			Source:    e.SourceFile,
			Target:    e.TargetFile,
			Statement: e.ImportStatement,
			Language:  e.Language,
		})
	}

	nodes := make([]DependencyNode, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, *n)
	}

	return &DependencyGraph{Nodes: nodes, Links: links}, nil
}

func (s *Store) GetContributors(ctx context.Context, repoID string) ([]Contributor, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT repo_id, name, email, total_commits, first_commit_date, last_commit_date, insertions, deletions
		FROM contributors
		WHERE repo_id = ?
		ORDER BY total_commits DESC
	`, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contributors []Contributor
	for rows.Next() {
		var c Contributor
		if err := rows.Scan(&c.RepoID, &c.Name, &c.Email, &c.TotalCommits, &c.FirstCommitDate, &c.LastCommitDate, &c.Insertions, &c.Deletions); err != nil {
			return nil, err
		}
		contributors = append(contributors, c)
	}
	return contributors, nil
}

func (s *Store) GetContributorActivity(ctx context.Context, repoID string, email string) ([]ContributorActivity, error) {
	var rows *sql.Rows
	var err error

	if email != "" {
		rows, err = s.db.QueryContext(ctx, `
			SELECT repo_id, email, day_of_week, hour_of_day, commit_count
			FROM contributor_activity
			WHERE repo_id = ? AND email = ?
			ORDER BY day_of_week, hour_of_day
		`, repoID, email)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT repo_id, 'all' as email, day_of_week, hour_of_day, SUM(commit_count) as commit_count
			FROM contributor_activity
			WHERE repo_id = ?
			GROUP BY day_of_week, hour_of_day
			ORDER BY day_of_week, hour_of_day
		`, repoID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []ContributorActivity
	for rows.Next() {
		var a ContributorActivity
		if err := rows.Scan(&a.RepoID, &a.Email, &a.DayOfWeek, &a.HourOfDay, &a.CommitCount); err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}

func (s *Store) GetTimeline(ctx context.Context, repoID string) ([]TimelinePoint, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT COALESCE(substr(author_date, 1, 10), 'unknown') as day,
		       COUNT(DISTINCT hash) as commit_count,
		       COALESCE(SUM(insertions), 0) as total_insertions,
		       COALESCE(SUM(deletions), 0) as total_deletions,
		       COUNT(DISTINCT author_email) as author_count
		FROM commits
		WHERE repo_id = ?
		GROUP BY day
		ORDER BY day ASC
	`, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TimelinePoint
	for rows.Next() {
		var p TimelinePoint
		if err := rows.Scan(&p.Date, &p.CommitCount, &p.Insertions, &p.Deletions, &p.Authors); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

func (s *Store) GetOverviewStats(ctx context.Context, repoID string) (*OverviewStats, error) {
	repo, err := s.GetRepository(ctx, repoID)
	if err != nil || repo == nil {
		return nil, fmt.Errorf("repository not found: %w", err)
	}

	hotspots, err := s.GetHotspots(ctx, repoID, 10)
	if err != nil {
		return nil, err
	}

	contributors, err := s.GetContributors(ctx, repoID)
	if err != nil {
		return nil, err
	}

	recentCommits, err := s.QueryCommits(ctx, CommitFilter{RepoID: repoID, Limit: 10})
	if err != nil {
		return nil, err
	}

	timeline, err := s.GetTimeline(ctx, repoID)
	if err != nil {
		return nil, err
	}

	// Language breakdown from hotspots
	langBreakdown := make(map[string]int)
	hRows, err := s.db.QueryContext(ctx, `
		SELECT extension, COUNT(*), SUM(lines_of_code)
		FROM file_hotspots
		WHERE repo_id = ?
		GROUP BY extension
	`, repoID)
	if err == nil {
		defer hRows.Close()
		for hRows.Next() {
			var ext string
			var count, loc int
			if err := hRows.Scan(&ext, &count, &loc); err == nil {
				if ext == "" {
					ext = "other"
				}
				langBreakdown[ext] = count
			}
		}
	}

	// Total dependencies count
	var totalDeps int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dependencies WHERE repo_id = ?`, repoID).Scan(&totalDeps)

	return &OverviewStats{
		Repository:        *repo,
		TopHotspots:       hotspots,
		TopContributors:   contributors,
		RecentCommits:     recentCommits,
		LanguageBreakdown: langBreakdown,
		Timeline:          timeline,
		TotalDependencies: totalDeps,
	}, nil
}
