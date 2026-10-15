# Contributing to Commit Archaeologist

Thank you for your interest in contributing to **Commit Archaeologist**! We welcome bug reports, language parser additions, feature proposals, and pull requests.

---

## 1. Development Principles

1. **Local-First & Privacy Preserving**:
   - Analysis must remain 100% local. Never transmit source code or repository metadata to external networks.
   - AI features are strictly opt-in and disabled by default.

2. **Safe System Interaction**:
   - Never execute raw shell commands constructed with string interpolation.
   - Always validate paths, check for directory traversals (`../`), and sanitize arguments before calling `git`.

3. **High-Fidelity Visualizations**:
   - Visualizations must be grounded in real analyzed repository data from SQLite, not hardcoded mock metrics (except when explicitly running synthetic demo mode).

---

## 2. Setting Up Your Environment

### Prerequisites
- **Go** >= 1.22
- **Node.js** >= 18 and **npm** >= 9
- **Git** >= 2.30
- (Optional) **Rust** & **Cargo** if building native Tauri desktop installer bundles.

### Clone & Setup

```bash
git clone https://github.com/your-org/commit-archaeologist.git
cd commit-archaeologist

# Install frontend dependencies
npm install

# Build backend executable
go build -o bin/server ./cmd/server
```

---

## 3. Running Locally

### Option A: Concurrent Development (Recommended)

1. Start the backend analysis server:
   ```bash
   go run ./cmd/server
   ```
   *Listens on `http://127.0.0.1:3456` and initializes `~/.archeologist/archeologist.db`.*

2. In a separate terminal, start the Vite frontend dev server:
   ```bash
   npm run dev
   ```
   *Accessible at `http://localhost:5173` with hot module replacement and automatic backend proxying.*

### Option B: Desktop Shell via Tauri (Requires Rust/Cargo)

```bash
npm run tauri dev
```

---

## 4. Running the Test Suite

Both backend and frontend must pass all tests before submitting a Pull Request:

```bash
# Run all Go backend unit & integration tests
go test -v ./...

# Run frontend Vitest tests
npm run test

# Check frontend production build
npm run build
```

---

## 5. Adding AST Parsers for New Languages

To contribute an import/dependency extractor for a new programming language:
1. Open [`internal/ast/`](internal/ast/).
2. Add language identification to `DetectLanguage()` in [`internal/ast/parser.go`](internal/ast/parser.go).
3. Implement `<language>.go` with import extraction and relative path resolution logic.
4. Add comprehensive unit tests in [`internal/ast/parser_test.go`](internal/ast/parser_test.go).

---

## 6. Pull Request Guidelines

- Branch names should be descriptive: `feat/rust-ast-parser`, `fix/diff-newline-handling`.
- Keep commits atomic with conventional commit messages (`feat: ...`, `fix: ...`, `docs: ...`, `test: ...`).
- Ensure all CI tests pass.
