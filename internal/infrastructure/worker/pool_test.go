package worker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GrishaMelixov/wealthcheck/internal/infrastructure/worker"
)

func TestPool_ProcessesAllJobs(t *testing.T) {
	p := worker.New(4, 100)

	var count atomic.Int64
	for i := 0; i < 50; i++ {
		_ = p.Submit(func(ctx context.Context) error {
			count.Add(1)
			return nil
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}

	if got := count.Load(); got != 50 {
		t.Fatalf("expected 50 jobs processed, got %d", got)
	}
}

func TestPool_ErrPoolFullWhenQueueSaturated(t *testing.T) {
	p := worker.New(1, 1)

	// Block the single worker so the queue fills up
	blocker := make(chan struct{})
	_ = p.Submit(func(ctx context.Context) error {
		<-blocker
		return nil
	})

	_ = p.Submit(func(ctx context.Context) error { return nil }) // fills queue

	err := p.Submit(func(ctx context.Context) error { return nil }) // must fail
	if err == nil {
		t.Fatal("expected ErrPoolFull, got nil")
	}

	close(blocker)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = p.Shutdown(ctx)
}
