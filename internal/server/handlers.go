package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"commit-archaeologist/internal/db"
	"commit-archaeologist/internal/demo"
)

type AnalyzeRequest struct {
	Path string `json:"path"`
}

type SettingsResponse struct {
	AppName          string   `json:"appName"`
	Version          string   `json:"version"`
	AIEnabled        bool     `json:"aiEnabled"`
	LocalOnly        bool     `json:"localOnly"`
	SupportedParsers []string `json:"supportedParsers"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "commit-archaeologist",
	})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, SettingsResponse{
		AppName:          "Commit Archaeologist",
		Version:          "1.0.0",
		AIEnabled:        false, // Explicitly disabled by default as required
		LocalOnly:        true,  // Local by default, no code leaves machine
		SupportedParsers: []string{"Go", "TypeScript", "JavaScript", "Python"},
	})
}

func (s *Server) handleRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repos, err := s.store.ListRepositories(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list repositories: "+err.Error())
		return
	}
	if repos == nil {
		repos = []db.Repository{}
	}
	writeJSON(w, http.StatusOK, repos)
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path cannot be empty")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	overview, err := s.analyzer.AnalyzeRepository(ctx, req.Path, false, nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, "analysis failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, overview)
}

func (s *Server) handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	repoDir, err := demo.EnsureSyntheticRepo(ctx, s.demoDir, s.gitExec)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate demo repository: "+err.Error())
		return
	}

	overview, err := s.analyzer.AnalyzeRepository(ctx, repoDir, true, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to analyze demo repository: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, overview)
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	repoID := sanitizeStringParam(r.URL.Query().Get("repoId"))
	if repoID == "" {
		writeError(w, http.StatusBadRequest, "repoId parameter required")
		return
	}

	overview, err := s.store.GetOverviewStats(r.Context(), repoID)
	if err != nil {
		writeError(w, http.StatusNotFound, "overview stats not found: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, overview)
}

func (s *Server) handleCommits(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	repoID := sanitizeStringParam(q.Get("repoId"))
	if repoID == "" {
		writeError(w, http.StatusBadRequest, "repoId parameter required")
		return
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := db.CommitFilter{
		RepoID:    repoID,
		Author:    sanitizeStringParam(q.Get("author")),
		Search:    sanitizeStringParam(q.Get("search")),
		FilePath:  sanitizeStringParam(q.Get("file")),
		StartDate: parseTimeParam(q.Get("startDate")),
		EndDate:   parseTimeParam(q.Get("endDate")),
		Branch:    sanitizeStringParam(q.Get("branch")),
		Limit:     limit,
		Offset:    offset,
	}

	commits, err := s.store.QueryCommits(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query commits: "+err.Error())
		return
	}
	if commits == nil {
		commits = []db.Commit{}
	}

	writeJSON(w, http.StatusOK, commits)
}

func (s *Server) handleCommitDetail(w http.ResponseWriter, r *http.Request) {
	hash := sanitizeStringParam(r.URL.Query().Get("hash"))
	if hash == "" {
		writeError(w, http.StatusBadRequest, "hash parameter required")
		return
	}

	commit, err := s.store.GetCommitDetails(r.Context(), hash)
	if err != nil || commit == nil {
		writeError(w, http.StatusNotFound, "commit not found")
		return
	}

	writeJSON(w, http.StatusOK, commit)
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	repoID := sanitizeStringParam(q.Get("repoId"))
	hash := sanitizeStringParam(q.Get("hash"))
	filePath := sanitizeStringParam(q.Get("file"))

	if repoID == "" || hash == "" {
		writeError(w, http.StatusBadRequest, "repoId and hash parameters required")
		return
	}

	repo, err := s.store.GetRepository(r.Context(), repoID)
	if err != nil || repo == nil {
		writeError(w, http.StatusNotFound, "repository not found")
		return
	}

	diffRes, err := s.gitExec.GetCommitDiff(r.Context(), repo.Path, hash, filePath)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to get diff: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, diffRes)
}

func (s *Server) handleFileContent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	repoID := sanitizeStringParam(q.Get("repoId"))
	hash := sanitizeStringParam(q.Get("hash"))
	filePath := sanitizeStringParam(q.Get("file"))

	if repoID == "" || hash == "" || filePath == "" {
		writeError(w, http.StatusBadRequest, "repoId, hash, and file parameters required")
		return
	}

	repo, err := s.store.GetRepository(r.Context(), repoID)
	if err != nil || repo == nil {
		writeError(w, http.StatusNotFound, "repository not found")
		return
	}

	content, err := s.gitExec.GetFileContent(r.Context(), repo.Path, hash, filePath)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to get file content: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"filePath": filePath,
		"content":  content,
	})
}

func (s *Server) handleHotspots(w http.ResponseWriter, r *http.Request) {
	repoID := sanitizeStringParam(r.URL.Query().Get("repoId"))
	if repoID == "" {
		writeError(w, http.StatusBadRequest, "repoId parameter required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 100
	}

	hotspots, err := s.store.GetHotspots(r.Context(), repoID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get hotspots: "+err.Error())
		return
	}
	if hotspots == nil {
		hotspots = []db.FileHotspot{}
	}

	writeJSON(w, http.StatusOK, hotspots)
}

func (s *Server) handleDependencies(w http.ResponseWriter, r *http.Request) {
	repoID := sanitizeStringParam(r.URL.Query().Get("repoId"))
	if repoID == "" {
		writeError(w, http.StatusBadRequest, "repoId parameter required")
		return
	}

	graph, err := s.store.GetDependencyGraph(r.Context(), repoID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get dependency graph: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, graph)
}

func (s *Server) handleContributors(w http.ResponseWriter, r *http.Request) {
	repoID := sanitizeStringParam(r.URL.Query().Get("repoId"))
	if repoID == "" {
		writeError(w, http.StatusBadRequest, "repoId parameter required")
		return
	}

	contributors, err := s.store.GetContributors(r.Context(), repoID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get contributors: "+err.Error())
		return
	}
	if contributors == nil {
		contributors = []db.Contributor{}
	}

	writeJSON(w, http.StatusOK, contributors)
}

func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
	repoID := sanitizeStringParam(r.URL.Query().Get("repoId"))
	email := sanitizeStringParam(r.URL.Query().Get("email"))
	if repoID == "" {
		writeError(w, http.StatusBadRequest, "repoId parameter required")
		return
	}

	activities, err := s.store.GetContributorActivity(r.Context(), repoID, email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get activity: "+err.Error())
		return
	}
	if activities == nil {
		activities = []db.ContributorActivity{}
	}

	writeJSON(w, http.StatusOK, activities)
}

func (s *Server) handleTimeline(w http.ResponseWriter, r *http.Request) {
	repoID := sanitizeStringParam(r.URL.Query().Get("repoId"))
	if repoID == "" {
		writeError(w, http.StatusBadRequest, "repoId parameter required")
		return
	}

	timeline, err := s.store.GetTimeline(r.Context(), repoID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get timeline: "+err.Error())
		return
	}
	if timeline == nil {
		timeline = []db.TimelinePoint{}
	}

	writeJSON(w, http.StatusOK, timeline)
}
