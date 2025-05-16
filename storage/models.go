package storage

import "errors"

type AudioChunk struct {
	ID        string
	UserID    string
	SessionID string
	Timestamp int64
	Data      []byte
	Metadata  map[string]string
}

var ErrNotFound = errors.New("chunk not found")

type Store interface {
	Save(chunk *AudioChunk) error
	GetByID(id string) (*AudioChunk, error)
	GetByUser(userID string) ([]*AudioChunk, error)
	GetBySession(sessionID string) ([]*AudioChunk, error)
}
