package game

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Event struct {
	EventID      string    `json:"eventId"`
	MatchID      string    `json:"matchId"`
	MatchVersion int64     `json:"matchVersion"`
	Type         string    `json:"type"`
	Timestamp    time.Time `json:"timestamp"`
	Payload      any       `json:"payload"`
}

type EventBus struct {
	mu          sync.RWMutex
	history     map[string][]Event
	subscribers map[string]map[chan Event]struct{}
}

func NewEventBus() *EventBus {
	return &EventBus{
		history: make(map[string][]Event), subscribers: make(map[string]map[chan Event]struct{}),
	}
}

func (b *EventBus) Publish(matchID string, version int64, kind string, payload any) Event {
	event := Event{
		EventID: newID(), MatchID: matchID, MatchVersion: version,
		Type: kind, Timestamp: time.Now().UTC(), Payload: payload,
	}
	b.mu.Lock()
	history := append(b.history[matchID], event)
	if len(history) > 128 {
		history = history[len(history)-128:]
	}
	b.history[matchID] = history
	subscribers := make([]chan Event, 0, len(b.subscribers[matchID]))
	for subscriber := range b.subscribers[matchID] {
		subscribers = append(subscribers, subscriber)
	}
	b.mu.Unlock()
	for _, subscriber := range subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
	return event
}

func (b *EventBus) Subscribe(matchID, afterEventID string) (<-chan Event, []Event, func()) {
	channel := make(chan Event, 16)
	b.mu.Lock()
	if b.subscribers[matchID] == nil {
		b.subscribers[matchID] = make(map[chan Event]struct{})
	}
	b.subscribers[matchID][channel] = struct{}{}
	history := append([]Event(nil), b.history[matchID]...)
	b.mu.Unlock()

	if afterEventID != "" {
		index := -1
		for i := range history {
			if history[i].EventID == afterEventID {
				index = i
			}
		}
		if index >= 0 {
			history = history[index+1:]
		}
	}
	cancel := func() {
		b.mu.Lock()
		if subscribers := b.subscribers[matchID]; subscribers != nil {
			delete(subscribers, channel)
		}
		b.mu.Unlock()
	}
	return channel, history, cancel
}

func newID() string {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(data[:])
}
