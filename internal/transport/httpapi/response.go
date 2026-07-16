package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"xiangqi-lab/internal/analysis"
	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/game"
	"xiangqi-lab/internal/learning"
	"xiangqi-lab/internal/records"
)

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

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusBadRequest, "BAD_REQUEST", err.Error()
	switch {
	case errors.Is(err, game.ErrNotFound):
		status, code, message = http.StatusNotFound, "MATCH_NOT_FOUND", "对局不存在或已失效"
	case errors.Is(err, records.ErrNotFound), errors.Is(err, learning.ErrNotFound),
		errors.Is(err, analysis.ErrNotFound):
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
