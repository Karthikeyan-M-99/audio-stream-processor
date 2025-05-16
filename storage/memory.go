package storage

import (
	"sync"
)

type MemoryStore struct {
	mu         sync.RWMutex
	chunks     map[string]*AudioChunk
	userIdx    map[string]map[string]struct{}
	sessionIdx map[string]map[string]struct{}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		chunks:     make(map[string]*AudioChunk),
		userIdx:    make(map[string]map[string]struct{}),
		sessionIdx: make(map[string]map[string]struct{}),
	}
}

func (m *MemoryStore) Save(chunk *AudioChunk) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.chunks[chunk.ID] = chunk

	if _, ok := m.userIdx[chunk.UserID]; !ok {
		m.userIdx[chunk.UserID] = make(map[string]struct{})
	}
	m.userIdx[chunk.UserID][chunk.ID] = struct{}{}

	if _, ok := m.sessionIdx[chunk.SessionID]; !ok {
		m.sessionIdx[chunk.SessionID] = make(map[string]struct{})
	}
	m.sessionIdx[chunk.SessionID][chunk.ID] = struct{}{}

	return nil
}

func (m *MemoryStore) GetByID(id string) (*AudioChunk, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chunk, ok := m.chunks[id]
	if !ok {
		return nil, ErrNotFound
	}
	return chunk, nil
}

func (m *MemoryStore) GetByUser(userID string) ([]*AudioChunk, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids, ok := m.userIdx[userID]
	if !ok {
		return []*AudioChunk{}, nil
	}

	result := make([]*AudioChunk, 0, len(ids))
	for id := range ids {
		if chunk, ok := m.chunks[id]; ok {
			result = append(result, chunk)
		}
	}
	return result, nil
}

func (m *MemoryStore) GetBySession(sessionID string) ([]*AudioChunk, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids, ok := m.sessionIdx[sessionID]
	if !ok {
		return []*AudioChunk{}, nil
	}

	result := make([]*AudioChunk, 0, len(ids))
	for id := range ids {
		if chunk, ok := m.chunks[id]; ok {
			result = append(result, chunk)
		}
	}
	return result, nil
}
