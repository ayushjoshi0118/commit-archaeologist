package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"commit-archaeologist/internal/db"
)

// ParseLog retrieves all commits with their file statistics from the target repository.
func (e *Executor) ParseLog(ctx context.Context, repoPath string, maxCommits int) ([]db.Commit, error) {
	args := []string{
		"log",
		"--root",
		"--branches",
		"--tags",
		"--remotes",
		// 0x1e marks start of commit record, 0x00 separates metadata fields, 0x1f marks start of numstat
		"--pretty=format:%x1e%H%x00%h%x00%P%x00%an%x00%ae%x00%aI%x00%cI%x00%s%x00%b%x1f",
		"--numstat",
		"--summary",
	}

	if maxCommits > 0 {
		args = append(args, "-n", strconv.Itoa(maxCommits))
	}

	out, err := e.Run(ctx, repoPath, 60*time.Second, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch git log: %w", err)
	}

	raw := string(out)
	if strings.TrimSpace(raw) == "" {
		return []db.Commit{}, nil
	}

	rawRecords := strings.Split(raw, "\x1e")
	var commits []db.Commit

	for _, rec := range rawRecords {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}

		// Split metadata from numstats using 0x1f
		parts := strings.Split(rec, "\x1f")
		metaPart := parts[0]
		numstatPart := ""
		if len(parts) > 1 {
			numstatPart = parts[1]
		}

		metaFields := strings.Split(metaPart, "\x00")
		if len(metaFields) < 8 {
			continue
		}

		hash := strings.TrimSpace(metaFields[0])
		shortHash := strings.TrimSpace(metaFields[1])
		parents := strings.TrimSpace(metaFields[2])
		authorName := strings.TrimSpace(metaFields[3])
		authorEmail := strings.TrimSpace(metaFields[4])
		authorDateStr := strings.TrimSpace(metaFields[5])
		committerDateStr := strings.TrimSpace(metaFields[6])
		subject := strings.TrimSpace(metaFields[7])

		body := ""
		if len(metaFields) >= 9 {
			body = strings.TrimSpace(metaFields[8])
		}

		authorDate, _ := time.Parse(time.RFC3339, authorDateStr)
		committerDate, _ := time.Parse(time.RFC3339, committerDateStr)

		commitFiles, insertions, deletions := parseNumstat(numstatPart)

		commits = append(commits, db.Commit{
			Hash:              hash,
			ShortHash:         shortHash,
			ParentHashes:      parents,
			AuthorName:        authorName,
			AuthorEmail:       authorEmail,
			AuthorDate:        authorDate,
			CommitterDate:     committerDate,
			Subject:           subject,
			Body:              body,
			Branch:            "main",
			Insertions:        insertions,
			Deletions:         deletions,
			FilesChangedCount: len(commitFiles),
			Files:             commitFiles,
		})
	}

	return commits, nil
}

// parseNumstat parses the numstat lines and summaries for a commit.
func parseNumstat(raw string) ([]db.CommitFile, int, int) {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	var files []db.CommitFile
	var totalIns, totalDel int

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Check if this line is a numstat line: <ins>\t<del>\t<path>
		fields := strings.Split(line, "\t")
		if len(fields) >= 3 && isNumericOrDash(fields[0]) && isNumericOrDash(fields[1]) {
			ins, _ := strconv.Atoi(fields[0])
			del, _ := strconv.Atoi(fields[1])
			rawPath := fields[2]

			changeType := "M"
			oldPath := ""

			// Handle rename syntax: {old => new} or old => new
			filePath := cleanRenamePath(rawPath)
			if strings.Contains(rawPath, " => ") {
				changeType = "R"
				oldPath = extractOldPath(rawPath)
			}

			totalIns += ins
			totalDel += del

			files = append(files, db.CommitFile{
				FilePath:   filePath,
				ChangeType: changeType,
				OldPath:    oldPath,
				Insertions: ins,
				Deletions:  del,
			})
		} else if strings.HasPrefix(line, "create mode") {
			if len(files) > 0 {
				files[len(files)-1].ChangeType = "A"
			}
		} else if strings.HasPrefix(line, "delete mode") {
			if len(files) > 0 {
				files[len(files)-1].ChangeType = "D"
			}
		}
	}

	return files, totalIns, totalDel
}

func isNumericOrDash(s string) bool {
	if s == "-" {
		return true
	}
	_, err := strconv.Atoi(s)
	return err == nil
}

func cleanRenamePath(p string) string {
	// E.g. "pkg/{old => new}/file.go" -> "pkg/new/file.go"
	// E.g. "old => new" -> "new"
	if !strings.Contains(p, " => ") {
		return filepath.ToSlash(p)
	}

	start := strings.Index(p, "{")
	end := strings.Index(p, "}")
	if start != -1 && end != -1 && end > start {
		inner := p[start+1 : end]
		parts := strings.Split(inner, " => ")
		newPart := ""
		if len(parts) > 1 {
			newPart = parts[1]
		}
		cleaned := p[:start] + newPart + p[end+1:]
		cleaned = strings.ReplaceAll(cleaned, "//", "/")
		return filepath.ToSlash(cleaned)
	}

	parts := strings.Split(p, " => ")
	if len(parts) > 1 {
		return filepath.ToSlash(strings.TrimSpace(parts[1]))
	}
	return filepath.ToSlash(p)
}

func extractOldPath(p string) string {
	if !strings.Contains(p, " => ") {
		return ""
	}
	start := strings.Index(p, "{")
	end := strings.Index(p, "}")
	if start != -1 && end != -1 && end > start {
		inner := p[start+1 : end]
		parts := strings.Split(inner, " => ")
		cleaned := p[:start] + parts[0] + p[end+1:]
		cleaned = strings.ReplaceAll(cleaned, "//", "/")
		return filepath.ToSlash(cleaned)
	}
	parts := strings.Split(p, " => ")
	return filepath.ToSlash(strings.TrimSpace(parts[0]))
}

// ListTrackedFiles returns list of all currently tracked files in the repository.
func (e *Executor) ListTrackedFiles(ctx context.Context, repoPath string) ([]string, error) {
	out, err := e.Run(ctx, repoPath, 15*time.Second, "ls-files")
	if err != nil {
		return nil, err
	}

	var files []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		f := strings.TrimSpace(scanner.Text())
		if f != "" {
			files = append(files, filepath.ToSlash(f))
		}
	}
	return files, nil
}
