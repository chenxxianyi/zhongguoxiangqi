package httpapi

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"xiangqi-lab/internal/game"
)

func (s *Server) listMatches(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.matches.List()})
}

func (s *Server) createMatch(w http.ResponseWriter, r *http.Request) {
	var request game.CreateRequest
	if !s.decodeJSON(w, r, &request) {
		return
	}
	match, err := s.matches.Create(request, r.Header.Get("Idempotency-Key"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, match)
}

func (s *Server) getMatch(w http.ResponseWriter, r *http.Request) {
	match, err := s.matches.Get(r.PathValue("id"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, match)
}

func (s *Server) legalMoves(w http.ResponseWriter, r *http.Request) {
	moves, err := s.matches.LegalMoves(r.PathValue("id"), r.URL.Query().Get("from"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, moves)
}

func (s *Server) applyMove(w http.ResponseWriter, r *http.Request) {
	var request game.MoveRequest
	if !s.decodeJSON(w, r, &request) {
		return
	}
	match, err := s.matches.ApplyPlayerMove(r.PathValue("id"), request, r.Header.Get("Idempotency-Key"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, match)
}

func (s *Server) undo(w http.ResponseWriter, r *http.Request) {
	var request game.VersionRequest
	if !s.decodeJSON(w, r, &request) {
		return
	}
	match, err := s.matches.Undo(r.PathValue("id"), request, r.Header.Get("Idempotency-Key"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, match)
}

func (s *Server) resign(w http.ResponseWriter, r *http.Request) {
	var request game.VersionRequest
	if !s.decodeJSON(w, r, &request) {
		return
	}
	match, err := s.matches.Resign(r.PathValue("id"), request, r.Header.Get("Idempotency-Key"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, match)
}

func (s *Server) offerDraw(w http.ResponseWriter, r *http.Request) {
	var request game.VersionRequest
	if !s.decodeJSON(w, r, &request) {
		return
	}
	match, accepted, err := s.matches.OfferDraw(r.PathValue("id"), request, r.Header.Get("Idempotency-Key"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accepted": accepted, "match": match})
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		s.websocket(w, r)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		s.writeError(w, r, errors.New("streaming is unavailable"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	after := r.Header.Get("Last-Event-ID")
	if after == "" {
		after = r.URL.Query().Get("afterEventId")
	}
	channel, history, cancel := s.events.Subscribe(r.PathValue("id"), after)
	defer cancel()
	for _, event := range history {
		writeSSE(w, event)
	}
	flusher.Flush()
	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case event := <-channel:
			writeSSE(w, event)
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = io.WriteString(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
