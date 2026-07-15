package records

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"xiangqi-lab/internal/domain/xiangqi"
)

var (
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("record already exists")
)

type ImportRequest struct {
	Name       string   `json:"name"`
	Format     string   `json:"format"`
	Content    string   `json:"content"`
	InitialFEN string   `json:"initialFen,omitempty"`
	Result     string   `json:"result,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

type ImportError struct {
	Ply     int    `json:"ply,omitempty"`
	Token   string `json:"token,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ImportBatch struct {
	ID             string        `json:"id"`
	Status         string        `json:"status"`
	Name           string        `json:"name"`
	Format         string        `json:"format"`
	TotalGames     int           `json:"totalGames"`
	ImportedGames  int           `json:"importedGames"`
	DuplicateGames int           `json:"duplicateGames"`
	FailedGames    int           `json:"failedGames"`
	RecordIDs      []string      `json:"recordIds"`
	Errors         []ImportError `json:"errors"`
	CreatedAt      time.Time     `json:"createdAt"`
	CompletedAt    time.Time     `json:"completedAt"`
}

type Move struct {
	Ply       int    `json:"ply"`
	Move      string `json:"move"`
	Side      string `json:"side"`
	FENBefore string `json:"fenBefore"`
	FENAfter  string `json:"fenAfter"`
	HashAfter string `json:"hashAfter"`
}

