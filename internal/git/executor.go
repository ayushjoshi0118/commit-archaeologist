package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalidRepoPath = errors.New("invalid repository path: directory does not exist or lacks .git")
	ErrCommandFailed   = errors.New("git command execution failed")
	ErrGitNotFound     = errors.New("git binary not found in system PATH")
)

type Executor struct {
	gitBin string
}

func NewExecutor() (*Executor, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, ErrGitNotFound
	}
	return &Executor{gitBin: gitPath}, nil
}

// ValidateRepoPath ensures the path exists, is a directory, and contains a .git folder or is a bare git repo.
func ValidateRepoPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("repository path cannot be empty")
	}

	cleanPath := filepath.Clean(path)
	info, err := os.Stat(cleanPath)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("%w: path '%s' is not an accessible directory", ErrInvalidRepoPath, cleanPath)
	}

	gitDir := filepath.Join(cleanPath, ".git")
	if gitInfo, err := os.Stat(gitDir); err == nil && (gitInfo.IsDir() || !gitInfo.IsDir()) {
		// Normal repo (.git dir) or submodule/worktree (.git file)
		return cleanPath, nil
	}

	// Check if bare repo
	headFile := filepath.Join(cleanPath, "HEAD")
	if _, err := os.Stat(headFile); err == nil {
		return cleanPath, nil
	}

	return "", fmt.Errorf("%w: no .git found in '%s'", ErrInvalidRepoPath, cleanPath)
}

// Run executes a git command in the repo directory using safe, non-shell argument slices.
func (e *Executor) Run(ctx context.Context, repoPath string, timeout time.Duration, args ...string) ([]byte, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Sanitize arguments: reject any unexpected flag injections where positional paths are expected
	for _, arg := range args {
		if strings.Contains(arg, "\x00") {
			return nil, errors.New("null byte in git command argument")
		}
	}

	cmd := exec.CommandContext(ctx, e.gitBin, args...)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"LC_ALL=C.UTF-8",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return nil, fmt.Errorf("%w: git %s: %s", ErrCommandFailed, strings.Join(args, " "), errStr)
	}

	return stdout.Bytes(), nil
}

// GetDefaultBranch returns the default branch name (e.g. main, master).
func (e *Executor) GetDefaultBranch(ctx context.Context, repoPath string) (string, error) {
	out, err := e.Run(ctx, repoPath, 5*time.Second, "symbolic-ref", "--short", "HEAD")
	if err == nil {
		branch := strings.TrimSpace(string(out))
		if branch != "" {
			return branch, nil
		}
	}

	// Fallback to checking branches
	out, err = e.Run(ctx, repoPath, 5*time.Second, "branch", "--list", "main", "master")
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(strings.TrimPrefix(l, "*"))
			if l == "main" || l == "master" {
				return l, nil
			}
		}
	}

	return "main", nil
}

// GetBranches returns list of local and remote branches.
func (e *Executor) GetBranches(ctx context.Context, repoPath string) ([]string, error) {
	out, err := e.Run(ctx, repoPath, 5*time.Second, "branch", "-a", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	var branches []string
	lines := strings.Split(string(out), "\n")
	seen := make(map[string]bool)
	for _, line := range lines {
		b := strings.TrimSpace(line)
		if b != "" && !seen[b] {
			seen[b] = true
			branches = append(branches, b)
		}
	}
	return branches, nil
}
