package ast

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	jsImportRegex  = regexp.MustCompile(`(?m)^\s*(?:import\s+(?:[\w*\s{},]*\s+from\s+)?|export\s+(?:[\w*\s{},]*\s+from\s+)?)['"]([^'"]+)['"]`)
	jsRequireRegex = regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	jsDynamicRegex = regexp.MustCompile(`import\s*\(\s*['"]([^'"]+)['"]\s*\)`)
)

func parseJSImports(relPath string, data []byte, fileSet map[string]bool) []ExtractedImport {
	content := string(data)
	lines := strings.Split(content, "\n")
	var results []ExtractedImport
	seen := make(map[string]bool)

	for idx, line := range lines {
		lineNum := idx + 1
		trimmed := strings.TrimSpace(line)

		// Check ES6 import/export
		matches := jsImportRegex.FindStringSubmatch(trimmed)
		if len(matches) > 1 {
			importPath := matches[1]
			addJSImport(relPath, importPath, lineNum, trimmed, fileSet, seen, &results)
			continue
		}

		// Check CommonJS require
		rMatches := jsRequireRegex.FindAllStringSubmatch(trimmed, -1)
		for _, m := range rMatches {
			if len(m) > 1 {
				addJSImport(relPath, m[1], lineNum, trimmed, fileSet, seen, &results)
			}
		}

		// Check Dynamic import()
		dMatches := jsDynamicRegex.FindAllStringSubmatch(trimmed, -1)
		for _, m := range dMatches {
			if len(m) > 1 {
				addJSImport(relPath, m[1], lineNum, trimmed, fileSet, seen, &results)
			}
		}
	}

	return results
}

func addJSImport(sourceFile, importPath string, lineNum int, raw string, fileSet map[string]bool, seen map[string]bool, results *[]ExtractedImport) {
	if seen[importPath] {
		return
	}
	seen[importPath] = true

	target := resolveJSImportTarget(sourceFile, importPath, fileSet)
	lang := DetectLanguage(sourceFile)

	*results = append(*results, ExtractedImport{
		ImportPath: importPath,
		TargetFile: target,
		LineNumber: lineNum,
		Language:   lang,
		Raw:        raw,
	})
}

func resolveJSImportTarget(sourceFile, importPath string, fileSet map[string]bool) string {
	// Only resolve relative or alias paths within the repo
	if !strings.HasPrefix(importPath, ".") && !strings.HasPrefix(importPath, "/") && !strings.HasPrefix(importPath, "$") && !strings.HasPrefix(importPath, "@") {
		return ""
	}

	sourceDir := path.Dir(filepath.ToSlash(sourceFile))
	baseTarget := path.Clean(path.Join(sourceDir, importPath))

	// Handle alias prefix like $lib or @/
	if strings.HasPrefix(importPath, "$lib/") {
		baseTarget = "src/lib/" + strings.TrimPrefix(importPath, "$lib/")
	} else if strings.HasPrefix(importPath, "@/") {
		baseTarget = "src/" + strings.TrimPrefix(importPath, "@/")
	}

	// Try exact match
	if fileSet[baseTarget] {
		return baseTarget
	}

	// Try extensions
	extensions := []string{".ts", ".tsx", ".js", ".jsx", ".svelte", ".mjs", ".json"}
	for _, ext := range extensions {
		candidate := baseTarget + ext
		if fileSet[candidate] {
			return candidate
		}
	}

	// Try index files in directory
	for _, ext := range extensions {
		candidate := path.Join(baseTarget, "index"+ext)
		if fileSet[candidate] {
			return candidate
		}
	}

	return ""
}
