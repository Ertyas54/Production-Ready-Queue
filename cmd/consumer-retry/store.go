package main

import (
	"sync"
	"time"
)

type RetryRecord struct {
	ID      string
	OrderID string
	Attempt int
	Error   string
	Status  string
	Time    time.Time
}

type RetryStore struct {
	mu       sync.Mutex
	messages []RetryRecord
}

func NewRetryStore() *RetryStore {
	return &RetryStore{
		messages: make([]RetryRecord, 0),
	}
}

func (s *RetryStore) Add(record RetryRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, record)
}

func (s *RetryStore) Stats() (total, success, failed, retrying int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	total = len(s.messages)
	for _, r := range s.messages {
		switch r.Status {
		case "success":
			success++
		case "failed":
			failed++
		case "retrying":
			retrying++
		}
	}
	return
}

func (s *RetryStore) Last(n int) []RetryRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.messages) == 0 {
		return nil
	}

	start := 0
	if len(s.messages) > n {
		start = len(s.messages) - n
	}

	result := make([]RetryRecord, len(s.messages)-start)
	copy(result, s.messages[start:])
	return result
}
