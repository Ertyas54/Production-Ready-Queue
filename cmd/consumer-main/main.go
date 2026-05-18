package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"production-ready-queue/internal/message"
	"production-ready-queue/internal/processor"
	"production-ready-queue/internal/queue"
)

func main() {
	log.SetFlags(log.Ltime)
	log.Println("[MAIN CONSUMER] Запуск Main Consumer сервиса")

	qm, err := queue.NewQueueManager()
	if err != nil {
		log.Fatalf("[MAIN CONSUMER] Ошибка подключения: %v", err)
	}
	defer qm.Close()

	queues, err := qm.DeclareQueues()
	if err != nil {
		log.Fatalf("[MAIN CONSUMER] Ошибка создания очередей: %v", err)
	}

	msgs, err := qm.ConsumeMessages(queues.Main)
	if err != nil {
		log.Fatalf("[MAIN CONSUMER] Ошибка подписки: %v", err)
	}

	log.Printf("[MAIN CONSUMER] Подписан на очередь: %s", queues.Main)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			log.Println("[MAIN CONSUMER] Получен сигнал завершения")
			return
		case delivery := <-msgs:
			var msg message.Message
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				log.Printf("[MAIN CONSUMER] Ошибка десериализации: %v", err)
				delivery.Ack(false)
				continue
			}

			log.Printf("[MAIN CONSUMER] Получено: %s", msg.String())

			err := processor.ProcessMessage(msg)

			if err == nil {
				log.Printf("[MAIN CONSUMER] Успешно обработан: %s", msg.ID)
				delivery.Ack(false)
				continue
			}

			msg.RetryCount++

			if processor.IsPermanentError(err) {
				log.Printf("[MAIN CONSUMER] %s: ошибка валидации → DEAD LETTER QUEUE (%v)", msg.ID, err)
				qm.PublishMessage(queues.Dead, msg)
				delivery.Ack(false)
				continue
			}

			if msg.RetryCount >= processor.MaxRetries {
				log.Printf("[MAIN CONSUMER] %s: превышен лимит попыток → DEAD LETTER QUEUE", msg.ID)
				qm.PublishMessage(queues.Dead, msg)
			} else {
				log.Printf("[MAIN CONSUMER] %s: временная ошибка → RETRY QUEUE (%v)", msg.ID, err)
				qm.PublishMessage(queues.Retry, msg)
			}
			delivery.Ack(false)
		}
	}
}
