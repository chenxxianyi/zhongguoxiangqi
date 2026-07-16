package httpapi

import (
	"net/http"

	"xiangqi-lab/internal/analysis"
)

func (s *Server) createAnalysisJob(w http.ResponseWriter, r *http.Request) {
	var request analysis.CreateJobRequest
	if !s.decodeJSON(w, r, &request) {
		return
	}
	job, err := s.analysis.CreateJob(request)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, job)
}

func (s *Server) getAnalysisJob(w http.ResponseWriter, r *http.Request) {
	job, err := s.analysis.GetJob(r.PathValue("id"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) getMatchAnalysis(w http.ResponseWriter, r *http.Request) {
	result, err := s.analysis.GetResult(r.PathValue("id"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
