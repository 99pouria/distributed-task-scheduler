package storage

import (
	"testing"
	"time"

	"github.com/99pouria/distributed-task-scheduler/internal/scheduler/task"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskStorage(t *testing.T) {
	redisURL := "localhost:6379"
	redisPassword := "password"
	store := NewRedisTaskStore(redisURL, redisPassword, 0)

	t.Run("SetAndGetTask", func(t *testing.T) {
		newTask := &task.Task{
			ID:        uuid.NewString(),
			Status:    task.StatusPending,
			Priority:  task.High,
			CreatedAt: time.Now(),
		}

		err := store.SetTask(newTask)
		require.NoError(t, err)

		retrievedTask, err := store.GetTask(newTask.ID)
		require.NoError(t, err)
		assert.Equal(t, newTask.ID, retrievedTask.ID)
		assert.Equal(t, task.StatusPending, retrievedTask.Status)
	})

	t.Run("GetNonExistentTask", func(t *testing.T) {
		_, err := store.GetTask("non-existent-id")
		assert.Error(t, err)
	})

	t.Run("SetTaskWithEmptyID", func(t *testing.T) {
		newTask := &task.Task{ID: ""}
		err := store.SetTask(newTask)
		assert.Error(t, err)
	})
}

// Benchmark for TaskStore operations
func BenchmarkInMemoryTaskStore_SetTask(b *testing.B) {
	redisURL := "localhost:6379"
	redisPassword := "password"
	store := NewRedisTaskStore(redisURL, redisPassword, 0)

	b.Run("BenchmarkSet", func(b *testing.B) {
		newTask := &task.Task{ID: "benchmark-id", Status: task.StatusPending}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.SetTask(newTask)
		}
	})

	b.Run("BenchmarkGet", func(b *testing.B) {
		newTask := &task.Task{ID: "benchmark-id", Status: task.StatusPending}
		store.SetTask(newTask)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.GetTask(newTask.ID)
		}
	})
}
