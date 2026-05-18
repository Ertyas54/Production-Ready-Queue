package main

import (
	"fmt"
	"strings"

	"production-ready-queue/internal/processor"
)

func PrintReport(store *RetryStore) {
	total, success, failed, retrying := store.Stats()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("RETRY QUEUE - ОТЧЁТ")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Статистика ретраев:\n")
	fmt.Printf("  Всего ретраев: %d\n", total)
	fmt.Printf("  Успешно после ретрая: %d\n", success)
	fmt.Printf("  Исчерпаны попытки (→ DLQ): %d\n", failed)
	fmt.Printf("  В процессе: %d\n", retrying)
	fmt.Println(strings.Repeat("-", 60))

	last := store.Last(10)
	if len(last) > 0 && total > 10 {
		fmt.Printf("   (показаны последние %d из %d)\n", len(last), total)
	}

	for _, r := range last {
		statusText := "RETRY"
		switch r.Status {
		case "success":
			statusText = "SUCCESS"
		case "failed":
			statusText = "FAILED"
		}
		fmt.Printf("[%s] %s | Заказ: %s | Попытка: %d/%d | Ошибка: %s | %s\n",
			statusText, r.ID, r.OrderID, r.Attempt, processor.MaxRetries, r.Error, r.Time.Format("15:04:05"))
	}
	fmt.Println(strings.Repeat("=", 60))
}
