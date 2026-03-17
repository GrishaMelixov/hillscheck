package port

import "context"

// Job is a unit of async work submitted to the pool.
type Job func(ctx context.Context) error

// WorkerPool manages a fixed set of goroutines for background processing.
// On graceful shutdown the pool drains its queue before returning.
type WorkerPool interface {
	Submit(job Job) error
	Shutdown(ctx context.Context) error
}
