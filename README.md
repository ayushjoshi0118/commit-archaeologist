# Commit Archaeologist

[![CI Pipeline](https://github.com/commit-archaeologist/commit-archaeologist/actions/workflows/ci.yml/badge.svg)](https://github.com/commit-archaeologist/commit-archaeologist/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![Svelte](https://img.shields.io/badge/Svelte-5.0-FF3E00?logo=svelte)](https://svelte.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?logo=typescript)](https://www.typescriptlang.org)
[![Tauri](https://img.shields.io/badge/Tauri-v2-24C8D8?logo=tauri)](https://tauri.app)

**Commit Archaeologist** is a high-performance desktop developer tool designed to explore the evolution, code churn, and architectural structure of Git repositories through interactive visual intelligence.

Inspired by modern code observability platforms (Linear, Sourcegraph, Datadog), it combines a **pure-Go analysis engine**, **AST-based dependency extraction**, **embedded SQLite persistence**, and **interactive D3.js visualizations** inside a sleek dark-themed Svelte 5 desktop shell.

---

## ⚡ Key Capabilities

- **Interactive Commit Timeline**: Explore branch topologies, merge histories, and author swimlanes via a zoomable D3 DAG canvas.
- **Code Hotspots & Churn Analytics**: Discover files with high modification frequency, multi-author volatility, and architectural complexity using interactive Treemap and Churn vs. LOC scatter matrices.
- **Multi-Language AST Dependency Graph**: Parse and visualize source file imports and cross-file dependencies in a force-directed physics simulation with drag, zoom, and neighbor isolation.
- **Commit Explorer & Unified Diffs**: Filter commits by author, date, message, or file path with side-by-side file tabs and syntax-highlighted hunk diffs.
- **Contributor Analytics & 7x24 Punchcard**: Inspect developer work cadences and temporal commit distributions across days of the week and hours of the day.
- **Zero-Config Demo Mode**: Instantly test all features using a bundled, synthetic multi-language repository ("ChronoLog") featuring realistic commit histories and inter-file dependencies.
- **Local-First & Air-Gapped**: 100% of the analysis, parsing, and caching happens on your local machine. No source code or metadata is ever transmitted to external servers.

---

## 🛠 Tech Stack

| Layer | Technology | Rationale |
|---|---|---|
| **Analysis Backend** | Go (1.22+) | Fast, safe concurrency, direct Git execution, zero-overhead AST parsing |
| **AST / Parsing** | `go/ast`, Tree-sitter, Regex AST | Multi-language import resolution and symbol extraction |
| **Desktop Shell** | Tauri v2 / Modern Webview | Lightweight native desktop shell with low memory footprint |
| **Frontend UI** | Svelte 5 + TypeScript + Vite | Reactive, compilation-driven UI with instant state synchronization |
| **Visualizations** | D3.js v7 | Custom interactive force-directed networks, treemaps, and timelines |
| **Local Persistence** | SQLite (`modernc.org/sqlite`) | CGo-free, pure-Go embedded database for cross-platform persistence |
| **Testing** | Go testing + Vitest | Full backend unit/integration tests and frontend component tests |

---

## 🚀 Quick Start

### Prerequisites
- [Go](https://go.dev/dl/) (>= 1.22)
- [Node.js](https://nodejs.org/) (>= 18) and [npm](https://www.npmjs.com/)
- [Git](https://git-scm.com/) installed and on system PATH
- *(Optional)* [Rust](https://www.rust-lang.org/) (only if compiling native Tauri desktop installers)

### 1. Clone & Install Dependencies

```bash
git clone https://github.com/your-org/commit-archaeologist.git
cd commit-archaeologist

# Install frontend dependencies
npm install

# Build backend executable
go build -o bin/server.exe ./cmd/server
```

### 2. Launch Development Environment

Run the Go backend server:
```bash
go run ./cmd/server
```
*The local API server starts at `http://127.0.0.1:3456` and initializes the local database at `~/.archeologist/archeologist.db`.*

In a separate terminal, start the frontend:
```bash
npm run dev
```
Open **`http://localhost:5173`** in your browser.

### 3. Launch Desktop Shell (Tauri)

If Rust/Cargo is installed on your machine:
```bash
npm run tauri dev
```

---

## 🧪 Testing & Verification

Both the backend and frontend include automated test suites:

```bash
# Run all Go unit and integration tests
go test -v ./...

# Run frontend Vitest test suite
npm run test

# Run production frontend bundle build
npm run build
```

---

## 🧠 Supported Languages & AST Capabilities

Commit Archaeologist includes an extensible AST dependency resolution pipeline:

| Language | Extensions | Extracted Dependencies |
|---|---|---|
| **Go** | `.go` | Internal package imports, standard library exclusion, function & struct symbols |
| **TypeScript** | `.ts`, `.tsx` | ES6 imports, export from, dynamic imports, path alias resolution (`$lib`, `@/`) |
| **JavaScript** | `.js`, `.jsx`, `.mjs`, `.cjs` | CommonJS `require()`, ES6 imports, dynamic imports |
| **Python** | `.py` | `from .module import`, `import module`, package hierarchy resolution |

### Limitations
- **Binary & Media Files**: Non-text assets (`.png`, `.wasm`, `.zip`) are excluded from AST parsing.
- **Dynamic Computed Imports**: Non-literal import paths (e.g. `require(dynamicVar)`) are recorded as unresolvable.
- **External Dependencies**: Remote third-party packages (e.g. `npm` packages, Go standard library, Python pip packages) are excluded from the inter-file repository dependency graph to emphasize internal architectural links.

---

## 🔒 Security & Privacy

1. **Strict Input Sanitization**: No git commands are constructed via string interpolation or passed to raw shell interpreters.
2. **Directory Traversal Protection**: All repository paths and file diff targets are strictly validated and canonicalized.
3. **No External Network Calls**: The backend server binds strictly to loopback (`127.0.0.1`).
4. **AI Features Disabled by Default**: No code or metadata is sent to any external AI provider.

---

## 📂 Project Structure

```
commit-archaeologist/
├── .github/
│   ├── workflows/ci.yml         # GitHub Actions CI workflow
│   └── ISSUE_TEMPLATE/          # Bug report and feature request templates
├── cmd/
│   └── server/main.go           # Go backend entrypoint
├── internal/
│   ├── analyzer/                # Git analysis orchestration & hotspot scoring
│   ├── ast/                     # Multi-language AST import & dependency extractors
│   ├── db/                      # SQLite persistence store & schema migrations
│   ├── demo/                    # Synthetic demo repo generator ("ChronoLog")
│   ├── git/                     # Safe Git invoker, log parser, and diff engine
│   └── server/                  # HTTP REST API router, handlers, and CORS middleware
├── src-tauri/                   # Tauri v2 desktop shell configuration & Rust bridge
├── src/
│   ├── lib/
│   │   ├── api/client.ts        # Typed frontend API client
│   │   ├── components/          # Svelte views and interactive D3 visualizations
│   │   ├── stores/app.ts        # Reactive state stores
│   │   └── types/index.ts       # Shared TypeScript models
│   ├── App.svelte               # Main application layout and view router
│   ├── app.css                  # Tailwind styles and custom dark theme
│   └── main.ts                  # Svelte 5 application bootstrap
├── tests/                       # Vitest frontend tests
├── ARCHITECTURE.md              # System design & database schema documentation
├── CONTRIBUTING.md              # Contribution and development guidelines
├── LICENSE                      # MIT License
├── README.md                    # Project overview & documentation
└── ROADMAP.md                   # Feature development roadmap
```

---

## 📄 License

Licensed under the [MIT License](LICENSE).
