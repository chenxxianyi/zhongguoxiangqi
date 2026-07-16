package httpapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"xiangqi-lab/internal/analysis"
	"xiangqi-lab/internal/config"
	"xiangqi-lab/internal/engine/builtin"
	"xiangqi-lab/internal/game"
	"xiangqi-lab/internal/learning"
	"xiangqi-lab/internal/records"
)

func TestMatchHTTPFlowAndStructuredError(t *testing.T) {
	searchEngine := builtin.New()
	repository := game.NewMemoryRepository()
	events := game.NewEventBus()
	matches := game.NewService(repository, searchEngine, events, 50*time.Millisecond)
	recordService := records.NewService()
	server := NewServer(
		config.Config{AllowedOrigin: "http://localhost:5666", MaxUploadBytes: 2 << 20, DataMode: "memory"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		matches, events, recordService, learning.NewService(recordService),
		analysis.NewService(matches, searchEngine), searchEngine,
	)
	api := httptest.NewServer(server.Handler())
	defer api.Close()

	created := requestJSON[game.Snapshot](t, http.MethodPost, api.URL+"/api/v1/matches",
		map[string]any{"playerColor": "red", "difficulty": 1, "aiMode": "library"}, http.StatusCreated)
	if created.Version != 1 || created.SideToMove != "red" || created.AIMode != game.AIModeLibrary {
		t.Fatalf("created: %+v", created)
	}
	legalMoves := requestJSON[game.LegalMovesResponse](
		t, http.MethodGet, api.URL+"/api/v1/matches/"+created.ID+"/legal-moves?from=a3",
		nil, http.StatusOK,
	)
	if legalMoves.MatchVersion != created.Version || len(legalMoves.Moves) != 1 ||
		legalMoves.Moves[0].Move != "a3a4" {
		t.Fatalf("legal moves: %+v", legalMoves)
	}
	moved := requestJSON[game.Snapshot](t, http.MethodPost, api.URL+"/api/v1/matches/"+created.ID+"/moves",
		map[string]any{"move": "a3a4", "expectedMatchVersion": 1}, http.StatusAccepted)
	if len(moved.Moves) != 1 || moved.Status != game.StatusAIThinking {
		t.Fatalf("moved: %+v", moved)
	}
	errorResponse := requestJSON[errorBody](t, http.MethodPost, api.URL+"/api/v1/matches/"+created.ID+"/moves",
		map[string]any{"move": "c3c4", "expectedMatchVersion": 1}, http.StatusConflict)
	if errorResponse.Code != "MATCH_VERSION_CONFLICT" || errorResponse.RequestID == "" {
		t.Fatalf("error response: %+v", errorResponse)
	}
}

func TestHealthAndRecordImport(t *testing.T) {
	searchEngine := builtin.New()
	matches := game.NewService(game.NewMemoryRepository(), searchEngine, game.NewEventBus(), time.Second)
	recordService := records.NewService()
	server := NewServer(
		config.Config{MaxUploadBytes: 2 << 20, DataMode: "memory"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		matches, game.NewEventBus(), recordService, learning.NewService(recordService),
		analysis.NewService(matches, searchEngine), searchEngine,
	)
	api := httptest.NewServer(server.Handler())
	defer api.Close()
	readiness := requestJSON[struct {
		Dependencies map[string]string `json:"dependencies"`
	}](t, http.MethodGet, api.URL+"/health/ready", nil, http.StatusOK)
	if readiness.Dependencies["authoritativeStore"] != "memory" {
		t.Fatalf("authoritative store = %q, want memory", readiness.Dependencies["authoritativeStore"])
	}
	batch := requestJSON[records.ImportBatch](t, http.MethodPost, api.URL+"/api/v1/records/imports",
		map[string]any{"name": "demo", "format": "iccs", "content": "a3a4 a6a5"}, http.StatusCreated)
	if batch.ImportedGames != 1 {
		t.Fatalf("batch: %+v", batch)
	}
	multipartBatch := requestMultipartImport(t, api.URL+"/api/v1/records/imports", map[string]string{
		"name": "multipart demo", "format": "iccs",
	}, "demo.iccs", "c3c4 c6c5", http.StatusCreated)
	if multipartBatch.ImportedGames != 1 {
		t.Fatalf("multipart batch: %+v", multipartBatch)
	}
}

func TestWebSocketEventEnvelope(t *testing.T) {
	searchEngine := builtin.New()
	events := game.NewEventBus()
	matches := game.NewService(game.NewMemoryRepository(), searchEngine, events, time.Second)
	recordService := records.NewService()
	server := NewServer(
		config.Config{MaxUploadBytes: 2 << 20, DataMode: "memory"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		matches, events, recordService, learning.NewService(recordService),
		analysis.NewService(matches, searchEngine), searchEngine,
	)
	api := httptest.NewServer(server.Handler())
	defer api.Close()
	address, _ := url.Parse(api.URL)
	connection, err := net.Dial("tcp", address.Host)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(2 * time.Second))
	_, _ = io.WriteString(connection,
		"GET /api/v1/matches/match-ws/stream HTTP/1.1\r\n"+
			"Host: "+address.Host+"\r\n"+
			"Upgrade: websocket\r\nConnection: Upgrade\r\n"+
			"Sec-WebSocket-Version: 13\r\n"+
			"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n\r\n")
	reader := bufio.NewReader(connection)
	status, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(status, "101 Switching Protocols") {
		t.Fatalf("handshake status %q: %v", status, err)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line == "\r\n" {
			break
		}
	}
	time.Sleep(10 * time.Millisecond)
	events.Publish("match-ws", 2, "match.snapshot", map[string]string{"fen": "demo"})
	first, err := reader.ReadByte()
	if err != nil || first&0x0f != 0x1 {
		t.Fatalf("frame opcode: %x %v", first, err)
	}
	second, _ := reader.ReadByte()
	length := uint64(second & 0x7f)
	if length == 126 {
		var raw [2]byte
		_, _ = io.ReadFull(reader, raw[:])
		length = uint64(binary.BigEndian.Uint16(raw[:]))
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		t.Fatal(err)
	}
	var event game.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatal(err)
	}
	if event.Type != "match.snapshot" || event.MatchVersion != 2 || event.EventID == "" {
		t.Fatalf("event: %+v", event)
	}
}

func requestJSON[T any](t *testing.T, method, url string, body any, wantStatus int) T {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}
	request, err := http.NewRequestWithContext(context.Background(), method, url, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		data, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, want %d: %s", response.StatusCode, wantStatus, data)
	}
	var result T
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result
}

func requestMultipartImport(t *testing.T, url string, fields map[string]string, filename, content string, wantStatus int) records.ImportBatch {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	file, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(file, content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, &body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		data, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, want %d: %s", response.StatusCode, wantStatus, data)
	}
	var result records.ImportBatch
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result
}
