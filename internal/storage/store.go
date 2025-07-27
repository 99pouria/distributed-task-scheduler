package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/99pouria/distributed-task-scheduler/internal/scheduler/task"

	"github.com/redis/go-redis/v9"
)

// RedisTaskStore represents a redis storage for tasks
type RedisTaskStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisTaskStore connects to redis and returns a new instance of RedisTaskStore
func NewRedisTaskStore(addr, password string, db int) *RedisTaskStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisTaskStore{
		client: rdb,
		ctx:    context.Background(),
	}
}

// GetTask returns Task of given id from the storage
func (s *RedisTaskStore) GetTask(id string) (*task.Task, error) {
	data, err := s.client.Get(s.ctx, id).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("task with ID %s not found", id)
	} else if err != nil {
		return nil, err
	}

	var task task.Task
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// SetTask adds a task to the storage
func (s *RedisTaskStore) SetTask(task *task.Task) error {
	if task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := s.client.Set(s.ctx, task.ID, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to set task in Redis: %w", err)
	}

	return nil
}
