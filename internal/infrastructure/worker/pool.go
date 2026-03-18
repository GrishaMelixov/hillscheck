package worker

import (
	"context"
	"sync"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

// Pool is a fixed-size goroutine pool backed by a buffered channel.
// It implements port.WorkerPool.
type Pool struct {
	jobs   chan port.Job
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func New(workers, queueSize int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		jobs:   make(chan port.Job, queueSize),
		ctx:    ctx,
		cancel: cancel,
	}
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.run()
	}
	return p
}

func (p *Pool) run() {
	defer p.wg.Done()
	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			_ = job(p.ctx)
		case <-p.ctx.Done():
			return
		}
	}
}

// Submit enqueues a job without blocking.
// Returns domain.ErrPoolFull if the buffered channel is at capacity.
func (p *Pool) Submit(job port.Job) error {
	select {
	case p.jobs <- job:
		return nil
	default:
		return domain.ErrPoolFull
	}
}

// Shutdown closes the job queue and waits for workers to drain remaining jobs.
// If ctx expires before drain completes, workers are force-cancelled.
func (p *Pool) Shutdown(ctx context.Context) error {
	close(p.jobs)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		p.cancel()
		p.wg.Wait()
		return ctx.Err()
	}
}
