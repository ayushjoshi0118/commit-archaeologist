package demo

import (
	"context"
	"os/exec"
	"testing"

	"commit-archaeologist/internal/git"
)

func TestEnsureSyntheticRepo(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not installed, skipping demo test")
	}

	tempDir := t.TempDir()
	gitExec, err := git.NewExecutor()
	if err != nil {
		t.Fatalf("failed to create git executor: %v", err)
	}

	ctx := context.Background()
	repoDir, err := EnsureSyntheticRepo(ctx, tempDir, gitExec)
	if err != nil {
		t.Fatalf("failed to create synthetic repo: %v", err)
	}

	validated, err := git.ValidateRepoPath(repoDir)
	if err != nil || validated == "" {
		t.Fatalf("synthetic repo path invalid: %v", err)
	}

	commits, err := gitExec.ParseLog(ctx, repoDir, 50)
	if err != nil {
		t.Fatalf("failed to parse commits from synthetic repo: %v", err)
	}

	if len(commits) < 6 {
		t.Errorf("expected at least 6 commits in synthetic repo, got %d", len(commits))
	}

	// Verify idempotency (calling again returns existing repo)
	again, err := EnsureSyntheticRepo(ctx, tempDir, gitExec)
	if err != nil || again != repoDir {
		t.Errorf("expected idempotent repoDir return, got %s (err: %v)", again, err)
	}
	_ = gitBin
}
