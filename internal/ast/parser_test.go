package ast

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseGoImportsAndSymbols(t *testing.T) {
	goCode := []byte(`package main

import (
	"fmt"
	"commit-archaeologist/internal/db"
)

type Worker struct {
	ID int
}

func ProcessTask(w *Worker) error {
	fmt.Println(w.ID)
	return nil
}
`)

	fileSet := map[string]bool{
		"main.go":              true,
		"internal/db/store.go": true,
	}

	imports := parseGoImports("main.go", goCode, fileSet)
	if len(imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(imports))
	}

	var foundLocal bool
	for _, imp := range imports {
		if imp.ImportPath == "commit-archaeologist/internal/db" {
			if imp.TargetFile != "internal/db/store.go" {
				t.Errorf("expected target 'internal/db/store.go', got '%s'", imp.TargetFile)
			}
			foundLocal = true
		}
	}
	if !foundLocal {
		t.Errorf("failed to resolve internal Go import target")
	}

	symbols := ParseGoSymbols("main.go", goCode)
	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}
	if symbols[0].Name != "Worker" || symbols[0].Kind != "struct" {
		t.Errorf("expected struct Worker, got %+v", symbols[0])
	}
	if symbols[1].Name != "ProcessTask" || symbols[1].Kind != "function" {
		t.Errorf("expected func ProcessTask, got %+v", symbols[1])
	}
}

func TestParseJSImports(t *testing.T) {
	tsCode := []byte(`import { Header } from './components/Header';
import { fetchCommits } from '../api/client';
import type { Commit } from '$lib/types';
import axios from 'axios';
`)

	fileSet := map[string]bool{
		"src/views/Overview.svelte":           true,
		"src/views/components/Header.svelte":  true,
		"src/api/client.ts":                   true,
		"src/lib/types.ts":                    true,
	}

	imports := parseJSImports("src/views/Overview.svelte", tsCode, fileSet)
	if len(imports) < 3 {
		t.Fatalf("expected at least 3 imports, got %d", len(imports))
	}

	resolvedTargets := make(map[string]string)
	for _, imp := range imports {
		resolvedTargets[imp.ImportPath] = imp.TargetFile
	}

	if resolvedTargets["./components/Header"] != "src/views/components/Header.svelte" {
		t.Errorf("expected './components/Header' to resolve to 'src/views/components/Header.svelte', got '%s'", resolvedTargets["./components/Header"])
	}
	if resolvedTargets["../api/client"] != "src/api/client.ts" {
		t.Errorf("expected '../api/client' to resolve to 'src/api/client.ts', got '%s'", resolvedTargets["../api/client"])
	}
	if resolvedTargets["$lib/types"] != "src/lib/types.ts" {
		t.Errorf("expected '$lib/types' to resolve to 'src/lib/types.ts', got '%s'", resolvedTargets["$lib/types"])
	}
	if resolvedTargets["axios"] != "" {
		t.Errorf("external package 'axios' should not resolve to internal file, got '%s'", resolvedTargets["axios"])
	}
}

func TestParsePythonImports(t *testing.T) {
	pyCode := []byte(`import os
from .config import settings
from services.analytics import calculate_hotspots
`)

	fileSet := map[string]bool{
		"app/main.py":               true,
		"app/config.py":             true,
		"services/analytics.py":     true,
	}

	imports := parsePythonImports("app/main.py", pyCode, fileSet)
	if len(imports) != 3 {
		t.Fatalf("expected 3 imports, got %d", len(imports))
	}

	resolvedTargets := make(map[string]string)
	for _, imp := range imports {
		resolvedTargets[imp.ImportPath] = imp.TargetFile
	}

	if resolvedTargets[".config"] != "app/config.py" {
		t.Errorf("expected '.config' to resolve to 'app/config.py', got '%s'", resolvedTargets[".config"])
	}
	if resolvedTargets["services.analytics"] != "services/analytics.py" {
		t.Errorf("expected 'services.analytics' to resolve to 'services/analytics.py', got '%s'", resolvedTargets["services.analytics"])
	}
}

func TestEngineParseRepoFiles(t *testing.T) {
	tempDir := t.TempDir()

	f1 := filepath.Join(tempDir, "a.ts")
	f2 := filepath.Join(tempDir, "b.ts")

	_ = os.WriteFile(f1, []byte("import { b } from './b';\n"), 0644)
	_ = os.WriteFile(f2, []byte("export const b = 42;\n"), 0644)

	engine := NewEngine()
	edges, locMap, err := engine.ParseRepoFiles(context.Background(), "test-repo", tempDir, []string{"a.ts", "b.ts"})
	if err != nil {
		t.Fatalf("ParseRepoFiles failed: %v", err)
	}

	if len(edges) != 1 {
		t.Fatalf("expected 1 dependency edge, got %d", len(edges))
	}
	if edges[0].SourceFile != "a.ts" || edges[0].TargetFile != "b.ts" {
		t.Errorf("expected a.ts -> b.ts, got %s -> %s", edges[0].SourceFile, edges[0].TargetFile)
	}
	if locMap["a.ts"] != 1 || locMap["b.ts"] != 1 {
		t.Errorf("expected 1 line of code each, got %+v", locMap)
	}
}
