package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"production-ready-queue/internal/queue"
)

func main() {
	log.SetFlags(log.Ltime)
	log.Println("[PRODUCER] Producer сервис запущен (режим ожидания)")
	log.Println("[PRODUCER] Сообщения отправляются только через веб-интерфейс (порт 8082)")

	qm, err := queue.NewQueueManager()
	if err != nil {
		log.Fatalf("[PRODUCER] Ошибка подключения: %v", err)
	}
	defer qm.Close()

	_, err = qm.DeclareQueues()
	if err != nil {
		log.Fatalf("[PRODUCER] Ошибка создания очередей: %v", err)
	}

	log.Printf("[PRODUCER] Подключен к RabbitMQ")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[PRODUCER] Завершение работы")
}
