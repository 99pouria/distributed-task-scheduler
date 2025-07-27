package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/99pouria/distributed-task-scheduler/internal/queue"
	"github.com/99pouria/distributed-task-scheduler/internal/scheduler/task"
	"github.com/99pouria/distributed-task-scheduler/internal/storage"

	"github.com/rabbitmq/amqp091-go"
)

// Worker pulls tasks from the broker and executes them.
type Worker struct {
	id     int
	broker *queue.Broker
	store  storage.RedisTaskStore
}

// NewWorker returns a new instance of Worker
func NewWorker(id int, broker *queue.Broker, store storage.RedisTaskStore) *Worker {
	return &Worker{id: id, broker: broker, store: store}
}

// Start begins the worker's task consumption loop.
func (w *Worker) Start(ctx context.Context) {
	log.Printf("Starting worker %d", w.id)

	ch, err := w.broker.Conn.Channel()
	if err != nil {
		log.Fatalf("Worker %d failed to open channel: %v", w.id, err)
	}
	defer ch.Close()

	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	msgs, err := ch.Consume(
		queue.TaskQueueName,
		"",
		false, // we ack manually when a task processed successfully
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Worker %d failed to register a consumer: %v", w.id, err)
	}

	go func() {
		for d := range msgs {
			w.processDelivery(d)
		}
	}()

	<-ctx.Done()
	log.Printf("Shutting down worker %d", w.id)
}

// processDelivery tries to consume delivered task with random chance of failure.
// In failure case, it requeues the task for another attempt.
func (w *Worker) processDelivery(d amqp091.Delivery) {
	var receivedTask task.Task
	if err := json.Unmarshal(d.Body, &receivedTask); err != nil {
		log.Printf("Error unmarshalling task: %v. Rejecting message.", err)
		d.Reject(false)
		return
	}

	log.Printf("Worker %d received task %s with priority %s", w.id, receivedTask.ID, receivedTask.Priority.String())
	startTime := time.Now()

	receivedTask.Status = task.StatusRunning
	w.store.SetTask(&receivedTask)

	// Proccessing task (with a 10% faliure chance)
	time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)
	if rand.Intn(10) == 0 {
		log.Printf("Worker %d failed to process task %s", w.id, receivedTask.ID)
		receivedTask.Status = task.StatusFailed
		w.store.SetTask(&receivedTask)
		TasksFailed.WithLabelValues(receivedTask.Priority.String()).Inc()
		
		// Nack and requeue the task for another attempt
		d.Nack(false, true)
		return
	}

	log.Printf("Worker %d finished processing task %s", w.id, receivedTask.ID)
	receivedTask.Status = task.StatusCompleted
	w.store.SetTask(&receivedTask)

	processingTime := time.Since(startTime).Seconds()
	TaskProcessingDuration.WithLabelValues(receivedTask.Priority.String()).Observe(processingTime)
	TasksProcessed.WithLabelValues(receivedTask.Priority.String()).Inc()

	// Acknowledge the message was processed successfully
	d.Ack(false) 
}
