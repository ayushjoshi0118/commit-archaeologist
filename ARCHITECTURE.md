# Commit Archaeologist Architecture Documentation

## 1. High-Level System Architecture

Commit Archaeologist is built on a clean separation of concerns between an **analytical Go backend engine** and a **responsive Svelte frontend shell** styled with modern observability aesthetics.

```mermaid
flowchart TD
    subgraph UI_Layer["Frontend Presentation Layer (Svelte + D3.js + TypeScript)"]
        SPA["Svelte Single Page Application"]
        Router["View Router (Overview, Timeline, Hotspots, Dependencies, Explorer, Analytics)"]
        Stores["Reactive State Stores (selectedRepo, filters, cache)"]
        D3Timeline["D3.js Commit Timeline (DAG & Swimlanes)"]
        D3Hotspots["D3.js Treemap & Churn Matrix"]
        D3Graph["D3.js Force-Directed Graph"]
        D3Punchcard["D3.js 7x24 Punchcard Matrix"]
        
        SPA --> Router
        Router --> Stores
        Router --> D3Timeline
        Router --> D3Hotspots
        Router --> D3Graph
        Router --> D3Punchcard
    end

    subgraph Bridge["IPC / Local REST API Bridge (127.0.0.1:3456)"]
        HTTP["HTTP / JSON REST API"]
    end

    subgraph Backend_Layer["Analytical Engine (Go stdlib & AST Parsers)"]
        ServerMux["HTTP Controller & CORS Middleware"]
        Analyzer["Repository Analyzer Engine"]
        GitExecutor["Safe Git Process Invoker (Sanitized Args)"]
        ASTEngine["AST & Dependency Extractor\n(Go, TS/JS, Python)"]
        DemoGen["Synthetic Demo Repo Engine ('ChronoLog')"]
        
        ServerMux --> Analyzer
        Analyzer --> GitExecutor
        Analyzer --> ASTEngine
        Analyzer --> DemoGen
    end

    subgraph Persistence_Layer["Local Storage Layer (SQLite)"]
        SQLite[("Embedded SQLite DB\n(modernc.org/sqlite - Pure Go)")]
    end

    SPA <-->|REST JSON| HTTP
    HTTP <--> ServerMux
    Analyzer <--> SQLite
```

---

## 2. Analytical Pipeline & Ingestion Flow

When a user opens a local Git directory (or triggers Demo mode):

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant UI as Svelte Frontend
    participant Server as Go API Server
    participant Git as Git CLI (Safe Invoker)
    participant AST as AST Engine
    participant DB as SQLite Store

    User->>UI: Select Local Repo Path
    UI->>Server: POST /api/analyze { path }
    Server->>Git: Validate directory & check .git
    Server->>Git: Run git log --root --numstat (with record delimiters)
    Git-->>Server: Stream commit log & change statistics
    Server->>Git: Run git ls-files (tracked source files)
    Git-->>Server: Tracked file paths
    Server->>AST: Parse file ASTs & resolve import dependencies
    AST-->>Server: Dependency edges & lines of code (LOC)
    Server->>Server: Calculate churn scores & contributor activity punchcards
    Server->>DB: Save commits, files, hotspots, deps, and authors (Transactional)
    DB-->>Server: Commit successful
    Server-->>UI: Return OverviewStats JSON payload
    UI->>User: Render interactive dashboard & D3 visualizations
```

---

## 3. SQLite Database Schema Design

The persistence layer stores all extracted metrics locally in `~/.archeologist/archeologist.db`:

```mermaid
erDiagram
    REPOSITORIES ||--o{ COMMITS : contains
    REPOSITORIES ||--o{ FILE_HOTSPOTS : tracks
    REPOSITORIES ||--o{ DEPENDENCIES : contains
    REPOSITORIES ||--o{ CONTRIBUTORS : records
    REPOSITORIES ||--o{ CONTRIBUTOR_ACTIVITY : records
    COMMITS ||--o{ COMMIT_FILES : touches

    REPOSITORIES {
        string id PK
        string name
        string path UK
        string default_branch
        datetime last_analyzed_at
        int total_commits
        int total_files
        int total_authors
        boolean is_demo
    }

    COMMITS {
        string hash PK
        string repo_id FK
        string short_hash
        string parent_hashes
        string author_name
        string author_email
        datetime author_date
        datetime committer_date
        string subject
        string body
        string branch
        int insertions
        int deletions
        int files_changed_count
    }

    COMMIT_FILES {
        int id PK
        string commit_hash FK
        string repo_id
        string file_path
        string change_type
        string old_path
        int insertions
        int deletions
    }

    FILE_HOTSPOTS {
        string repo_id PK,FK
        string file_path PK
        string extension
        int commit_count
        int author_count
        int total_insertions
        int total_deletions
        int lines_of_code
        datetime last_modified
        real churn_score
    }

    DEPENDENCIES {
        int id PK
        string repo_id FK
        string source_file
        string target_file
        string import_statement
        string language
        int line_number
    }

    CONTRIBUTORS {
        string repo_id PK,FK
        string email PK
        string name
        int total_commits
        datetime first_commit_date
        datetime last_commit_date
        int insertions
        int deletions
    }

    CONTRIBUTOR_ACTIVITY {
        string repo_id PK,FK
        string email PK
        int day_of_week PK
        int hour_of_day PK
        int commit_count
    }
```

---

## 4. Architectural Algorithms

### 4.1 Churn Scoring Formula
Hotspots are calculated by combining commit frequency ($C$), author diffusion ($A$), and modification volume ($I + D$):

$$\text{ChurnScore} = (C \times 1.5) + (A \times 2.0) + (\log_2(1 + I + D) \times 1.2)$$

- Files with multiple modifying authors score higher due to increased coordination overhead.
- Logarithmic scaling on total line delta prevents massive generated files or one-off refactors from disproportionately skewing the metrics.

### 4.2 Multi-Language AST Dependency Resolution
- **Go**: Uses `go/parser` and `go/token` to extract import blocks. Internal module package paths (e.g. `myproject/internal/db`) are resolved to physical `.go` files in the repository.
- **TypeScript / JavaScript**: Parses ES6 `import ... from '...'`, `export ... from '...'`, `require('...')`, and dynamic `import('...')`. Relative paths (`./components/Header`) and alias paths (`$lib/`, `@/`) are matched against candidate extensions (`.ts`, `.tsx`, `.svelte`, `.js`, `/index.ts`).
- **Python**: Scans `from <mod> import <sym>` and `import <mod>`. Resolves relative dot notation (`.config`) and package hierarchies (`pkg.subpkg`) to matching `.py` or `__init__.py` files.

---

## 5. Security & Isolation Model

- **Safe Subprocess Execution**: Arguments are passed strictly as formatted slices to `exec.CommandContext()`. No strings are ever passed through a shell interpreter (`cmd.exe` or `/bin/sh`), preventing argument injection vulnerabilities.
- **Air-Gapped Operation**: The Go server binds exclusively to `127.0.0.1`. No data leaves the local machine.
- **AI Opt-in Guarantee**: AI features remain disabled by default with zero external API dependencies.
