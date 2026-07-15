package analysis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"xiangqi-lab/internal/domain/xiangqi"
	"xiangqi-lab/internal/engine"
	"xiangqi-lab/internal/game"
)

var ErrNotFound = errors.New("analysis resource not found")

type MatchSource interface {
	Get(string) (game.Snapshot, error)
}

type CreateJobRequest struct {
	MatchID string `json:"matchId"`
}

type Job struct {
	ID          string    `json:"id"`
	MatchID     string    `json:"matchId"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	Analyzed    int       `json:"analyzedMoves"`
	Total       int       `json:"totalMoves"`
	ErrorCode   string    `json:"errorCode,omitempty"`
	Message     string    `json:"message,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt time.Time `json:"completedAt,omitempty"`
}

type MoveAnalysis struct {
	Ply            int    `json:"ply"`
	ActualMove     string `json:"actualMove"`
	BestMove       string `json:"bestMove"`
	Side           string `json:"side"`
	Classification string `json:"classification"`
	ScoreLossCP    *int   `json:"scoreLossCp,omitempty"`
	Depth          int    `json:"depth"`
	Nodes          uint64 `json:"nodes"`
}

type Result struct {
	MatchID       string         `json:"matchId"`
	Engine        string         `json:"engine"`
	Status        string         `json:"status"`
	AnalyzedMoves int            `json:"analyzedMoves"`
	BestMoveRate  float64        `json:"bestMoveRate"`
	Moves         []MoveAnalysis `json:"moves"`
	GeneratedAt   time.Time      `json:"generatedAt"`
}

type Service struct {
	mu      sync.RWMutex
	matches MatchSource
	engine  engine.Engine
	jobs    map[string]Job
	results map[string]Result
}

func NewService(matches MatchSource, analysisEngine engine.Engine) *Service {
	return &Service{
		matches: matches, engine: analysisEngine,
		jobs: make(map[string]Job), results: make(map[string]Result),
	}
}

func (s *Service) CreateJob(request CreateJobRequest) (Job, error) {
	match, err := s.matches.Get(request.MatchID)
	if err != nil {
		return Job{}, err
	}
	job := Job{
		ID: id(), MatchID: match.ID, Status: "queued", Total: len(match.Moves),
		CreatedAt: time.Now().UTC(),
	}
	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()
	go s.analyze(job.ID, match)
	return job, nil
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

func (s *Service) GetResult(matchID string) (Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, ok := s.results[matchID]
	if !ok {
		return Result{}, ErrNotFound
	}
	result.Moves = append([]MoveAnalysis(nil), result.Moves...)
	return result, nil
}

func (s *Service) analyze(jobID string, match game.Snapshot) {
	s.updateJob(jobID, func(job *Job) { job.Status = "running" })
	items := make([]MoveAnalysis, 0, len(match.Moves))
	bestCount := 0
	for index, moveRecord := range match.Moves {
		position, err := xiangqi.ParseFEN(moveRecord.FENBefore)
		if err != nil {
			s.fail(jobID, "INVALID_MATCH_HISTORY", "对局历史中的 FEN 无法解析")
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		result, err := s.engine.Analyze(ctx, engine.AnalyzeRequest{
			Position: position, MaxDepth: 2, MaxNodes: 8_000,
			MoveTime: 200 * time.Millisecond, MultiPV: 3,
		})
		cancel()
		if err != nil {
			s.fail(jobID, "ENGINE_FAILED", "复盘引擎未能完成分析")
			return
		}
		item := MoveAnalysis{
			Ply: index + 1, ActualMove: moveRecord.Move, BestMove: result.BestMoveICCS,
			Side: moveRecord.Side, Classification: "outside_top_candidates",
			Depth: result.Depth, Nodes: result.Nodes,
		}
		if item.ActualMove == item.BestMove {
			item.Classification = "best"
			zero := 0
			item.ScoreLossCP = &zero
			bestCount++
		} else if len(result.Candidates) > 0 {
			bestScore := result.Candidates[0].ScoreCP
			for _, candidate := range result.Candidates {
				if candidate.MoveICCS != item.ActualMove {
					continue
				}
				loss := bestScore - candidate.ScoreCP
				if moveRecord.Side == "black" {
					loss = candidate.ScoreCP - bestScore
				}
				if loss < 0 {
					loss = 0
				}
				item.ScoreLossCP = &loss
				switch {
				case loss <= 30:
					item.Classification = "excellent"
				case loss <= 100:
					item.Classification = "inaccuracy"
				case loss <= 250:
					item.Classification = "mistake"
				default:
					item.Classification = "blunder"
				}
				break
			}
		}
		items = append(items, item)
		progress := 100
		if len(match.Moves) > 0 {
			progress = (index + 1) * 100 / len(match.Moves)
		}
		s.updateJob(jobID, func(job *Job) {
			job.Progress, job.Analyzed = progress, index+1
		})
	}
	rate := 0.0
	if len(items) > 0 {
		rate = float64(bestCount) / float64(len(items))
	}
	generated := time.Now().UTC()
	s.mu.Lock()
	s.results[match.ID] = Result{
		MatchID: match.ID, Engine: s.engine.Name(), Status: "completed",
		AnalyzedMoves: len(items), BestMoveRate: rate, Moves: items, GeneratedAt: generated,
	}
	job := s.jobs[jobID]
	job.Status, job.Progress, job.Analyzed, job.CompletedAt = "completed", 100, len(items), generated
	s.jobs[jobID] = job
	s.mu.Unlock()
}

func (s *Service) fail(jobID, code, message string) {
	s.updateJob(jobID, func(job *Job) {
		job.Status, job.ErrorCode, job.Message = "failed", code, message
		job.CompletedAt = time.Now().UTC()
	})
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
