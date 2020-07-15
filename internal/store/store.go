package store

import (
	"io"
	"sync"
)

// Getter -
type Getter interface {
	Get(key string) map[io.Writer]struct{}
}

// Setter -
type Setter interface {
	Set(key string, client io.Writer) int
}

// Remover -
type Remover interface {
	Remove(key string, client io.Writer) int
}

// Store -
type Store interface {
	Getter
	Setter
	Remover
}

type store struct {
	topics map[string]map[io.Writer]struct{}
	mu     *sync.RWMutex
}

// Set -
func (s *store) Set(key string, client io.Writer) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.topics[key]; !ok {
		s.topics[key] = make(map[io.Writer]struct{})
	}

	s.topics[key][client] = struct{}{}

	return len(s.topics[key])
}

// Get -
func (s *store) Get(key string) map[io.Writer]struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if clients, ok := s.topics[key]; ok {
		return clients
	}

	return nil
}

// Remove -
func (s *store) Remove(key string, client io.Writer) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.topics[key]; ok {
		delete(s.topics[key], client)

		if len(s.topics[key]) == 0 {
			delete(s.topics, key)
		}
	}

	return len(s.topics[key])
}

// New -
func New() Store {
	return &store{
		topics: make(map[string]map[io.Writer]struct{}),
		mu:     new(sync.RWMutex),
	}
}
