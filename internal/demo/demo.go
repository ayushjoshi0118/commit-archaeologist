package demo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"commit-archaeologist/internal/git"
)

type DemoAuthor struct {
	Name  string
	Email string
}

var sampleAuthors = []DemoAuthor{
	{Name: "Elena Rostova", Email: "elena.rostova@devarch.internal"},
	{Name: "Marcus Chen", Email: "marcus.chen@devarch.internal"},
	{Name: "Sarah Jenkins", Email: "sarah.j@devarch.internal"},
	{Name: "Alex Rivera", Email: "alex.rivera@devarch.internal"},
}

// EnsureSyntheticRepo creates or populates a realistic synthetic git repository on disk.
func EnsureSyntheticRepo(ctx context.Context, targetDir string, gitExec *git.Executor) (string, error) {
	repoDir := filepath.Join(targetDir, "sample-repo-chronolog")

	// If already exists and is valid git repo, return it
	if _, err := git.ValidateRepoPath(repoDir); err == nil {
		return repoDir, nil
	}

	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create demo repo dir: %w", err)
	}

	// Git init
	if _, err := gitExec.Run(ctx, repoDir, 10*time.Second, "init", "-b", "main"); err != nil {
		// Fallback for older git without -b flag
		if _, err := gitExec.Run(ctx, repoDir, 10*time.Second, "init"); err != nil {
			return "", fmt.Errorf("failed to git init demo repo: %w", err)
		}
	}

	now := time.Now().UTC()

	// Commit 1: Initial commit by Elena
	c1Time := now.Add(-30 * 24 * time.Hour).Format(time.RFC3339)
	createFile(repoDir, "README.md", "# ChronoLog Engine\n\nHigh throughput distributed telemetry ingest engine.\n")
	createFile(repoDir, "go.mod", "module chronolog\n\ngo 1.22\n")
	createFile(repoDir, "cmd/server/main.go", `package main

import (
	"fmt"
	"chronolog/internal/storage"
)

func main() {
	fmt.Println("Starting ChronoLog Ingest Engine...")
	storage.InitDB()
}
`)
	createFile(repoDir, "internal/storage/sqlite.go", `package storage

import "fmt"

func InitDB() {
	fmt.Println("Initializing SQLite storage...")
}
`)
	commitAs(ctx, gitExec, repoDir, sampleAuthors[0], "chore: initial project skeleton and storage scaffold", c1Time)

	// Commit 2: TypeScript frontend ingestion SDK by Marcus
	c2Time := now.Add(-25 * 24 * time.Hour).Format(time.RFC3339)
	createFile(repoDir, "sdk/src/types.ts", `export interface TelemetryEvent {
    id: string;
    timestamp: number;
    payload: Record<string, unknown>;
}
`)
	createFile(repoDir, "sdk/src/client.ts", `import { TelemetryEvent } from './types';

export class ChronoClient {
    async send(event: TelemetryEvent): Promise<boolean> {
        console.log("Dispatching event:", event.id);
        return true;
    }
}
`)
	commitAs(ctx, gitExec, repoDir, sampleAuthors[1], "feat(sdk): create TypeScript telemetry client and type definitions", c2Time)

	// Commit 3: Python benchmarks and analysis scripts by Sarah
	c3Time := now.Add(-20 * 24 * time.Hour).Format(time.RFC3339)
	createFile(repoDir, "scripts/config.py", `BATCH_SIZE = 5000
TIMEOUT_SECONDS = 15
`)
	createFile(repoDir, "scripts/benchmark.py", `import time
from config import BATCH_SIZE

def run_benchmark():
    print(f"Executing ingest benchmark with batch size {BATCH_SIZE}")
`)
	commitAs(ctx, gitExec, repoDir, sampleAuthors[2], "feat(scripts): add performance benchmarking harness", c3Time)

	// Commit 4: Refactor storage layer by Elena (high churn hotspot!)
	c4Time := now.Add(-15 * 24 * time.Hour).Format(time.RFC3339)
	createFile(repoDir, "internal/storage/sqlite.go", `package storage

import (
	"database/sql"
	"fmt"
)

type StorageEngine struct {
	db *sql.DB
}

func InitDB() *StorageEngine {
	fmt.Println("Initializing SQLite storage with WAL mode enabled...")
	return &StorageEngine{}
}

func (s *StorageEngine) InsertBatch(records int) error {
	fmt.Printf("Writing batch of %d telemetry records\n", records)
	return nil
}
`)
	commitAs(ctx, gitExec, repoDir, sampleAuthors[0], "refactor(storage): enable WAL mode and batch ingestion buffering", c4Time)

	// Commit 5: Branch 'feature/compression' by Alex
	_, _ = gitExec.Run(ctx, repoDir, 5*time.Second, "checkout", "-b", "feature/compression")
	c5Time := now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)
	createFile(repoDir, "internal/compress/snappy.go", `package compress

import "fmt"

func CompressBlock(data []byte) []byte {
	fmt.Println("Compressing data block with Snappy...")
	return data
}
`)
	commitAs(ctx, gitExec, repoDir, sampleAuthors[3], "feat(compress): add Snappy block compression handler", c5Time)

	// Commit 6: Merge branch into main
	_, _ = gitExec.Run(ctx, repoDir, 5*time.Second, "checkout", "main")
	_, _ = gitExec.Run(ctx, repoDir, 5*time.Second, "merge", "--no-ff", "-m", "Merge branch 'feature/compression' into main", "feature/compression")

	// Commit 7: SDK retry policy and tests by Marcus
	c7Time := now.Add(-5 * 24 * time.Hour).Format(time.RFC3339)
	createFile(repoDir, "sdk/src/retry.ts", `export function exponentialBackoff(attempt: number): number {
    return Math.min(1000 * Math.pow(2, attempt), 10000);
}
`)
	createFile(repoDir, "sdk/src/client.ts", `import { TelemetryEvent } from './types';
import { exponentialBackoff } from './retry';

export class ChronoClient {
    async send(event: TelemetryEvent): Promise<boolean> {
        console.log("Dispatching event with backoff policy:", event.id, exponentialBackoff(1));
        return true;
    }
}
`)
	commitAs(ctx, gitExec, repoDir, sampleAuthors[1], "feat(sdk): add exponential backoff retry mechanism to client", c7Time)

	// Commit 8: Documentation and configuration polish by Sarah
	c8Time := now.Add(-1 * 24 * time.Hour).Format(time.RFC3339)
	createFile(repoDir, "README.md", `# ChronoLog Engine

High throughput distributed telemetry ingest engine.

## Features
- Scalable SQLite WAL storage engine
- Zero-copy Snappy block compression
- Modern TypeScript SDK with retry resilience
- Python analytics & benchmark suite
`)
	commitAs(ctx, gitExec, repoDir, sampleAuthors[2], "docs: update architecture documentation and client usage guide", c8Time)

	return repoDir, nil
}

func createFile(baseDir, relPath, content string) {
	fullPath := filepath.Join(baseDir, relPath)
	_ = os.MkdirAll(filepath.Dir(fullPath), 0755)
	_ = os.WriteFile(fullPath, []byte(content), 0644)
}

func commitAs(ctx context.Context, gitExec *git.Executor, repoDir string, author DemoAuthor, message, dateStr string) {
	_, _ = gitExec.Run(ctx, repoDir, 5*time.Second, "add", "-A")

	cmd := []string{
		"-c", fmt.Sprintf("user.name=%s", author.Name),
		"-c", fmt.Sprintf("user.email=%s", author.Email),
		"commit",
		"--date", dateStr,
		"-m", message,
	}
	_, _ = gitExec.Run(ctx, repoDir, 5*time.Second, cmd...)
}
