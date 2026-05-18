package message

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("валидация не пройдена: %s - %s", e.Field, e.Message)
}

func ValidateOrder(order Order) error {
	if order.OrderID == "" {
		return &ValidationError{
			Field:   "OrderID",
			Message: "не может быть пустым",
		}
	}

	if order.UserID <= 0 {
		return &ValidationError{
			Field:   "UserID",
			Message: fmt.Sprintf("должен быть положительным, получено: %d", order.UserID),
		}
	}

	if order.ProductID <= 0 {
		return &ValidationError{
			Field:   "ProductID",
			Message: fmt.Sprintf("должен быть положительным, получено: %d", order.ProductID),
		}
	}

	if order.Quantity <= 0 {
		return &ValidationError{
			Field:   "Quantity",
			Message: fmt.Sprintf("должно быть больше нуля, получено: %d", order.Quantity),
		}
	}

	if order.Price <= 0 {
		return &ValidationError{
			Field:   "Price",
			Message: fmt.Sprintf("должна быть положительной, получено: %.2f", order.Price),
		}
	}

	if !strings.Contains(order.Email, "@") {
		return &ValidationError{
			Field:   "Email",
			Message: fmt.Sprintf("некорректный формат: %s", order.Email),
		}
	}

	return nil
}

func IsPermanentError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}
