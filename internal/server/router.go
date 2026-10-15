package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"commit-archaeologist/internal/analyzer"
	"commit-archaeologist/internal/db"
	"commit-archaeologist/internal/git"
)

type Server struct {
	store    *db.Store
	gitExec  *git.Executor
	analyzer *analyzer.Analyzer
	demoDir  string
	mux      *http.ServeMux
}

func NewServer(store *db.Store, gitExec *git.Executor, analyzer *analyzer.Analyzer, demoDir string) *Server {
	s := &Server{
		store:    store,
		gitExec:  gitExec,
		analyzer: analyzer,
		demoDir:  demoDir,
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS Middleware
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/settings", s.handleSettings)
	s.mux.HandleFunc("/api/repositories", s.handleRepositories)
	s.mux.HandleFunc("/api/analyze", s.handleAnalyze)
	s.mux.HandleFunc("/api/demo", s.handleDemo)
	s.mux.HandleFunc("/api/overview", s.handleOverview)
	s.mux.HandleFunc("/api/commits", s.handleCommits)
	s.mux.HandleFunc("/api/commit-detail", s.handleCommitDetail)
	s.mux.HandleFunc("/api/diff", s.handleDiff)
	s.mux.HandleFunc("/api/file-content", s.handleFileContent)
	s.mux.HandleFunc("/api/hotspots", s.handleHotspots)
	s.mux.HandleFunc("/api/dependencies", s.handleDependencies)
	s.mux.HandleFunc("/api/contributors", s.handleContributors)
	s.mux.HandleFunc("/api/activity", s.handleActivity)
	s.mux.HandleFunc("/api/timeline", s.handleTimeline)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseTimeParam(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return &t
	}
	// Try YYYY-MM-DD
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		return &t
	}
	return nil
}

func sanitizeStringParam(s string) string {
	return strings.TrimSpace(s)
}
