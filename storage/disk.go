package storage

import (
	"encoding/json"
	"os"
	"sync"
)

type DiskStore struct {
	filename string
	mu       sync.Mutex
}

func NewDiskStore(filename string) (*DiskStore, error) {
	f, err := os.OpenFile(filename, os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	f.Close()
	return &DiskStore{filename: filename}, nil
}

func (d *DiskStore) Save(chunk *AudioChunk) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	f, err := os.OpenFile(d.filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

func (d *DiskStore) GetByID(id string) (*AudioChunk, error) {
	return nil, ErrNotFound
}

func (d *DiskStore) GetByUser(userID string) ([]*AudioChunk, error) {
	return nil, ErrNotFound
}

func (d *DiskStore) GetBySession(sessionID string) ([]*AudioChunk, error) {
	return nil, ErrNotFound
}
