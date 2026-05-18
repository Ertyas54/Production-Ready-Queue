package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/streadway/amqp"
	"production-ready-queue/internal/message"
	"production-ready-queue/internal/processor"
	"production-ready-queue/internal/queue"
)

func main() {
	log.SetFlags(log.Ltime)
	log.Println("[RETRY CONSUMER] Запуск Retry Consumer сервиса")

	qm, err := queue.NewQueueManager()
	if err != nil {
		log.Fatalf("[RETRY CONSUMER] Ошибка подключения: %v", err)
	}
	defer qm.Close()

	queues, err := qm.DeclareQueues()
	if err != nil {
		log.Fatalf("[RETRY CONSUMER] Ошибка создания очередей: %v", err)
	}

	msgs, err := qm.ConsumeMessages(queues.Retry)
	if err != nil {
		log.Fatalf("[RETRY CONSUMER] Ошибка подписки: %v", err)
	}

	log.Printf("[RETRY CONSUMER] Подписан на очередь: %s", queues.Retry)

	store := NewRetryStore()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			PrintReport(store)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			log.Println("[RETRY CONSUMER] Получен сигнал завершения")
			PrintReport(store)
			return
		case delivery := <-msgs:
			processDelivery(qm, queues, store, delivery)
		}
	}
}

func processDelivery(qm *queue.QueueManager, queues *queue.QueueNames, store *RetryStore, delivery amqp.Delivery) {
	var msg message.Message
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		delivery.Ack(false)
		return
	}

	attempt := msg.RetryCount + 1
	log.Printf("[RETRY] Обработка %s (попытка %d/%d)", msg.ID, attempt, processor.MaxRetries)

	time.Sleep(500 * time.Millisecond)

	err := processor.ProcessMessage(msg)

	if err == nil {
		log.Printf("[RETRY] %s успешно обработан после ретрая!", msg.ID)
		store.Add(RetryRecord{
			ID: msg.ID, OrderID: msg.Order.OrderID,
			Attempt: attempt, Error: "-", Status: "success", Time: time.Now(),
		})
		delivery.Ack(false)
		return
	}

	msg.RetryCount++

	if processor.IsPermanentError(err) || msg.RetryCount >= processor.MaxRetries {
		reason := "исчерпаны попытки"
		if processor.IsPermanentError(err) {
			reason = "ошибка валидации"
		}
		log.Printf("[RETRY] %s → DEAD LETTER QUEUE (%s)", msg.ID, reason)
		store.Add(RetryRecord{
			ID: msg.ID, OrderID: msg.Order.OrderID,
			Attempt: attempt, Error: err.Error(), Status: "failed", Time: time.Now(),
		})
		qm.PublishMessage(queues.Dead, msg)
	} else {
		log.Printf("[RETRY] %s ошибка, ещё ретрай (%d/%d): %v", msg.ID, attempt, processor.MaxRetries, err)
		store.Add(RetryRecord{
			ID: msg.ID, OrderID: msg.Order.OrderID,
			Attempt: attempt, Error: err.Error(), Status: "retrying", Time: time.Now(),
		})
		qm.PublishMessage(queues.Retry, msg)
	}
	delivery.Ack(false)
}