type Record struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Format      string              `json:"format"`
	SourceHash  string              `json:"sourceHash"`
	ContentHash string              `json:"contentHash"`
	InitialFEN  string              `json:"initialFen"`
	FinalFEN    string              `json:"finalFen"`
	Result      string              `json:"result,omitempty"`
	Outcome     xiangqi.Outcome     `json:"outcome"`
	Termination xiangqi.Termination `json:"termination,omitempty"`
	MoveCount   int                 `json:"moveCount"`
	Moves       []Move              `json:"moves,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
}

type Service struct {
	mu          sync.RWMutex
	records     map[string]Record
	batches     map[string]ImportBatch
	contentHash map[string]string
}

func NewService() *Service {
	return &Service{
		records: make(map[string]Record), batches: make(map[string]ImportBatch),
		contentHash: make(map[string]string),
	}
}

func (s *Service) Import(request ImportRequest) ImportBatch {
	now := time.Now().UTC()
	batch := ImportBatch{
		ID: id(), Status: "processing", Name: cleanName(request.Name),
		Format: strings.ToLower(strings.TrimSpace(request.Format)), TotalGames: 1,
		RecordIDs: []string{}, Errors: []ImportError{}, CreatedAt: now,
	}
	if batch.Format == "" {
		batch.Format = "iccs"
	}
	s.mu.Lock()
	s.batches[batch.ID] = batch
	s.mu.Unlock()

	record, importErr := parseRecord(request, batch.Format)
	batch.CompletedAt = time.Now().UTC()
	batch.Status = "completed"
	if importErr != nil {
		batch.FailedGames = 1
		batch.Errors = append(batch.Errors, *importErr)
	} else {
		s.mu.Lock()
		if existingID, exists := s.contentHash[record.ContentHash]; exists {
			batch.DuplicateGames = 1
			batch.RecordIDs = append(batch.RecordIDs, existingID)
		} else {
			s.records[record.ID] = record
			s.contentHash[record.ContentHash] = record.ID
			batch.ImportedGames = 1
			batch.RecordIDs = append(batch.RecordIDs, record.ID)
		}
		s.mu.Unlock()
	}
	s.mu.Lock()
	s.batches[batch.ID] = batch
	s.mu.Unlock()
	return cloneBatch(batch)
}

func (s *Service) GetImport(id string) (ImportBatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	batch, ok := s.batches[id]
	if !ok {
		return ImportBatch{}, ErrNotFound
	}
	return cloneBatch(batch), nil
}

func (s *Service) List() []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Record, 0, len(s.records))
	for _, record := range s.records {
		item := cloneRecord(record)
		item.Moves = nil
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items
}

func (s *Service) AllWithMoves() []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Record, 0, len(s.records))
	for _, record := range s.records {
		items = append(items, cloneRecord(record))
	}
	return items
}

func (s *Service) Get(recordID string) (Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[recordID]
	if !ok {
		return Record{}, ErrNotFound
	}
	return cloneRecord(record), nil
}

func (s *Service) Delete(recordID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.records[recordID]
	if !ok {
		return ErrNotFound
	}
	delete(s.contentHash, record.ContentHash)
	delete(s.records, recordID)
	return nil
}

type projectJSON struct {
	Name       string   `json:"name"`
	InitialFEN string   `json:"initialFen"`
	Moves      []string `json:"moves"`
	Result     string   `json:"result"`
	Tags       []string `json:"tags"`
}

func parseRecord(request ImportRequest, format string) (Record, *ImportError) {
	name := cleanName(request.Name)
	initialFEN := strings.TrimSpace(request.InitialFEN)
	result := strings.TrimSpace(request.Result)
	tags := append([]string(nil), request.Tags...)
	var tokens []string
	switch format {
	case "json":
		var document projectJSON
		if err := json.Unmarshal([]byte(request.Content), &document); err != nil {
			return Record{}, &ImportError{Code: "INVALID_JSON", Message: "棋谱 JSON 无法解析"}
		}
		if document.Name != "" {
			name = cleanName(document.Name)
		}
		if document.InitialFEN != "" {
			initialFEN = document.InitialFEN
		}
		if document.Result != "" {
			result = document.Result
		}
		if len(document.Tags) > 0 {
			tags = document.Tags
		}
		tokens = document.Moves
	case "iccs", "txt", "pgn":
		tokens = coordinateTokens(request.Content)
	default:
		return Record{}, &ImportError{Code: "UNSUPPORTED_FORMAT", Message: "目前支持 iccs、txt、pgn 坐标着法和项目 JSON"}
	}
	if len(tokens) == 0 {
		return Record{}, &ImportError{Code: "NO_MOVES", Message: "未找到坐标着法"}
	}
	if len(tokens) > 1000 {
		return Record{}, &ImportError{Code: "TOO_MANY_MOVES", Message: "单盘棋谱最多允许 1000 个 ply"}
	}
	if initialFEN == "" {
		initialFEN = xiangqi.InitialFEN
	}
	position, err := xiangqi.ParseFEN(initialFEN)
	if err != nil {
		return Record{}, &ImportError{Code: "INVALID_INITIAL_FEN", Message: err.Error()}
	}
	moves := make([]Move, 0, len(tokens))
	canonical := make([]string, 0, len(tokens))
	for index, token := range tokens {
		move, err := xiangqi.ParseMove(token)
		if err != nil {
			return Record{}, &ImportError{Ply: index + 1, Token: token, Code: "INVALID_MOVE_FORMAT", Message: err.Error()}
		}
		if !position.IsLegal(move) {
			return Record{}, &ImportError{Ply: index + 1, Token: token, Code: "ILLEGAL_MOVE", Message: "该着在当前权威局面中不合法"}
		}
		before := position
		position, _, err = position.Apply(move)
		if err != nil {
			return Record{}, &ImportError{Ply: index + 1, Token: token, Code: "ILLEGAL_MOVE", Message: err.Error()}
		}
		canonical = append(canonical, move.ICCS())
		moves = append(moves, Move{
			Ply: index + 1, Move: move.ICCS(), Side: before.SideToMove().String(),
			FENBefore: before.FEN(), FENAfter: position.FEN(), HashAfter: fmt.Sprintf("%016x", position.Hash()),
		})
	}
	sourceSum := sha256.Sum256([]byte(request.Content))
	contentSum := sha256.Sum256([]byte(initialFEN + "|" + strings.Join(canonical, " ")))
	adjudication := position.Adjudicate()
	if name == "" {
		name = "导入棋谱 " + time.Now().Format("2006-01-02 15:04")
	}
	return Record{
		ID: id(), Name: name, Format: format,
		SourceHash: hex.EncodeToString(sourceSum[:]), ContentHash: hex.EncodeToString(contentSum[:]),
		InitialFEN: initialFEN, FinalFEN: position.FEN(), Result: result,
		Outcome: adjudication.Outcome, Termination: adjudication.Termination,
		MoveCount: len(moves), Moves: moves, Tags: tags, CreatedAt: time.Now().UTC(),
	}, nil
}

func coordinateTokens(content string) []string {
	// Coordinates are deliberately strict. Chinese descriptive notation is not
	// silently guessed; unsupported tokens are ignored only when they are PGN
	// metadata, move numbers, results or comments.
	fields := strings.FieldsFunc(content, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == ';'
	})
	tokens := make([]string, 0, len(fields))
	inBraceComment := false
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if strings.HasPrefix(field, "{") {
			inBraceComment = true
		}
		if inBraceComment {
			if strings.HasSuffix(field, "}") {
				inBraceComment = false
			}
			continue
		}
		field = strings.Trim(field, "()[]{}")
		if len(field) == 4 {
			if _, err := xiangqi.ParseMove(field); err == nil {
				tokens = append(tokens, strings.ToLower(field))
			}
		}
	}
	return tokens
}

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\x00", "")
	if len([]rune(name)) > 120 {
		name = string([]rune(name)[:120])
	}
	return name
}

func cloneRecord(record Record) Record {
	record.Moves = append([]Move(nil), record.Moves...)
	record.Tags = append([]string(nil), record.Tags...)
	return record
}

func cloneBatch(batch ImportBatch) ImportBatch {
	batch.RecordIDs = append([]string(nil), batch.RecordIDs...)
	batch.Errors = append([]ImportError(nil), batch.Errors...)
	return batch
}

func id() string {
	var data [16]byte
	_, _ = rand.Read(data[:])
	return hex.EncodeToString(data[:])
}
