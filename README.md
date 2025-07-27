# distributed-task-scheduler
Designed and implemented a distributed task scheduler in Go that handles prioritized tasks across multiple worker nodes
## Overview

`distributed-task-scheduler` is a Go-based system designed to efficiently schedule and execute prioritized tasks across multiple workers . The scheduler ensures high availability, fault tolerance, and optimal resource utilization.

## Features

- **Distributed Task Scheduling:** Assigns tasks to workers in a cluster.
- **Task Prioritization:** Supports prioritizing tasks to ensure critical jobs are executed first.
- **Scalability:** Manually add or remove workers to scale horizontally.
- **Fault Tolerance:** Failed tasks requeues and reassigns to a worker to healthy process again.
- **Monitoring:** Provides basic metrics and logging for task execution and node health.

## Architecture

- **API Server:** Central API server that serves client's requests. There are main three routes:
    - ***/submit*** creates a new task and adds it to queue.
    - ***/status/{id}*** gets status of a task.
    - ***/metrics*** uses for monitoring data that is compatible with Prometheus.
- **Workers Pool:** Execute assigned tasks and report status back to the scheduler. Also, when a task fails, worker sends a nack signal to the queue in order to requeue the task. Afterwards, a worker picks failed task and tries to consume it.
- **Priority Queue:** There is a rabbitmq service that holds a prioriry queue containing tasks. Each worker will pick a task from this queue.
- **Redis Database:** A redis server established in order to store each task's JSON encoded detail as a key value.   

### Architecture Diagram
In this diagram, you can see the schema of the system:<br><br>
![Architecture Diagram](Architecture%20Diagram.png)


## Tests And Benchmarks
- The storage layer has been thoroughly tested under various scenarios:
    - Creating and storing a task, then reading it back successfully.
    - Attempting to read a task that was never stored, verifying the correct handling of missing data.
    - Attempting to store an invalid or malformed task, ensuring the system correctly handles and rejects incorrect inputs.
- In addition to functional tests, benchmarking was performed to measure the performance of storing and reading tasks in the storage system. The average time per operation (ns/op) is very close for both:
    - **SetTask** takes about 324,235 ns/op (approximately 324 microseconds).
    - **GetTask** takes about 323,429 ns/op (approximately 323 microseconds).
