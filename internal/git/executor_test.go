package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestValidateRepoPath(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Not a repo yet
	if _, err := ValidateRepoPath(tempDir); err == nil {
		t.Errorf("expected error for non-git directory, got nil")
	}

	// 2. Init git repo
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create fake .git dir: %v", err)
	}

	validated, err := ValidateRepoPath(tempDir)
	if err != nil {
		t.Fatalf("expected valid repo, got error: %v", err)
	}
	if validated != tempDir {
		t.Errorf("expected '%s', got '%s'", tempDir, validated)
	}
}

func TestParseUnifiedDiff(t *testing.T) {
	sampleDiff := `diff --git a/src/index.ts b/src/index.ts
index 1111111..2222222 100644
--- a/src/index.ts
+++ b/src/index.ts
@@ -1,3 +1,4 @@
 import { logger } from './utils';
+import { parse } from './parser';
 
 export function run() {
`

	diffs := parseUnifiedDiff(sampleDiff)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 file diff, got %d", len(diffs))
	}

	fd := diffs[0]
	if fd.FilePath != "src/index.ts" {
		t.Errorf("expected file path 'src/index.ts', got '%s'", fd.FilePath)
	}
	if len(fd.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(fd.Hunks))
	}

	hunk := fd.Hunks[0]
	hasAdd := false
	for _, line := range hunk.Lines {
		if line.Type == "add" && line.Content == "import { parse } from './parser';" {
			hasAdd = true
			break
		}
	}
	if !hasAdd {
		t.Errorf("expected to find added line in hunk")
	}
}

func TestCleanRenamePath(t *testing.T) {
	tests := []struct {
		input       string
		expectedNew string
		expectedOld string
	}{
		{"pkg/{v1 => v2}/api.go", "pkg/v2/api.go", "pkg/v1/api.go"},
		{"old.go => new.go", "new.go", "old.go"},
		{"src/utils.ts", "src/utils.ts", ""},
	}

	for _, tc := range tests {
		newPath := cleanRenamePath(tc.input)
		oldPath := extractOldPath(tc.input)
		if newPath != tc.expectedNew {
			t.Errorf("cleanRenamePath(%s) = %s, expected %s", tc.input, newPath, tc.expectedNew)
		}
		if oldPath != tc.expectedOld {
			t.Errorf("extractOldPath(%s) = %s, expected %s", tc.input, oldPath, tc.expectedOld)
		}
	}
}

func TestRealGitExecution(t *testing.T) {
	execBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not installed, skipping test")
	}

	tempDir := t.TempDir()
	e := &Executor{gitBin: execBin}
	ctx := context.Background()

	// Initialize git repo
	if _, err := e.Run(ctx, tempDir, 0, "init"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	// Configure user for commits in test
	_, _ = e.Run(ctx, tempDir, 0, "config", "user.name", "Test User")
	_, _ = e.Run(ctx, tempDir, 0, "config", "user.email", "test@example.com")

	// Create test file
	testFile := filepath.Join(tempDir, "hello.txt")
	if err := os.WriteFile(testFile, []byte("Hello Commit Archaeologist\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Add & commit
	if _, err := e.Run(ctx, tempDir, 0, "add", "hello.txt"); err != nil {
		t.Fatalf("git add failed: %v", err)
	}
	if _, err := e.Run(ctx, tempDir, 0, "commit", "-m", "Initial commit"); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	// Test ParseLog
	commits, err := e.ParseLog(ctx, tempDir, 10)
	if err != nil {
		t.Fatalf("ParseLog failed: %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}
	if commits[0].Subject != "Initial commit" {
		t.Errorf("expected subject 'Initial commit', got '%s'", commits[0].Subject)
	}
	if commits[0].AuthorName != "Test User" {
		t.Errorf("expected author 'Test User', got '%s'", commits[0].AuthorName)
	}
	if len(commits[0].Files) != 1 || commits[0].Files[0].FilePath != "hello.txt" {
		t.Errorf("expected 1 commit file hello.txt, got %+v", commits[0].Files)
	}

	// Test diff
	diffRes, err := e.GetCommitDiff(ctx, tempDir, commits[0].Hash, "hello.txt")
	if err != nil {
		t.Fatalf("GetCommitDiff failed: %v", err)
	}
	if len(diffRes.Files) != 1 {
		t.Fatalf("expected 1 diff file, got %d", len(diffRes.Files))
	}
}
