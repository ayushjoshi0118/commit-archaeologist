package git

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	validHashRegex = regexp.MustCompile(`^[a-fA-F0-9]{4,64}$`)
)

type DiffLine struct {
	Type    string `json:"type"` // "add", "del", "context", "header"
	Content string `json:"content"`
	OldLine int    `json:"oldLine,omitempty"`
	NewLine int    `json:"newLine,omitempty"`
}

type DiffHunk struct {
	Header string     `json:"header"`
	Lines  []DiffLine `json:"lines"`
}

type FileDiff struct {
	FilePath   string     `json:"filePath"`
	OldPath    string     `json:"oldPath,omitempty"`
	ChangeType string     `json:"changeType"` // "modified", "added", "deleted", "renamed"
	Hunks      []DiffHunk `json:"hunks"`
	Raw        string     `json:"raw"`
}

type CommitDiffResponse struct {
	CommitHash string     `json:"commitHash"`
	Files      []FileDiff `json:"files"`
}

// GetCommitDiff fetches the diff for a commit, optionally restricted to a specific file.
func (e *Executor) GetCommitDiff(ctx context.Context, repoPath, commitHash, targetFile string) (*CommitDiffResponse, error) {
	if !validHashRegex.MatchString(commitHash) {
		return nil, errors.New("invalid commit hash")
	}

	args := []string{"show", "--patch", "--no-color", commitHash}
	if targetFile != "" {
		cleaned := filepath.ToSlash(filepath.Clean(targetFile))
		if strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
			return nil, errors.New("invalid target file path")
		}
		args = append(args, "--", cleaned)
	}

	out, err := e.Run(ctx, repoPath, 20*time.Second, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit diff: %w", err)
	}

	fileDiffs := parseUnifiedDiff(string(out))
	return &CommitDiffResponse{
		CommitHash: commitHash,
		Files:      fileDiffs,
	}, nil
}

// GetFileContent retrieves the file content at a specific commit.
func (e *Executor) GetFileContent(ctx context.Context, repoPath, commitHash, targetFile string) (string, error) {
	if !validHashRegex.MatchString(commitHash) {
		return "", errors.New("invalid commit hash")
	}

	cleaned := filepath.ToSlash(filepath.Clean(targetFile))
	if strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", errors.New("invalid target file path")
	}

	spec := fmt.Sprintf("%s:%s", commitHash, cleaned)
	out, err := e.Run(ctx, repoPath, 10*time.Second, "show", spec)
	if err != nil {
		return "", fmt.Errorf("failed to fetch file content for %s: %w", spec, err)
	}

	return string(out), nil
}

// parseUnifiedDiff parses git show / git diff output into structured file diffs and hunks.
func parseUnifiedDiff(raw string) []FileDiff {
	var diffs []FileDiff
	scanner := bufio.NewScanner(strings.NewReader(raw))

	var currentFile *FileDiff
	var currentHunk *DiffHunk
	var rawBuffer bytes.Buffer

	oldLineNum := 0
	newLineNum := 0

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "diff --git ") {
			if currentFile != nil {
				if currentHunk != nil {
					currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
					currentHunk = nil
				}
				currentFile.Raw = rawBuffer.String()
				diffs = append(diffs, *currentFile)
				rawBuffer.Reset()
			}

			parts := strings.Split(line, " ")
			filePath := ""
			if len(parts) >= 4 {
				filePath = strings.TrimPrefix(parts[3], "b/")
			}

			currentFile = &FileDiff{
				FilePath:   filePath,
				ChangeType: "modified",
			}
			rawBuffer.WriteString(line + "\n")
			continue
		}

		if currentFile == nil {
			continue
		}

		rawBuffer.WriteString(line + "\n")

		if strings.HasPrefix(line, "new file mode") {
			currentFile.ChangeType = "added"
		} else if strings.HasPrefix(line, "deleted file mode") {
			currentFile.ChangeType = "deleted"
		} else if strings.HasPrefix(line, "similarity index") || strings.HasPrefix(line, "rename from") {
			currentFile.ChangeType = "renamed"
		} else if strings.HasPrefix(line, "@@ ") {
			if currentHunk != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}
			currentHunk = &DiffHunk{
				Header: line,
				Lines:  []DiffLine{},
			}
			// Extract line numbers from @@ -a,b +c,d @@
			hParts := strings.Split(line, "@@")
			if len(hParts) >= 2 {
				ranges := strings.TrimSpace(hParts[1])
				parseHunkRange(ranges, &oldLineNum, &newLineNum)
			}
		} else if currentHunk != nil {
			if strings.HasPrefix(line, "+") {
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Type:    "add",
					Content: line[1:],
					NewLine: newLineNum,
				})
				newLineNum++
			} else if strings.HasPrefix(line, "-") {
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Type:    "del",
					Content: line[1:],
					OldLine: oldLineNum,
				})
				oldLineNum++
			} else if strings.HasPrefix(line, " ") || line == "" {
				content := line
				if strings.HasPrefix(line, " ") {
					content = line[1:]
				}
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Type:    "context",
					Content: content,
					OldLine: oldLineNum,
					NewLine: newLineNum,
				})
				oldLineNum++
				newLineNum++
			}
		}
	}

	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		currentFile.Raw = rawBuffer.String()
		diffs = append(diffs, *currentFile)
	}

	return diffs
}

func parseHunkRange(ranges string, oldLineNum, newLineNum *int) {
	parts := strings.Split(ranges, " ")
	for _, p := range parts {
		if strings.HasPrefix(p, "-") {
			nums := strings.Split(strings.TrimPrefix(p, "-"), ",")
			if n, err := strconvAtoiSafe(nums[0]); err == nil {
				*oldLineNum = n
			}
		} else if strings.HasPrefix(p, "+") {
			nums := strings.Split(strings.TrimPrefix(p, "+"), ",")
			if n, err := strconvAtoiSafe(nums[0]); err == nil {
				*newLineNum = n
			}
		}
	}
}

func strconvAtoiSafe(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
