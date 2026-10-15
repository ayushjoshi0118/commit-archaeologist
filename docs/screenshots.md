# Application Visual Tour & Screenshot Guide

This guide describes each view in Commit Archaeologist and how to capture application screenshots for documentation or release notes.

---

## 1. Main Screens & Visual Guide

### 1.1 Overview Dashboard
- **Location**: Navigation -> `Overview`
- **Key Visual Elements**:
  - Top KPI cards: Total Commits, Active Authors, Tracked Files, AST Dependencies.
  - Interactive D3 Commit Velocity Area Chart (daily commit volume gradient).
  - Code Hotspots quick preview table with churn score badges.
  - Recent commits timeline with author and addition/deletion deltas.

### 1.2 Commit Timeline
- **Location**: Navigation -> `Commit Timeline`
- **Key Visual Elements**:
  - Horizontal timeline canvas with author swimlanes.
  - DAG bezier curves connecting parent commits.
  - Commit nodes color-coded by author.
  - Commit Inspector drawer displaying full commit subject, body, author, timestamp, parent hashes, and diff button.

### 1.3 Code Hotspots & Churn Analysis
- **Location**: Navigation -> `Code Hotspots`
- **Key Visual Elements**:
  - Treemap Heatmap: Area represents file size/LOC, warm color gradient represents churn score.
  - Churn vs LOC Matrix: Scatter bubble plot with LOC on X-axis and commit modifications on Y-axis.
  - Floating tooltip displaying commit count, unique author count, and line churn.

### 1.4 AST Dependency Graph
- **Location**: Navigation -> `Dependency Graph`
- **Key Visual Elements**:
  - Interactive D3 force-directed physics network.
  - Nodes colored by language badge (Cyan: Go, Blue: TypeScript, Yellow: JavaScript, Royal: Python).
  - Directional dependency arrows.
  - Selected node highlights upstream dependents and downstream dependencies.

### 1.5 Commit Explorer & Unified Diff Viewer
- **Location**: Navigation -> `Commit Explorer`
- **Key Visual Elements**:
  - Filterable commit list with search and file path filters.
  - Files changed pill tabs with addition/deletion indicators.
  - High-fidelity unified code diff with emerald additions, rose deletions, line numbers, and hunk headers.

### 1.6 Contributor Analytics & 7x24 Punchcard
- **Location**: Navigation -> `Contributor Analytics`
- **Key Visual Elements**:
  - 7x24 Punchcard Matrix: Bubble radius indicates commit activity at that hour of the day across days of the week.
  - Contributor cards ranking authors with additions, deletions, and active time spans.

---

## 2. Capturing High-Resolution Screenshots

To capture clean promotional screenshots:
1. Start the application:
   ```bash
   go run ./cmd/server
   npm run dev
   ```
2. Open `http://localhost:5173` in a browser or Tauri desktop window.
3. Click **"Launch Demo (ChronoLog)"** in the sidebar. This loads the bundled synthetic multi-language repository with complete history.
4. Use standard browser developer tools or OS capture:
   - **Chrome / Edge**: Press `Ctrl+Shift+P` -> type `Capture full size screenshot`.
   - **Windows**: `Win + Shift + S`
   - **macOS**: `Cmd + Shift + 4`
