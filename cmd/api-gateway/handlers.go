package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"production-ready-queue/internal/message"
)

func (s *APIServer) handleValidOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	s.msgCount++
	count := s.msgCount
	s.mu.Unlock()

	msg := message.NewValidOrder(count)
	msg.Timestamp = time.Now().UnixNano()
	s.qm.PublishMessage(s.queues.Main, msg)

	respondJSON(w, map[string]interface{}{
		"message": fmt.Sprintf("Корректный заказ %s отправлен. Обработан успешно.", msg.Order.OrderID),
	})
}

func (s *APIServer) handleRetryOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	s.msgCount++
	count := s.msgCount
	s.mu.Unlock()

	msg := message.NewValidOrder(count)
	msg.Order.OrderID = "RETRY-TEST"
	msg.Body = "Заказ с сетевой ошибкой (демонстрация Retry)"
	msg.Timestamp = time.Now().UnixNano()
	s.qm.PublishMessage(s.queues.Main, msg)

	respondJSON(w, map[string]interface{}{
		"message": fmt.Sprintf("%s отправлен. Пройдёт через Retry → Dead Letter Queue.", msg.ID),
	})
}

func (s *APIServer) handleInvalidOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	s.msgCount++
	count := s.msgCount
	s.mu.Unlock()

	msg := message.NewInvalidOrder(count)
	msg.Timestamp = time.Now().UnixNano()
	s.qm.PublishMessage(s.queues.Main, msg)

	respondJSON(w, map[string]interface{}{
		"message": fmt.Sprintf("%s отправлен. Ошибка валидации → Dead Letter Queue.", msg.ID),
	})
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
