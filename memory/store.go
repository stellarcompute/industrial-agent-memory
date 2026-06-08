package memory

import (
	"context"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("memory not found")

type Store interface {
	Save(ctx context.Context, item Memory) error
	Get(ctx context.Context, id string) (Memory, error)
	List(ctx context.Context) ([]Memory, error)
	Delete(ctx context.Context, id string) error
}

type InMemoryStore struct {
	mu    sync.RWMutex
	items map[string]Memory
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{items: make(map[string]Memory)}
}

func (s *InMemoryStore) Save(_ context.Context, item Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[item.ID] = item
	return nil
}

func (s *InMemoryStore) Get(_ context.Context, id string) (Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	if !ok {
		return Memory{}, ErrNotFound
	}

	return item, nil
}

func (s *InMemoryStore) List(_ context.Context) ([]Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Memory, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}

	return items, nil
}

func (s *InMemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}

	delete(s.items, id)
	return nil
}
