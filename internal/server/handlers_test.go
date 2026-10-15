package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"testing"

	"commit-archaeologist/internal/analyzer"
	"commit-archaeologist/internal/db"
	"commit-archaeologist/internal/demo"
	"commit-archaeologist/internal/git"
)

func TestServerHandlers(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed, skipping server handlers test")
	}

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_server.db")
	store, err := db.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	gitExec, err := git.NewExecutor()
	if err != nil {
		t.Fatalf("failed to create git executor: %v", err)
	}

	an := analyzer.NewAnalyzer(store, gitExec)
	srv := NewServer(store, gitExec, an, tempDir)

	// 1. Health check
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for health, got %d", w.Code)
	}

	// 2. Settings
	req = httptest.NewRequest("GET", "/api/settings", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for settings, got %d", w.Code)
	}
	var setResp SettingsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &setResp); err != nil {
		t.Fatalf("failed to decode settings: %v", err)
	}
	if setResp.AIEnabled {
		t.Errorf("AI should be disabled by default")
	}

	// 3. Demo generation & analysis
	req = httptest.NewRequest("POST", "/api/demo", bytes.NewBuffer([]byte("{}")))
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for demo, got %d (body: %s)", w.Code, w.Body.String())
	}
	var overview db.OverviewStats
	if err := json.Unmarshal(w.Body.Bytes(), &overview); err != nil {
		t.Fatalf("failed to decode overview stats: %v", err)
	}
	repoID := overview.Repository.ID
	if repoID == "" {
		t.Fatalf("expected non-empty repoId")
	}

	// 4. Commits endpoint
	req = httptest.NewRequest("GET", "/api/commits?repoId="+repoID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for commits, got %d", w.Code)
	}
	var commits []db.Commit
	if err := json.Unmarshal(w.Body.Bytes(), &commits); err != nil {
		t.Fatalf("failed to decode commits: %v", err)
	}
	if len(commits) == 0 {
		t.Errorf("expected commits list to be non-empty")
	}

	// 5. Hotspots endpoint
	req = httptest.NewRequest("GET", "/api/hotspots?repoId="+repoID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for hotspots, got %d", w.Code)
	}

	// 6. Dependencies endpoint
	req = httptest.NewRequest("GET", "/api/dependencies?repoId="+repoID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for dependencies, got %d", w.Code)
	}

	// 7. Contributors & Activity endpoint
	req = httptest.NewRequest("GET", "/api/contributors?repoId="+repoID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for contributors, got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/api/activity?repoId="+repoID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for activity, got %d", w.Code)
	}

	// 8. Timeline endpoint
	req = httptest.NewRequest("GET", "/api/timeline?repoId="+repoID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for timeline, got %d", w.Code)
	}
}

func init() {
	_ = context.Background()
	_ = demo.EnsureSyntheticRepo
}
