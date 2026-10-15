# Commit Archaeologist Roadmap

This roadmap outlines key architectural milestones and upcoming capabilities planned for Commit Archaeologist.

---

## 🧭 Released (v1.0.0 - Current)

- [x] **Safe Git Execution Engine**: Sanitized argument vector execution, path validation, directory traversal prevention.
- [x] **Local Persistence**: Pure-Go SQLite persistence (`modernc.org/sqlite`) for commits, file stats, hotspots, dependencies, and author activity.
- [x] **Multi-Language AST Dependency Extractor**:
  - Go import and package symbol resolver (`go/parser` & `go/ast`)
  - TypeScript & JavaScript ES6/CJS/Dynamic import graph extractor
  - Python relative and absolute module dependency resolver
- [x] **Interactive D3.js Visualizations**:
  - Chronological commit timeline with DAG bezier links and swimlanes
  - Code Hotspots with Treemap and Churn vs LOC Complexity matrix
  - Force-directed interactive dependency network with zoom, pan, drag, and neighbor highlight
  - Contributor activity 7x24 punchcard matrix
- [x] **Commit Explorer & Diff Viewer**: Full unified diff viewer with addition/deletion styling and file tabs.
- [x] **Zero-Config Demo Mode**: Bundled synthetic sample repository ("ChronoLog") with realistic multi-branch git history and cross-file dependencies.
- [x] **Complete Testing**: 100% passing Go tests (`go test ./...`) and Vitest frontend suite (`npm run test`).

---

## 🚀 Near-Term (v1.1 - v1.2)

- [ ] **Additional Tree-Sitter Grammars**:
  - Rust (`use crate::...`, `mod ...`)
  - Java / Kotlin (`import com.example...`)
  - C / C++ (`#include "..."`)
- [ ] **Submodule & Monorepo Support**:
  - Multi-root Git workspaces with independent package subtrees.
  - Per-package churn and dependency boundaries (e.g. Nx, Turborepo, pnpm workspaces).
- [ ] **Export & Reporting Engine**:
  - Export SVG / PNG snapshots of interactive dependency graphs and commit timelines.
  - Export CSV / JSON summary of churn hotspots and risk metrics for CI audits.
- [ ] **Coupled File Analysis (Co-change Mining)**:
  - Temporal coupling detector: identify files that are frequently modified together in the same commits despite having no direct AST imports.

---

## 🔮 Future Explorations (v2.0)

- [ ] **Blame Map & Ownership Evolution**:
  - Interactive file line ownership heatmap showing code decay and original author turnover.
- [ ] **Architectural Boundary Rules**:
  - Custom dependency linting (e.g., alert if UI components directly import DB storage packages).
- [ ] **Optional Local LLM Explanation**:
  - 100% local, offline Ollama / llama.cpp integration to synthesize release notes or explain complex multi-file diffs without sending code to cloud services.
