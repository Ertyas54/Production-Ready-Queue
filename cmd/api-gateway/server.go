package main

import (
	"sync"

	"production-ready-queue/internal/queue"
)

type APIServer struct {
	qm       *queue.QueueManager
	queues   *queue.QueueNames
	msgCount int
	mu       sync.Mutex
}

func NewAPIServer(qm *queue.QueueManager, queues *queue.QueueNames) *APIServer {
	return &APIServer{
		qm:     qm,
		queues: queues,
	}
}
