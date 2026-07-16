package learning

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/records"
)

var (
	ErrNotFound        = errors.New("learning resource not found")
	ErrVersionNotReady = errors.New("learning version is not ready")
)

type RecordSource interface {
	AllWithMoves() []records.Record
}

type CreateJobRequest struct {
	Name      string   `json:"name"`
	RecordIDs []string `json:"recordIds,omitempty"`
}

type Job struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Name        string    `json:"name"`
	Progress    int       `json:"progress"`
	RecordCount int       `json:"recordCount"`
	MoveCount   int       `json:"moveCount"`
	VersionID   string    `json:"versionId,omitempty"`
	ErrorCode   string    `json:"errorCode,omitempty"`
	Message     string    `json:"message,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt time.Time `json:"completedAt,omitempty"`
}

type BookEntry struct {
	PositionHash string `json:"positionHash"`
	FEN          string `json:"fen"`
	SideToMove   string `json:"sideToMove"`
	Move         string `json:"move"`
	Samples      int    `json:"samples"`
	RedWins      int    `json:"redWins"`
	BlackWins    int    `json:"blackWins"`
	Draws        int    `json:"draws"`
}

type QualityReport struct {
	ValidRecords     int `json:"validRecords"`
	ValidMoves       int `json:"validMoves"`
	CoveredPositions int `json:"coveredPositions"`
	LowSampleEntries int `json:"lowSampleEntries"`
}

type Version struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Status      string        `json:"status"`
	Algorithm   string        `json:"algorithm"`
	Quality     QualityReport `json:"quality"`
	EntryCount  int           `json:"entryCount"`
	Entries     []BookEntry   `json:"entries,omitempty"`
	CreatedAt   time.Time     `json:"createdAt"`
	ActivatedAt time.Time     `json:"activatedAt,omitempty"`
}

type Service struct {
	mu       sync.RWMutex
	source   RecordSource
	jobs     map[string]Job
	versions map[string]Version
	activeID string
}

func NewService(source RecordSource) *Service {
	return &Service{source: source, jobs: make(map[string]Job), versions: make(map[string]Version)}
}

func (s *Service) CreateJob(request CreateJobRequest) Job {
	job := Job{
		ID: id(), Status: "queued", Name: request.Name, Progress: 0,
		CreatedAt: time.Now().UTC(),
	}
	if job.Name == "" {
		job.Name = "棋谱学习版本"
	}
	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()
	go s.build(job.ID, request)
	return job
}

func (s *Service) GetJob(jobID string) (Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[jobID]
	if !ok {
		return Job{}, ErrNotFound
	}
	return job, nil
}

func (s *Service) ListVersions() []Version {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Version, 0, len(s.versions))
	for _, version := range s.versions {
		copy := version
		copy.Entries = nil
		items = append(items, copy)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items
}

func (s *Service) GetVersion(versionID string) (Version, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	version, ok := s.versions[versionID]
	if !ok {
		return Version{}, ErrNotFound
	}
	version.Entries = append([]BookEntry(nil), version.Entries...)
	return version, nil
}

func (s *Service) Activate(versionID string) (Version, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	version, ok := s.versions[versionID]
	if !ok {
		return Version{}, ErrNotFound
	}
	if version.Status != "ready" && version.Status != "superseded" && version.Status != "active" {
		return Version{}, ErrVersionNotReady
	}
	if current, ok := s.versions[s.activeID]; ok && current.ID != versionID {
		current.Status = "superseded"
		s.versions[current.ID] = current
	}
	version.Status = "active"
	version.ActivatedAt = time.Now().UTC()
	s.versions[versionID] = version
	s.activeID = versionID
	return version, nil
}

func (s *Service) Rollback(versionID string) (Version, error) {
	return s.Activate(versionID)
}

func (s *Service) SelectBookMove(position xiangqi.Position, mode string) (xiangqi.Move, bool) {
	if mode != "library" && mode != "style" {
		return xiangqi.Move{}, false
	}
	hash := fmt.Sprintf("%016x", position.Hash())
	side := position.SideToMove()

	s.mu.RLock()
	version, ok := s.versions[s.activeID]
	s.mu.RUnlock()
	if !ok || version.Status != "active" {
		return xiangqi.Move{}, false
	}

	type candidate struct {
		entry BookEntry
		score int
	}
	candidates := make([]candidate, 0)
	for _, entry := range version.Entries {
		if entry.PositionHash != hash || entry.SideToMove != side.String() {
			continue
		}
		move, err := xiangqi.ParseMove(entry.Move)
		if err != nil || !position.IsLegal(move) {
			continue
		}
		candidates = append(candidates, candidate{entry: entry, score: bookScore(entry, side, mode)})
	}
	if len(candidates) == 0 {
		return xiangqi.Move{}, false
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].entry.Move < candidates[j].entry.Move
		}
		return candidates[i].score > candidates[j].score
	})
	move, err := xiangqi.ParseMove(candidates[0].entry.Move)
	if err != nil {
		return xiangqi.Move{}, false
	}
	return move, true
}

func (s *Service) build(jobID string, request CreateJobRequest) {
	s.updateJob(jobID, func(job *Job) {
		job.Status, job.Progress = "running", 5
	})
	all := s.source.AllWithMoves()
	selected := make(map[string]struct{}, len(request.RecordIDs))
	for _, recordID := range request.RecordIDs {
		selected[recordID] = struct{}{}
	}
	input := make([]records.Record, 0, len(all))
	for _, record := range all {
		if len(selected) == 0 {
			input = append(input, record)
		} else if _, ok := selected[record.ID]; ok {
			input = append(input, record)
		}
	}
	if len(input) == 0 {
		s.updateJob(jobID, func(job *Job) {
			job.Status, job.Progress, job.ErrorCode = "failed", 100, "NO_RECORDS"
			job.Message = "没有可用于构建学习版本的有效棋谱"
			job.CompletedAt = time.Now().UTC()
		})
		return
	}
	type key struct{ hash, move string }
	entries := make(map[key]*BookEntry)
	positions := make(map[string]struct{})
	moveCount := 0
	for _, record := range input {
		for _, move := range record.Moves {
			position, err := xiangqi.ParseFEN(move.FENBefore)
			if err != nil {
				continue
			}
			hash := fmt.Sprintf("%016x", position.Hash())
			itemKey := key{hash: hash, move: move.Move}
			entry := entries[itemKey]
			if entry == nil {
				entry = &BookEntry{
					PositionHash: hash, FEN: position.FEN(),
					SideToMove: position.SideToMove().String(), Move: move.Move,
				}
				entries[itemKey] = entry
			}
			entry.Samples++
			switch record.Result {
			case "1-0", "red_win":
				entry.RedWins++
			case "0-1", "black_win":
				entry.BlackWins++
			case "1/2-1/2", "draw":
				entry.Draws++
			}
			positions[hash] = struct{}{}
			moveCount++
		}
	}
	book := make([]BookEntry, 0, len(entries))
	lowSample := 0
	for _, entry := range entries {
		book = append(book, *entry)
		if entry.Samples < 3 {
			lowSample++
		}
	}
	sort.Slice(book, func(i, j int) bool {
		if book[i].PositionHash == book[j].PositionHash {
			return book[i].Move < book[j].Move
		}
		return book[i].PositionHash < book[j].PositionHash
	})
	version := Version{
		ID: id(), Name: request.Name, Status: "ready", Algorithm: "position-book-count-v1",
		Quality: QualityReport{
			ValidRecords: len(input), ValidMoves: moveCount,
			CoveredPositions: len(positions), LowSampleEntries: lowSample,
		},
		EntryCount: len(book), Entries: book, CreatedAt: time.Now().UTC(),
	}
	if version.Name == "" {
		version.Name = "学习版本 " + version.CreatedAt.Format("2006-01-02 15:04")
	}
	s.mu.Lock()
	s.versions[version.ID] = version
	job := s.jobs[jobID]
	job.Status, job.Progress = "completed", 100
	job.RecordCount, job.MoveCount, job.VersionID = len(input), moveCount, version.ID
	job.CompletedAt = time.Now().UTC()
	s.jobs[jobID] = job
	s.mu.Unlock()
}

func bookScore(entry BookEntry, side xiangqi.Color, mode string) int {
	if mode == "style" {
		return entry.Samples*100 + entry.Draws
	}
	score := entry.Samples * 10
	if side == xiangqi.Red {
		score += entry.RedWins*100 + entry.Draws*20 - entry.BlackWins*100
	} else {
		score += entry.BlackWins*100 + entry.Draws*20 - entry.RedWins*100
	}
	return score
}

func (s *Service) updateJob(jobID string, update func(*Job)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job := s.jobs[jobID]
	update(&job)
	s.jobs[jobID] = job
}

func id() string {
	var data [16]byte
	_, _ = rand.Read(data[:])
	return hex.EncodeToString(data[:])
}
