package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Handler func(ctx context.Context, e Event) error

type Pool struct {
	queue   chan Event
	handler Handler
	workers int
	log     *zap.Logger
	wg      sync.WaitGroup
}

func NewPool(workers, queueSize int, log *zap.Logger, handler Handler) *Pool {
	if workers < 1 {
		workers = 1
	}
	if queueSize < 1 {
		queueSize = 64
	}
	return &Pool{
		queue:   make(chan Event, queueSize),
		handler: handler,
		workers: workers,
		log:     log,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.run(ctx, i)
	}
	p.log.Info("worker pool started",
		zap.Int("workers", p.workers),
		zap.Int("queue_size", cap(p.queue)),
	)
}

func (p *Pool) Publish(e Event) {
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now()
	}
	select {
	case p.queue <- e:
	case <-time.After(50 * time.Millisecond):
		p.log.Warn("worker queue full, dropping event",
			zap.String("type", string(e.Type)),
			zap.Int64("user_id", e.UserID),
		)
	}
}

func (p *Pool) Stop() {
	close(p.queue)
	p.wg.Wait()
	p.log.Info("worker pool stopped")
}

func (p *Pool) run(ctx context.Context, id int) {
	defer p.wg.Done()
	log := p.log.With(zap.Int("worker_id", id))
	log.Info("worker started")

	for {
		select {
		case <-ctx.Done():
			log.Info("worker stopping (context cancelled)")
			return
		case e, ok := <-p.queue:
			if !ok {
				log.Info("worker stopping (queue closed)")
				return
			}
			start := time.Now()

			hctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			if err := p.handler(hctx, e); err != nil {
				log.Error("event handler failed",
					zap.String("type", string(e.Type)),
					zap.Error(err),
				)
			}
			cancel()
			log.Debug("event processed",
				zap.String("type", string(e.Type)),
				zap.Duration("took", time.Since(start)),
			)
		}
	}
}
