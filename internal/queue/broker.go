package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/99pouria/distributed-task-scheduler/internal/scheduler/task"

	"github.com/rabbitmq/amqp091-go"
)

const TaskQueueName = "tasks_priority_queue"

// Broker handles the connection and communication with RabbitMQ.
type Broker struct {
	Conn *amqp091.Connection
}

// NewBroker connects to a rabbitmq server and returns a Broker instance
func NewBroker(url string) (*Broker, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return &Broker{Conn: conn}, nil
}

// Setup declares the priority queue.
func (b *Broker) Setup() error {
	ch, err := b.Conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare a single queue with priority support
	args := amqp091.Table{
		"x-max-priority": 10,
	}
	_, err = ch.QueueDeclare(
		TaskQueueName,
		true,
		false,
		false,
		false,
		args,
	)
	return err
}

// PublishTask sends a task to the queue.
func (b *Broker) PublishTask(ctx context.Context, task *task.Task) error {
	ch, err := b.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = ch.PublishWithContext(
		ctx,
		"",
		TaskQueueName,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent, // Make message persistent
			Priority:     uint8(task.Priority),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	log.Printf("Published task %s with priority %s", task.ID, task.Priority.String())
	return nil
}

// Close closes connection to the broker server.
func (b *Broker) Close() {
	if b.Conn != nil {
		b.Conn.Close()
	}
}
