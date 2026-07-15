package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"xiangqi-lab/internal/analysis"
	"xiangqi-lab/internal/config"
	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/engine"
	"xiangqi-lab/internal/engine/difficulty"
	"xiangqi-lab/internal/game"
	"xiangqi-lab/internal/learning"
	"xiangqi-lab/internal/records"
)

type Server struct {
	config   config.Config
	logger   *slog.Logger
	matches  *game.Service
	events   *game.EventBus
	records  *records.Service
	learning *learning.Service
	analysis *analysis.Service
	engine   engine.Engine
	handler  http.Handler
}

type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
	Details   any    `json:"details,omitempty"`
}

type contextKey string

const requestIDKey contextKey = "request-id"

func NewServer(
	cfg config.Config,
	logger *slog.Logger,
	matches *game.Service,
	events *game.EventBus,
	recordService *records.Service,
	learningService *learning.Service,
	analysisService *analysis.Service,
	searchEngine engine.Engine,
) *Server {
	server := &Server{
		config: cfg, logger: logger, matches: matches, events: events,
		records: recordService, learning: learningService, analysis: analysisService,
		engine: searchEngine,
	}
	mux := http.NewServeMux()
	server.routes(mux)
	server.handler = server.middleware(mux)
	return server
}

func (s *Server) Handler() http.Handler { return s.handler }

func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health/live", s.live)
	mux.HandleFunc("GET /health/ready", s.ready)

	mux.HandleFunc("GET /api/v1/matches", s.listMatches)
	mux.HandleFunc("POST /api/v1/matches", s.createMatch)
	mux.HandleFunc("GET /api/v1/matches/{id}", s.getMatch)
	mux.HandleFunc("POST /api/v1/matches/{id}/moves", s.applyMove)
	mux.HandleFunc("POST /api/v1/matches/{id}/undo", s.undo)
	mux.HandleFunc("POST /api/v1/matches/{id}/resign", s.resign)
	mux.HandleFunc("POST /api/v1/matches/{id}/draw-offers", s.offerDraw)
	mux.HandleFunc("GET /api/v1/matches/{id}/stream", s.stream)

	mux.HandleFunc("POST /api/v1/records/imports", s.importRecords)
	mux.HandleFunc("GET /api/v1/records", s.listRecords)
	mux.HandleFunc("GET /api/v1/records/{rest...}", s.routeRecordsGet)
	mux.HandleFunc("DELETE /api/v1/records/{id}", s.deleteRecord)

	mux.HandleFunc("POST /api/v1/learning/jobs", s.createLearningJob)
	mux.HandleFunc("GET /api/v1/learning/jobs/{id}", s.getLearningJob)
	mux.HandleFunc("GET /api/v1/learning/versions", s.listLearningVersions)
	mux.HandleFunc("GET /api/v1/learning/versions/{id}", s.getLearningVersion)
	mux.HandleFunc("POST /api/v1/learning/versions/{id}/activate", s.activateLearningVersion)
	mux.HandleFunc("POST /api/v1/learning/versions/{id}/rollback", s.rollbackLearningVersion)

	mux.HandleFunc("POST /api/v1/analysis/jobs", s.createAnalysisJob)
	mux.HandleFunc("GET /api/v1/analysis/jobs/{id}", s.getAnalysisJob)
	mux.HandleFunc("GET /api/v1/matches/{id}/analysis", s.getMatchAnalysis)

	mux.HandleFunc("GET /api/v1/engines/health", s.engineHealth)
	mux.HandleFunc("GET /api/v1/difficulty-profiles", s.difficultyProfiles)
	mux.HandleFunc("GET /api/v1/about/licenses", s.licenses)
}

func (s *Server) live(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok", "service": "xiangqi-lab-api", "time": time.Now().UTC(),
	})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ready", "dataMode": s.config.DataMode,
		"dependencies": map[string]string{
			"authoritativeStore": "memory",
			"redis":              "not_configured_degraded",
			"externalEngine":     "not_configured_optional",
		},
	})
}

func (s *Server) listMatches(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) importRecords(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxUploadBytes)
	request, err := decodeImportRequest(r)
	if err != nil {
		s.writeError(w, r, fmt.Errorf("invalid import request: %w", err))
		return
	}
	batch := s.records.Import(request)
	writeJSON(w, http.StatusCreated, batch)
}

