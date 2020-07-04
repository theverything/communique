package store

import (
	"io"
	"sync"
)

// Store -
type Store interface {
	RegisterClient(client io.Writer)
	UnregisterClient(client io.Writer)
	JoinTopics(topics []string, client io.Writer)
	LeaveTopics(topics []string, client io.Writer)
	GetTopicClients(topic string) []io.Writer
}

type store struct {
	topic   map[string]map[io.Writer]struct{}
	clients map[io.Writer]map[string]struct{}
	mu      *sync.RWMutex
}

// RegisterClient -
func (s *store) RegisterClient(client io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check to see if the client is registered
	if _, ok := s.clients[client]; !ok {
		// add the client to the clients registry
		s.clients[client] = make(map[string]struct{})
	}
}

// UnregisterClient -
func (s *store) UnregisterClient(client io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check to see if the client is registered
	if topics, ok := s.clients[client]; ok {
		// loop through clients topics
		for topic := range topics {
			// remove client from topic
			delete(s.topic[topic], client)

			// if the topic has no registered clients delete it
			if len(s.topic[topic]) == 0 {
				delete(s.topic, topic)
			}
		}

		// remove client from clients registry
		delete(s.clients, client)
	}
}

// JoinTopics -
func (s *store) JoinTopics(topics []string, client io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// check to see if the client is registered
	if _, ok := s.clients[client]; !ok {
		// add the client to the clients registry
		s.clients[client] = make(map[string]struct{})
	}

	// loop through topics to join
	for _, topic := range topics {
		// check to see if topic is in registry, if not add it
		if _, ok := s.topic[topic]; !ok {
			s.topic[topic] = make(map[io.Writer]struct{})
		}

		// add client to topic
		s.topic[topic][client] = struct{}{}
		// add topic to client
		s.clients[client][topic] = struct{}{}
	}
}

// LeaveTopics -
func (s *store) LeaveTopics(topics []string, client io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// loop through topics to leave
	for _, topic := range topics {
		// remove client from topic
		delete(s.topic[topic], client)
		// remove topic from client
		delete(s.clients[client], topic)

		// if the topic has no registered clients delete it
		if len(s.topic[topic]) == 0 {
			delete(s.topic, topic)
		}
	}
}

// GetClientTopics -
func (s *store) GetTopicClients(topic string) []io.Writer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// check to see if topic exists
	if t, ok := s.topic[topic]; ok {
		cw := make([]io.Writer, 0, len(t))
		for c := range t {
			cw = append(cw, c)
		}

		return cw
	}

	return nil
}

// New -
func New() Store {
	return &store{
		topic:   make(map[string]map[io.Writer]struct{}),
		clients: make(map[io.Writer]map[string]struct{}),
		mu:      new(sync.RWMutex),
	}
}
