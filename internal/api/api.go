package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/99pouria/distributed-task-scheduler/internal/queue"
	"github.com/99pouria/distributed-task-scheduler/internal/scheduler"
	"github.com/99pouria/distributed-task-scheduler/internal/scheduler/task"
	"github.com/99pouria/distributed-task-scheduler/internal/storage"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// API holds dependencies for the HTTP handlers.
type API struct {
	broker *queue.Broker
	store  *storage.RedisTaskStore
}

func NewAPI(broker *queue.Broker, store *storage.RedisTaskStore) *API {
	return &API{broker: broker, store: store}
}

// RegisterRoutes sets up the HTTP routing.
func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /tasks", a.handleSubmitTask)
	mux.HandleFunc("GET /tasks/{id}", a.handleGetTaskStatus)
	mux.Handle("/metrics", promhttp.Handler())
}

// handleSubmitTask handles the creation of a new task.
func (a *API) handleSubmitTask(w http.ResponseWriter, r *http.Request) {
	var newTask task.Task
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newTask.ID = uuid.NewString()
	newTask.CreatedAt = time.Now().UTC()
	newTask.Status = task.StatusPending

	// Store task in initial state
	if err := a.store.SetTask(&newTask); err != nil {
		log.Printf("Error setting initial task state: %v", err)
		http.Error(w, "Failed to store task", http.StatusInternalServerError)
		return
	}

	// Publish to queue
	if err := a.broker.PublishTask(r.Context(), &newTask); err != nil {
		log.Printf("Error publishing task: %v", err)
		http.Error(w, "Failed to schedule task", http.StatusInternalServerError)
		return
	}

	scheduler.TasksSubmitted.WithLabelValues(newTask.Priority.String()).Inc()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(newTask)
}

// handleGetTaskStatus retrieves the status of a specific task.
func (a *API) handleGetTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	task, err := a.store.GetTask(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}