func (s *Server) getImport(w http.ResponseWriter, r *http.Request) {
	batch, err := s.records.GetImport(strings.TrimPrefix(r.PathValue("rest"), "imports/"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func (s *Server) listRecords(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.records.List()})
}

func (s *Server) getRecord(w http.ResponseWriter, r *http.Request) {
	record, err := s.records.Get(r.PathValue("rest"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *Server) getRecordMoves(w http.ResponseWriter, r *http.Request) {
	recordID := strings.TrimSuffix(r.PathValue("rest"), "/moves")
	record, err := s.records.Get(recordID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"recordId": record.ID, "items": record.Moves})
}

func (s *Server) routeRecordsGet(w http.ResponseWriter, r *http.Request) {
	rest := strings.Trim(r.PathValue("rest"), "/")
	switch {
	case strings.HasPrefix(rest, "imports/") && strings.Count(rest, "/") == 1:
		s.getImport(w, r)
	case strings.HasSuffix(rest, "/moves") && strings.Count(rest, "/") == 1:
		s.getRecordMoves(w, r)
	case rest != "" && !strings.Contains(rest, "/"):
		s.getRecord(w, r)
	default:
		s.writeError(w, r, records.ErrNotFound)
	}
}

func (s *Server) deleteRecord(w http.ResponseWriter, r *http.Request) {
	if err := s.records.Delete(r.PathValue("id")); err != nil {
		s.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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

func (s *Server) listLearningVersions(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) difficultyProfiles(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": difficulty.Profiles()})
}

func (s *Server) licenses(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"application": "Project source license is determined by the repository owner.",
		"externalEngines": []map[string]string{{
			"name": "Pikafish", "status": "not bundled",
			"notice": "Optional external process. Confirm GPLv3 obligations before distribution.",
		}},
	})
}

func (s *Server) decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		s.writeError(w, r, fmt.Errorf("invalid JSON request: %w", err))
		return false
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		s.writeError(w, r, errors.New("request body must contain one JSON object"))
		return false
	}
	return true
}

func decodeImportRequest(r *http.Request) (records.ImportRequest, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			return records.ImportRequest{}, err
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			return records.ImportRequest{}, err
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return records.ImportRequest{}, err
		}
		format := r.FormValue("format")
		if format == "" {
			format = extensionFormat(header)
		}
		return records.ImportRequest{
			Name: r.FormValue("name"), Format: format, Content: string(content),
			InitialFEN: r.FormValue("initialFen"), Result: r.FormValue("result"),
		}, nil
	}
	var request records.ImportRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return records.ImportRequest{}, err
	}
	return request, nil
}

func extensionFormat(header *multipart.FileHeader) string {
	name := strings.ToLower(header.Filename)
	switch {
	case strings.HasSuffix(name, ".json"):
		return "json"
	case strings.HasSuffix(name, ".pgn"):
		return "pgn"
	default:
		return "iccs"
	}
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusBadRequest, "BAD_REQUEST", err.Error()
	switch {
	case errors.Is(err, game.ErrNotFound), errors.Is(err, records.ErrNotFound),
		errors.Is(err, learning.ErrNotFound), errors.Is(err, analysis.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "请求的资源不存在"
	case errors.Is(err, game.ErrVersionConflict):
		status, code, message = http.StatusConflict, "MATCH_VERSION_CONFLICT", "对局版本已变化，请先获取最新快照"
	case errors.Is(err, game.ErrStateConflict), errors.Is(err, game.ErrNotPlayerTurn),
		errors.Is(err, game.ErrUndoDisabled), errors.Is(err, game.ErrNoMovesToUndo),
		errors.Is(err, game.ErrIdempotency), errors.Is(err, learning.ErrVersionNotReady):
		status, code = http.StatusConflict, "STATE_CONFLICT"
	case errors.Is(err, xiangqi.ErrIllegalMove), errors.Is(err, xiangqi.ErrWrongTurn):
		status, code, message = http.StatusUnprocessableEntity, "ILLEGAL_MOVE", "该着法在当前权威局面中不合法"
	}
	if strings.Contains(err.Error(), "request body too large") {
		status, code, message = http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "请求内容超过允许大小"
	}
	writeJSON(w, status, errorBody{
		Code: code, Message: message, RequestID: requestID(r.Context()),
	})
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" || len(requestID) > 128 {
			requestID = randomID()
		}
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if origin := r.Header.Get("Origin"); origin != "" && origin == s.config.AllowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key, X-Request-ID")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				s.logger.Error("http panic", "requestId", requestID, "error", fmt.Sprint(recovered))
				writeJSON(w, http.StatusInternalServerError, errorBody{
					Code: "INTERNAL_ERROR", Message: "服务内部错误", RequestID: requestID,
				})
			}
			s.logger.Info("http request", "requestId", requestID, "method", r.Method,
				"path", r.URL.Path, "durationMs", time.Since(started).Milliseconds())
		}()
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeSSE(w io.Writer, event game.Event) {
	data, _ := json.Marshal(event)
	_, _ = fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", event.EventID, event.Type, data)
}

func requestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

func randomID() string {
	var data [12]byte
	_, _ = rand.Read(data[:])
	return hex.EncodeToString(data[:])
}
