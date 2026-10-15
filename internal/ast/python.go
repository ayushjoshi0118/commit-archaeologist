package ast

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	pyImportRegex     = regexp.MustCompile(`^\s*import\s+([a-zA-Z0-9_\.]+)`)
	pyFromImportRegex = regexp.MustCompile(`^\s*from\s+([\.a-zA-Z0-9_]+)\s+import`)
)

func parsePythonImports(relPath string, data []byte, fileSet map[string]bool) []ExtractedImport {
	lines := strings.Split(string(data), "\n")
	var results []ExtractedImport
	seen := make(map[string]bool)

	for idx, line := range lines {
		lineNum := idx + 1
		trimmed := strings.TrimSpace(line)

		// E.g. "from .models import User" or "from utils import helper"
		if matches := pyFromImportRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			mod := matches[1]
			addPyImport(relPath, mod, lineNum, trimmed, fileSet, seen, &results)
			continue
		}

		// E.g. "import config"
		if matches := pyImportRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			mod := matches[1]
			addPyImport(relPath, mod, lineNum, trimmed, fileSet, seen, &results)
		}
	}

	return results
}

func addPyImport(sourceFile, importMod string, lineNum int, raw string, fileSet map[string]bool, seen map[string]bool, results *[]ExtractedImport) {
	if seen[importMod] {
		return
	}
	seen[importMod] = true

	target := resolvePythonImportTarget(sourceFile, importMod, fileSet)

	*results = append(*results, ExtractedImport{
		ImportPath: importMod,
		TargetFile: target,
		LineNumber: lineNum,
		Language:   LangPython,
		Raw:        raw,
	})
}

func resolvePythonImportTarget(sourceFile, importMod string, fileSet map[string]bool) string {
	sourceDir := path.Dir(filepath.ToSlash(sourceFile))

	// Handle relative imports: . or ..
	if strings.HasPrefix(importMod, ".") {
		dotCount := 0
		for _, ch := range importMod {
			if ch == '.' {
				dotCount++
			} else {
				break
			}
		}
		modName := importMod[dotCount:]
		currDir := sourceDir
		for i := 1; i < dotCount; i++ {
			currDir = path.Dir(currDir)
		}

		subPath := strings.ReplaceAll(modName, ".", "/")
		candidate1 := path.Join(currDir, subPath+".py")
		if fileSet[candidate1] {
			return candidate1
		}
		candidate2 := path.Join(currDir, subPath, "__init__.py")
		if fileSet[candidate2] {
			return candidate2
		}
		return ""
	}

	// Module path: convert dots to slashes: "pkg.core" -> "pkg/core.py"
	modPath := strings.ReplaceAll(importMod, ".", "/")
	candidates := []string{
		modPath + ".py",
		path.Join(modPath, "__init__.py"),
		path.Join(sourceDir, modPath+".py"),
		path.Join(sourceDir, modPath, "__init__.py"),
	}

	for _, c := range candidates {
		c = path.Clean(c)
		if fileSet[c] {
			return c
		}
	}

	return ""
}
