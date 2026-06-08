package memory

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type JSONLStore struct {
	path string
	mu   sync.Mutex
}

func NewJSONLStore(path string) *JSONLStore {
	return &JSONLStore{path: path}
}

func (s *JSONLStore) Save(ctx context.Context, item Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.readAll(ctx)
	if err != nil {
		return err
	}

	replaced := false
	for idx := range items {
		if items[idx].ID == item.ID {
			items[idx] = item
			replaced = true
			break
		}
	}
	if !replaced {
		items = append(items, item)
	}

	return s.writeAll(items)
}

func (s *JSONLStore) Get(ctx context.Context, id string) (Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.readAll(ctx)
	if err != nil {
		return Memory{}, err
	}

	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}

	return Memory{}, ErrNotFound
}

func (s *JSONLStore) List(ctx context.Context) ([]Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.readAll(ctx)
}

func (s *JSONLStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.readAll(ctx)
	if err != nil {
		return err
	}

	filtered := items[:0]
	found := false
	for _, item := range items {
		if item.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, item)
	}
	if !found {
		return ErrNotFound
	}

	return s.writeAll(filtered)
}

func (s *JSONLStore) readAll(ctx context.Context) ([]Memory, error) {
	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Memory{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	items := make([]Memory, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var item Memory
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *JSONLStore) writeAll(items []Memory) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	tmpPath := s.path + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	for _, item := range items {
		if err := encoder.Encode(item); err != nil {
			_ = file.Close()
			return err
		}
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.path)
}
