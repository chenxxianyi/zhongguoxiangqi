package httpapi

import (
	"log/slog"
	"net/http"

	"xiangqi-lab/internal/analysis"
	"xiangqi-lab/internal/config"
	"xiangqi-lab/internal/engine"
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
	mux.HandleFunc("GET /api/v1/matches/{id}/legal-moves", s.legalMoves)
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
