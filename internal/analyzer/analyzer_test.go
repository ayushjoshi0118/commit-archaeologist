package analyzer

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"commit-archaeologist/internal/db"
	"commit-archaeologist/internal/demo"
	"commit-archaeologist/internal/git"
)

func TestAnalyzeRepositoryEndToEnd(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed, skipping end-to-end analyzer test")
	}

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	store, err := db.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	gitExec, err := git.NewExecutor()
	if err != nil {
		t.Fatalf("failed to create git exec: %v", err)
	}

	ctx := context.Background()
	repoDir, err := demo.EnsureSyntheticRepo(ctx, tempDir, gitExec)
	if err != nil {
		t.Fatalf("failed to create demo repo: %v", err)
	}

	a := NewAnalyzer(store, gitExec)
	var stagesSeen []string
	progressCb := func(stage string, current, total int, msg string) {
		stagesSeen = append(stagesSeen, stage)
	}

	overview, err := a.AnalyzeRepository(ctx, repoDir, true, progressCb)
	if err != nil {
		t.Fatalf("AnalyzeRepository failed: %v", err)
	}

	if overview == nil {
		t.Fatalf("overview is nil")
	}
	if overview.Repository.TotalCommits < 6 {
		t.Errorf("expected at least 6 commits, got %d", overview.Repository.TotalCommits)
	}
	if len(overview.TopContributors) < 3 {
		t.Errorf("expected at least 3 contributors, got %d", len(overview.TopContributors))
	}
	if len(overview.TopHotspots) == 0 {
		t.Errorf("expected non-empty hotspots")
	}
	if len(stagesSeen) != 5 {
		t.Errorf("expected 5 progress stages, got %d", len(stagesSeen))
	}

	// Verify persistence in SQLite
	fetchedRepo, err := store.GetRepository(ctx, overview.Repository.ID)
	if err != nil || fetchedRepo == nil {
		t.Fatalf("failed to fetch saved repository: %v", err)
	}
	if fetchedRepo.TotalCommits != overview.Repository.TotalCommits {
		t.Errorf("mismatched saved commits: %d vs %d", fetchedRepo.TotalCommits, overview.Repository.TotalCommits)
	}

	// Verify dependency graph query
	graph, err := store.GetDependencyGraph(ctx, overview.Repository.ID)
	if err != nil {
		t.Fatalf("failed to get dependency graph: %v", err)
	}
	t.Logf("Found %d nodes and %d links in parsed dependency graph", len(graph.Nodes), len(graph.Links))
}
