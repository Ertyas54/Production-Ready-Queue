package main

import (
	"log"
	"net/http"
	"os"

	"production-ready-queue/internal/queue"
)

func main() {
	log.SetFlags(log.Ltime)
	log.Println("🌐 [API GATEWAY] Запуск...")

	qm, err := queue.NewQueueManager()
	if err != nil {
		log.Fatalf("[API GATEWAY] Ошибка подключения: %v", err)
	}
	defer qm.Close()

	queues, err := qm.DeclareQueues()
	if err != nil {
		log.Fatalf("[API GATEWAY] Ошибка создания очередей: %v", err)
	}

	server := NewAPIServer(qm, queues)

	http.HandleFunc("/api/send/valid-order", corsMiddleware(server.handleValidOrder))
	http.HandleFunc("/api/send/retry-order", corsMiddleware(server.handleRetryOrder))
	http.HandleFunc("/api/send/invalid-order", corsMiddleware(server.handleInvalidOrder))
	http.HandleFunc("/", serveStatic)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[API GATEWAY] http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch path {
	case "/style.css":
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "web/style.css")
	case "/script.js":
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "web/script.js")
	default:
		http.ServeFile(w, r, "web/index.html")
	}
}
