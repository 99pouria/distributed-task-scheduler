package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/99pouria/distributed-task-scheduler/internal/api"
	"github.com/99pouria/distributed-task-scheduler/internal/queue"
	"github.com/99pouria/distributed-task-scheduler/internal/scheduler"
	"github.com/99pouria/distributed-task-scheduler/internal/storage"
)

func main() {
	// Read environment variables
	rabbitURL := os.Getenv("RABBITMQ_URL")
	httpAddr := os.Getenv("HTTP_ADDR")
	redisURL := os.Getenv("REDIS_URL")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	workerCount, err := strconv.Atoi(os.Getenv("WORKER_COUNT"))
	if err != nil {
		log.Fatalf("Invalid WORKER_COUNT: %v, error: %v", os.Getenv("WORKER_COUNT"), err)
	}

	// Initialize dependencies
	store := storage.NewRedisTaskStore(redisURL, redisPassword, 0)

	log.Println(rabbitURL)
	broker, err := queue.NewBroker(rabbitURL)
	if err != nil {
		log.Fatalf("Could not connect to broker: %v", err)
	}
	defer broker.Close()

	if err := broker.Setup(); err != nil {
		log.Fatalf("Could not setup broker queue: %v", err)
	}

	// Start worker pool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 1; i <= workerCount; i++ {
		worker := scheduler.NewWorker(i, broker, *store)
		go worker.Start(ctx)
	}

	// Start API server
	api := api.NewAPI(broker, store)
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("Starting API server on %s", httpAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", httpAddr, err)
		}
	}()

	// Graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %+v", err)
	}
	cancel() // Signal workers to stop
	log.Println("Server stopped gracefully.")
}
