package httpapi

import (
	"net/http"

	"xiangqi-lab/internal/learning"
)

func (s *Server) createLearningJob(w http.ResponseWriter, r *http.Request) {
	var request learning.CreateJobRequest
	if !s.decodeJSON(w, r, &request) {
		return
	}
	writeJSON(w, http.StatusAccepted, s.learning.CreateJob(request))
}

func (s *Server) getLearningJob(w http.ResponseWriter, r *http.Request) {
	job, err := s.learning.GetJob(r.PathValue("id"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) listLearningVersions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.learning.ListVersions()})
}

func (s *Server) getLearningVersion(w http.ResponseWriter, r *http.Request) {
	version, err := s.learning.GetVersion(r.PathValue("id"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, version)
}

func (s *Server) activateLearningVersion(w http.ResponseWriter, r *http.Request) {
	version, err := s.learning.Activate(r.PathValue("id"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, version)
}

func (s *Server) rollbackLearningVersion(w http.ResponseWriter, r *http.Request) {
	version, err := s.learning.Rollback(r.PathValue("id"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, version)
}
