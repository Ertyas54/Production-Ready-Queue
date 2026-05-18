package processor

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"production-ready-queue/internal/message"
)

var (
	MaxRetries         int
	SimulatedErrorRate float64
)

func init() {
	MaxRetries = getEnvInt("MAX_RETRIES", 3)
	SimulatedErrorRate = getEnvFloat("SIMULATED_ERROR_RATE", 0.3)
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func ProcessMessage(msg message.Message) error {
	if msg.Order.OrderID == "RETRY-TEST" {
		return fmt.Errorf("временная ошибка сети: connection timeout (демонстрация Retry)")
	}

	if err := message.ValidateOrder(msg.Order); err != nil {
		return err
	}

	if rand.Float64() < SimulatedErrorRate {
		return fmt.Errorf("временная ошибка сети")
	}

	return nil
}

func IsPermanentError(err error) bool {
	_, ok := err.(*message.ValidationError)
	return ok
}
