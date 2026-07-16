package game

import (
	"sync"
)

// Repository defines the authoritative persistence boundary for matches.
// Application services depend on this contract rather than a concrete storage
// implementation so memory and durable adapters can be selected at startup.
type Repository interface {
	Name() string
	Create(match Match) error
	Get(id string) (Match, error)
	List() []Match
	Update(id string, expectedVersion int64, update func(*Match) error) (Match, error)
	Idempotency(route, key, digest string) (Match, bool, error)
	SaveIdempotency(route, key, digest string, match Match)
}

type MemoryRepository struct {
	mu          sync.RWMutex
	matches     map[string]*Match
	idempotency map[string]idempotentResult
}

var _ Repository = (*MemoryRepository)(nil)

type idempotentResult struct {
	Digest string
	Match  Match
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		matches: make(map[string]*Match), idempotency: make(map[string]idempotentResult),
	}
}

func (r *MemoryRepository) Name() string { return "memory" }

func (r *MemoryRepository) Create(match Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.matches[match.ID]; exists {
		return ErrStateConflict
	}
	copy := cloneMatch(match)
	r.matches[match.ID] = &copy
	return nil
}

func (r *MemoryRepository) Get(id string) (Match, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	match, ok := r.matches[id]
	if !ok {
		return Match{}, ErrNotFound
	}
	return cloneMatch(*match), nil
}

func (r *MemoryRepository) List() []Match {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]Match, 0, len(r.matches))
	for _, match := range r.matches {
		items = append(items, cloneMatch(*match))
	}
	return items
}

func (r *MemoryRepository) Update(id string, expectedVersion int64, update func(*Match) error) (Match, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	match, ok := r.matches[id]
	if !ok {
		return Match{}, ErrNotFound
	}
	if expectedVersion >= 0 && match.Version != expectedVersion {
		return Match{}, ErrVersionConflict
	}
	working := cloneMatch(*match)
	if err := update(&working); err != nil {
		return Match{}, err
	}
	r.matches[id] = &working
	return cloneMatch(working), nil
}

func (r *MemoryRepository) Idempotency(route, key, digest string) (Match, bool, error) {
	if key == "" {
		return Match{}, false, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.idempotency[route+"|"+key]
	if !ok {
		return Match{}, false, nil
	}
	if value.Digest != digest {
		return Match{}, false, ErrIdempotency
	}
	return cloneMatch(value.Match), true, nil
}

func (r *MemoryRepository) SaveIdempotency(route, key, digest string, match Match) {
	if key == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.idempotency[route+"|"+key] = idempotentResult{Digest: digest, Match: cloneMatch(match)}
}

func cloneMatch(match Match) Match {
	match.Moves = append([]MoveRecord(nil), match.Moves...)
	return match
}
