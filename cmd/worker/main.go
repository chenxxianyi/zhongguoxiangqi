package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"xiangqi-lab/internal/observability"
)

func main() {
	logger := observability.NewLogger()
	address := os.Getenv("XIANGQI_WORKER_ADDR")
	if address == "" {
		address = ":8081"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"mode":   "standby",
			"note":   "memory-mode jobs run inside the API process",
		})
	})
	logger.Info("worker standby endpoint listening", "address", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		logger.Error("worker stopped", slog.Any("error", err))
		os.Exit(1)
	}
}
