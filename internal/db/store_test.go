package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_archeologist.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// 1. Upsert Repository
	repo := &Repository{
		ID:             "repo-1",
		Name:           "commit-archaeologist",
		Path:           "C:/repos/test",
		DefaultBranch:  "main",
		LastAnalyzedAt: &now,
		TotalCommits:   5,
		TotalFiles:     12,
		TotalAuthors:   2,
		IsDemo:         false,
	}

	if err := store.UpsertRepository(ctx, repo); err != nil {
		t.Fatalf("failed to upsert repo: %v", err)
	}

	fetched, err := store.GetRepository(ctx, "repo-1")
	if err != nil || fetched == nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	if fetched.Name != "commit-archaeologist" {
		t.Errorf("expected repo name 'commit-archaeologist', got '%s'", fetched.Name)
	}

	// 2. Save and Query Commits
	commits := []Commit{
		{
			Hash:              "commit-abc1",
			RepoID:            "repo-1",
			ShortHash:         "abc1",
			AuthorName:        "Ada Lovelace",
			AuthorEmail:       "ada@example.com",
			AuthorDate:        now.Add(-2 * time.Hour),
			CommitterDate:     now.Add(-2 * time.Hour),
			Subject:           "Initial architecture scaffolding",
			Body:              "Setup project files",
			Branch:            "main",
			Insertions:        100,
			Deletions:         5,
			FilesChangedCount: 2,
			Files: []CommitFile{
				{
					FilePath:   "main.go",
					ChangeType: "A",
					Insertions: 80,
					Deletions:  0,
				},
				{
					FilePath:   "README.md",
					ChangeType: "A",
					Insertions: 20,
					Deletions:  5,
				},
			},
		},
		{
			Hash:              "commit-abc2",
			RepoID:            "repo-1",
			ShortHash:         "abc2",
			ParentHashes:      "commit-abc1",
			AuthorName:        "Charles Babbage",
			AuthorEmail:       "charles@example.com",
			AuthorDate:        now.Add(-1 * time.Hour),
			CommitterDate:     now.Add(-1 * time.Hour),
			Subject:           "Refactor AST engine",
			Branch:            "feature/ast",
			Insertions:        45,
			Deletions:         12,
			FilesChangedCount: 1,
		},
	}

	if err := store.SaveCommits(ctx, commits); err != nil {
		t.Fatalf("failed to save commits: %v", err)
	}

	// Query with filter
	res, err := store.QueryCommits(ctx, CommitFilter{RepoID: "repo-1", Search: "architecture"})
	if err != nil {
		t.Fatalf("failed to query commits: %v", err)
	}
	if len(res) != 1 || res[0].ShortHash != "abc1" {
		t.Fatalf("expected 1 search result with abc1, got %d", len(res))
	}

	// Commit details with files
	details, err := store.GetCommitDetails(ctx, "commit-abc1")
	if err != nil || details == nil {
		t.Fatalf("failed to get commit details: %v", err)
	}
	if len(details.Files) != 2 {
		t.Errorf("expected 2 commit files, got %d", len(details.Files))
	}

	// 3. Hotspots
	hotspots := []FileHotspot{
		{
			RepoID:          "repo-1",
			FilePath:        "main.go",
			Extension:       ".go",
			CommitCount:     15,
			AuthorCount:     2,
			TotalInsertions: 120,
			TotalDeletions:  10,
			LinesOfCode:     110,
			LastModified:    now,
			ChurnScore:      130.0,
		},
	}
	if err := store.SaveHotspots(ctx, hotspots); err != nil {
		t.Fatalf("failed to save hotspots: %v", err)
	}
	hsList, err := store.GetHotspots(ctx, "repo-1", 10)
	if err != nil || len(hsList) != 1 {
		t.Fatalf("expected 1 hotspot, got %d (err: %v)", len(hsList), err)
	}

	// 4. Dependencies
	deps := []DependencyEdge{
		{
			RepoID:          "repo-1",
			SourceFile:      "cmd/server/main.go",
			TargetFile:      "internal/db/store.go",
			ImportStatement: "commit-archaeologist/internal/db",
			Language:        "go",
			LineNumber:      5,
		},
	}
	if err := store.SaveDependencies(ctx, deps); err != nil {
		t.Fatalf("failed to save dependencies: %v", err)
	}
	graph, err := store.GetDependencyGraph(ctx, "repo-1")
	if err != nil || len(graph.Nodes) != 2 || len(graph.Links) != 1 {
		t.Fatalf("expected 2 nodes and 1 link in dependency graph, got %d nodes, %d links", len(graph.Nodes), len(graph.Links))
	}

	// 5. Contributors and Activity
	contribs := []Contributor{
		{
			RepoID:          "repo-1",
			Name:            "Ada Lovelace",
			Email:           "ada@example.com",
			TotalCommits:    10,
			FirstCommitDate: now.Add(-24 * time.Hour),
			LastCommitDate:  now,
			Insertions:      500,
			Deletions:       50,
		},
	}
	acts := []ContributorActivity{
		{
			RepoID:      "repo-1",
			Email:       "ada@example.com",
			DayOfWeek:   1,
			HourOfDay:   14,
			CommitCount: 5,
		},
	}
	if err := store.SaveContributors(ctx, contribs, acts); err != nil {
		t.Fatalf("failed to save contributors: %v", err)
	}

	activityList, err := store.GetContributorActivity(ctx, "repo-1", "")
	if err != nil || len(activityList) != 1 {
		t.Fatalf("expected 1 activity item, got %d", len(activityList))
	}

	// 6. Overview stats
	overview, err := store.GetOverviewStats(ctx, "repo-1")
	if err != nil || overview == nil {
		t.Fatalf("failed to get overview stats: %v", err)
	}
	if overview.Repository.Name != "commit-archaeologist" {
		t.Errorf("expected repo name 'commit-archaeologist', got '%s'", overview.Repository.Name)
	}
	if overview.TotalDependencies != 1 {
		t.Errorf("expected 1 dependency, got %d", overview.TotalDependencies)
	}
}
