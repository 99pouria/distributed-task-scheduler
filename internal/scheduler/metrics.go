package scheduler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TasksSubmitted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scheduler_tasks_submitted_total",
			Help: "Total number of tasks submitted.",
		},
		[]string{"priority"},
	)
	TasksProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scheduler_tasks_processed_total",
			Help: "Total number of tasks processed successfully.",
		},
		[]string{"priority"},
	)
	TasksFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scheduler_tasks_failed_total",
			Help: "Total number of tasks that failed processing.",
		},
		[]string{"priority"},
	)
	TaskProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "scheduler_task_processing_duration_seconds",
			Help:    "Histogram of task processing times.",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
		},
		[]string{"priority"},
	)
)
