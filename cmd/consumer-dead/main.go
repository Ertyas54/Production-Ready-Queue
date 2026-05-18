package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"production-ready-queue/internal/message"
	"production-ready-queue/internal/queue"
)

type DeadLetterStore struct {
	mu       sync.Mutex
	messages []message.Message
}

func (s *DeadLetterStore) Add(msg message.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, msg)
}

func (s *DeadLetterStore) Report() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("DEAD LETTER QUEUE - ОТЧЕТ")
	log.Println(strings.Repeat("=", 60))
	log.Printf("------------ Всего мертвых сообщений ---------------: %d\n", len(s.messages))

	for i, msg := range s.messages {
		log.Printf("%d. %s | Тело: %s | Создано: %s | Попыток: %d",
			i+1, msg.ID, msg.Body, msg.CreatedAt.Format("15:04:05"), msg.RetryCount)
	}
}

func main() {
	log.SetFlags(log.Ltime)
	log.Println("[DEAD CONSUMER] Запуск Dead Letter Consumer сервиса...")

	qm, err := queue.NewQueueManager()
	if err != nil {
		log.Fatalf("[DEAD CONSUMER] Ошибка подключения: %v", err)
	}
	defer qm.Close()

	queues, err := qm.DeclareQueues()
	if err != nil {
		log.Fatalf("[DEAD CONSUMER] Ошибка создания очередей: %v", err)
	}

	msgs, err := qm.ConsumeMessages(queues.Dead)
	if err != nil {
		log.Fatalf("[DEAD CONSUMER] Ошибка подписки: %v", err)
	}

	log.Printf("[DEAD CONSUMER] Подписан на очередь: %s", queues.Dead)

	store := &DeadLetterStore{}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			store.Report()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			log.Println("[DEAD CONSUMER] Получен сигнал завершения")
			store.Report()
			return
		case delivery := <-msgs:
			var msg message.Message
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				log.Printf("[DEAD CONSUMER] Ошибка десериализации: %v", err)
				delivery.Nack(false, false)
				continue
			}

			store.Add(msg)
			log.Printf("[DEAD CONSUMER] МЕРТВОЕ СООБЩЕНИЕ: %s | Требуется ручное вмешательство!", msg.ID)
			delivery.Ack(false)
		}
	}
}
