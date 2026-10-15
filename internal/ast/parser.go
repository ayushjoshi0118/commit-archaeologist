package ast

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"commit-archaeologist/internal/db"
)

type Language string

const (
	LangGo         Language = "go"
	LangTypeScript Language = "typescript"
	LangJavaScript Language = "javascript"
	LangPython     Language = "python"
	LangUnknown    Language = "unknown"
)

type ExtractedImport struct {
	ImportPath string
	TargetFile string
	LineNumber int
	Language   Language
	Raw        string
}

type ParsedSymbol struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"` // "function", "class", "interface", "struct", "variable"
	LineNumber int    `json:"lineNumber"`
	Exported   bool   `json:"exported"`
}

type FileASTResult struct {
	FilePath    string
	Language    Language
	LinesOfCode int
	Imports     []ExtractedImport
	Symbols     []ParsedSymbol
}

type Engine struct {
	mu sync.RWMutex
}

func NewEngine() *Engine {
	return &Engine{}
}

// DetectLanguage returns the Language enum based on file extension.
func DetectLanguage(filePath string) Language {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return LangGo
	case ".ts", ".tsx":
		return LangTypeScript
	case ".js", ".jsx", ".mjs", ".cjs":
		return LangJavaScript
	case ".py":
		return LangPython
	default:
		return LangUnknown
	}
}

// ParseRepoFiles scans all given files in repoPath, extracts imports & dependencies, and resolves cross-file links.
func (e *Engine) ParseRepoFiles(ctx context.Context, repoID, repoPath string, files []string) ([]db.DependencyEdge, map[string]int, error) {
	fileSet := make(map[string]bool)
	for _, f := range files {
		fileSet[filepath.ToSlash(f)] = true
	}

	locMap := make(map[string]int)
	var allEdges []db.DependencyEdge
	var mu sync.Mutex

	// Process files in parallel worker pool
	type task struct {
		relPath string
	}
	taskChan := make(chan task, 100)
	var wg sync.WaitGroup

	numWorkers := 8
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range taskChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				absPath := filepath.Join(repoPath, t.relPath)
				data, err := os.ReadFile(absPath)
				if err != nil {
					continue
				}

				loc := countLinesOfCode(data)
				lang := DetectLanguage(t.relPath)

				var imports []ExtractedImport
				switch lang {
				case LangGo:
					imports = parseGoImports(t.relPath, data, fileSet)
				case LangTypeScript, LangJavaScript:
					imports = parseJSImports(t.relPath, data, fileSet)
				case LangPython:
					imports = parsePythonImports(t.relPath, data, fileSet)
				}

				mu.Lock()
				locMap[t.relPath] = loc
				for _, imp := range imports {
					if imp.TargetFile != "" && fileSet[imp.TargetFile] && imp.TargetFile != t.relPath {
						allEdges = append(allEdges, db.DependencyEdge{
							RepoID:          repoID,
							SourceFile:      t.relPath,
							TargetFile:      imp.TargetFile,
							ImportStatement: imp.ImportPath,
							Language:        string(imp.Language),
							LineNumber:      imp.LineNumber,
						})
					}
				}
				mu.Unlock()
			}
		}()
	}

	for _, f := range files {
		taskChan <- task{relPath: filepath.ToSlash(f)}
	}
	close(taskChan)
	wg.Wait()

	return allEdges, locMap, nil
}

func countLinesOfCode(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	// If file does not end with newline but has bytes, count the last line
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}
