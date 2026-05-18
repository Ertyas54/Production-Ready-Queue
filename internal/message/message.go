package message

import (
	"fmt"
	"time"
)

type Order struct {
	OrderID   string  `json:"order_id"`
	UserID    int     `json:"user_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Email     string  `json:"email"`
}

type Message struct {
	ID         string    `json:"id"`
	Body       string    `json:"body"`
	Order      Order     `json:"order"`
	RetryCount int       `json:"retry_count"`
	CreatedAt  time.Time `json:"created_at"`
	Timestamp  int64     `json:"timestamp"`
}

func NewValidOrder(id int) Message {
	return Message{
		ID:        fmt.Sprintf("ORDER-%03d", id),
		Body:      "Новый заказ",
		CreatedAt: time.Now(),
		Timestamp: time.Now().UnixNano(),
		Order: Order{
			OrderID:   fmt.Sprintf("ORD-%d", 1000+id),
			UserID:    42,
			ProductID: 777,
			Quantity:  2,
			Price:     1999.99,
			Email:     "customer@example.com",
		},
	}
}

func NewInvalidOrder(id int) Message {
	errors := []Order{
		{OrderID: "", UserID: 42, ProductID: 777, Quantity: 2, Price: 100, Email: "ok@email.com"},
		{OrderID: "ORD-999", UserID: -1, ProductID: 777, Quantity: 2, Price: 100, Email: "ok@email.com"},
		{OrderID: "ORD-999", UserID: 42, ProductID: 0, Quantity: 2, Price: 100, Email: "ok@email.com"},
		{OrderID: "ORD-999", UserID: 42, ProductID: 777, Quantity: -5, Price: 100, Email: "ok@email.com"},
		{OrderID: "ORD-999", UserID: 42, ProductID: 777, Quantity: 2, Price: -100, Email: "ok@email.com"},
		{OrderID: "ORD-999", UserID: 42, ProductID: 777, Quantity: 2, Price: 100, Email: "not-an-email"},
		{OrderID: "ORD-999", UserID: 42, ProductID: 777, Quantity: 0, Price: 100, Email: "ok@email.com"},
	}

	invalidOrder := errors[id%len(errors)]

	return Message{
		ID:        fmt.Sprintf("ORDER-%03d", id),
		Body:      fmt.Sprintf("Заказ с ошибкой валидации (тип ошибки #%d)", id%len(errors)+1),
		CreatedAt: time.Now(),
		Timestamp: time.Now().UnixNano(),
		Order:     invalidOrder,
	}
}

func (m Message) String() string {
	return fmt.Sprintf("[%s] %s (retries=%d)", m.ID, m.Body, m.RetryCount)
}
