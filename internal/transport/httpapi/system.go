package httpapi

import (
	"context"
	"net/http"
	"time"

	"xiangqi-lab/internal/engine/difficulty"
)

func (s *Server) live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok", "service": "xiangqi-lab-api", "time": time.Now().UTC(),
	})
}

func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ready", "dataMode": s.config.DataMode,
		"dependencies": map[string]string{
			"authoritativeStore": s.matches.AuthoritativeStore(),
			"redis":              "not_configured_degraded",
			"externalEngine":     "not_configured_optional",
		},
	})
}

func (s *Server) engineHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()
	if err := s.engine.Health(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"name": s.engine.Name(), "status": "unavailable", "fallbackAvailable": false,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"name": s.engine.Name(), "status": "healthy", "type": "builtin",
		"externalEngine": "not_configured",
	})
}

func (s *Server) difficultyProfiles(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": difficulty.Profiles()})
}

func (s *Server) licenses(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"application": "Project source license is determined by the repository owner.",
		"externalEngines": []map[string]string{{
			"name": "Pikafish", "status": "not bundled",
			"notice": "Optional external process. Confirm GPLv3 obligations before distribution.",
		}},
	})
}
