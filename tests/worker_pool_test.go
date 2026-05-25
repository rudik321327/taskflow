package tests

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/taskflow/taskflow/internal/worker"
)

func TestWorkerPool_ProcessesEventsConcurrently(t *testing.T) {
	var processed int64
	pool := worker.NewPool(3, 64, zap.NewNop(), func(ctx context.Context, e worker.Event) error {
		atomic.AddInt64(&processed, 1)
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	const n = 20
	for i := 0; i < n; i++ {
		pool.Publish(worker.Event{Type: worker.EventTaskCreated, UserID: int64(i)})
	}

	pool.Stop()
	require.EqualValues(t, n, atomic.LoadInt64(&processed))
}

func TestWorkerPool_CancellationStopsWorkers(t *testing.T) {
	pool := worker.NewPool(2, 8, zap.NewNop(), func(ctx context.Context, e worker.Event) error {
		<-ctx.Done()
		return ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	pool.Start(ctx)

	pool.Publish(worker.Event{Type: worker.EventTaskCreated})
	cancel()

	done := make(chan struct{})
	go func() { pool.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("pool did not stop after context cancellation")
	}
}
