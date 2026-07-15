package game

import "errors"

var (
	ErrNotFound        = errors.New("match not found")
	ErrVersionConflict = errors.New("match version conflict")
	ErrStateConflict   = errors.New("match state conflict")
	ErrNotPlayerTurn   = errors.New("not the player's turn")
	ErrUndoDisabled    = errors.New("undo is disabled")
	ErrNoMovesToUndo   = errors.New("no moves to undo")
	ErrIdempotency     = errors.New("idempotency key was reused with different input")
)
