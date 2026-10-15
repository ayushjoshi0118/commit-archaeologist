package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

func parseGoImports(relPath string, data []byte, fileSet map[string]bool) []ExtractedImport {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, relPath, data, parser.ImportsOnly)
	if err != nil {
		// Fallback to line scanning if partial syntax error
		return fallbackScanGoImports(relPath, string(data), fileSet)
	}

	var results []ExtractedImport
	for _, imp := range node.Imports {
		if imp.Path == nil {
			continue
		}
		rawPath := strings.Trim(imp.Path.Value, `"`)
		line := fset.Position(imp.Pos()).Line

		target := resolveGoImportTarget(rawPath, fileSet)
		results = append(results, ExtractedImport{
			ImportPath: rawPath,
			TargetFile: target,
			LineNumber: line,
			Language:   LangGo,
			Raw:        imp.Path.Value,
		})
	}

	return results
}

func resolveGoImportTarget(importPath string, fileSet map[string]bool) string {
	// Look for a local package match
	// E.g. "commit-archaeologist/internal/db" -> search for internal/db/*.go in fileSet
	for candidate := range fileSet {
		if !strings.HasSuffix(candidate, ".go") {
			continue
		}
		dir := filepath.ToSlash(filepath.Dir(candidate))
		if strings.HasSuffix(importPath, "/"+dir) || importPath == dir || strings.HasSuffix(importPath, dir) {
			return candidate
		}
	}
	return ""
}

func fallbackScanGoImports(relPath string, content string, fileSet map[string]bool) []ExtractedImport {
	var results []ExtractedImport
	lines := strings.Split(content, "\n")
	inImportBlock := false

	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		lineNum := idx + 1

		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			continue
		}
		if inImportBlock && trimmed == ")" {
			inImportBlock = false
			continue
		}

		if inImportBlock || strings.HasPrefix(trimmed, "import ") {
			start := strings.Index(trimmed, `"`)
			end := strings.LastIndex(trimmed, `"`)
			if start != -1 && end > start {
				rawPath := trimmed[start+1 : end]
				target := resolveGoImportTarget(rawPath, fileSet)
				results = append(results, ExtractedImport{
					ImportPath: rawPath,
					TargetFile: target,
					LineNumber: lineNum,
					Language:   LangGo,
					Raw:        trimmed,
				})
			}
		}
	}

	return results
}

// ParseGoSymbols extracts function, struct, interface declarations from Go source code.
func ParseGoSymbols(relPath string, data []byte) []ParsedSymbol {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, relPath, data, 0)
	if err != nil {
		return nil
	}

	var symbols []ParsedSymbol
	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			symbols = append(symbols, ParsedSymbol{
				Name:       decl.Name.Name,
				Kind:       "function",
				LineNumber: fset.Position(decl.Pos()).Line,
				Exported:   decl.Name.IsExported(),
			})
		case *ast.TypeSpec:
			kind := "type"
			switch decl.Type.(type) {
			case *ast.StructType:
				kind = "struct"
			case *ast.InterfaceType:
				kind = "interface"
			}
			symbols = append(symbols, ParsedSymbol{
				Name:       decl.Name.Name,
				Kind:       kind,
				LineNumber: fset.Position(decl.Pos()).Line,
				Exported:   decl.Name.IsExported(),
			})
		}
		return true
	})

	return symbols
}
