package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
	"production-ready-queue/internal/message"
)

type QueueManager struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type QueueNames struct {
	Main  string
	Retry string
	Dead  string
}

func NewQueueManager() (*QueueManager, error) {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &QueueManager{conn: conn, channel: channel}, nil
}

func (qm *QueueManager) DeclareQueues() (*QueueNames, error) {
	queues := &QueueNames{
		Main:  getEnv("QUEUE_MAIN", "queue.main"),
		Retry: getEnv("QUEUE_RETRY", "queue.retry"),
		Dead:  getEnv("QUEUE_DEAD", "queue.dead"),
	}

	for _, qName := range []string{queues.Main, queues.Retry, queues.Dead} {
		_, err := qm.channel.QueueDeclare(qName, true, false, false, false, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to declare queue %s: %w", qName, err)
		}
	}

	log.Println("Все очереди созданы")
	return queues, nil
}

func (qm *QueueManager) PublishMessage(queueName string, msg message.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = qm.channel.Publish("", queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (qm *QueueManager) ConsumeMessages(queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := qm.channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}
	return msgs, nil
}

func (qm *QueueManager) Close() {
	if qm.channel != nil {
		qm.channel.Close()
	}
	if qm.conn != nil {
		qm.conn.Close()
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
