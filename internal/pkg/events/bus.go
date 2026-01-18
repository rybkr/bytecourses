package events

import (
	"bytecourses/internal/domain"
	"context"
	"log/slog"
	"sync"
)

type Handler func(ctx context.Context, event domain.DomainEvent) error

type EventBus interface {
	Subscribe(eventName string, handler Handler)
	Publish(ctx context.Context, event domain.DomainEvent) error
}

type InMemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	logger   *slog.Logger
}

func NewInMemoryEventBus(logger *slog.Logger) *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]Handler),
		logger:   logger,
	}
}

func (b *InMemoryEventBus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventName] = append(b.handlers[eventName], handler)
	b.logger.Debug("event handler registered",
		"event", eventName,
		"handlers_count", len(b.handlers[eventName]),
	)
}

func (b *InMemoryEventBus) Publish(ctx context.Context, event domain.DomainEvent) error {
	b.mu.RLock()
	handlers := b.handlers[event.EventName()]
	b.mu.RUnlock()

	if len(handlers) == 0 {
		b.logger.Debug("no handlers for event",
			"event", event.EventName(),
		)
		return nil
	}

	b.logger.Debug("publishing event",
		"event", event.EventName(),
		"handlers_count", len(handlers),
	)

	for i, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			b.logger.Error("event handler failed",
				"event", event.EventName(),
				"handler_index", i,
				"error", err,
			)
			// Continue processing other handlers even if one fails
		}
	}

	return nil
}

type eventJob struct {
	ctx   context.Context
	event domain.DomainEvent
}

type AsyncEventBus struct {
	*InMemoryEventBus
	nWorkers int
	queue    chan eventJob
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewAsyncEventBus(logger *slog.Logger, nWorkers int) *AsyncEventBus {
	ctx, cancel := context.WithCancel(context.Background())

	bus := &AsyncEventBus{
		InMemoryEventBus: NewInMemoryEventBus(logger),
		nWorkers:         nWorkers,
		queue:            make(chan eventJob, 100),
		ctx:              ctx,
		cancel:           cancel,
	}

	for i := 0; i < nWorkers; i++ {
		bus.wg.Add(1)
		go bus.worker(i)
	}

	return bus
}

func (b *AsyncEventBus) worker(id int) {
	defer b.wg.Done()
	b.logger.Debug("event worker started", "worker_id", id)

	for {
		select {
		case <-b.ctx.Done():
			b.logger.Debug("event worker stopped", "worker_id", id)
			return

		case job := <-b.queue:
			b.logger.Debug("worker processing event",
				"worker_id", id,
				"event", job.event.EventName(),
			)

			if err := b.InMemoryEventBus.Publish(job.ctx, job.event); err != nil {
				b.logger.Error("worker failed to process event",
					"worker_id", id,
					"event", job.event.EventName(),
					"error", err,
				)
			}
		}
	}
}

func (b *AsyncEventBus) Publish(ctx context.Context, event domain.DomainEvent) error {
	select {
	case b.queue <- eventJob{
		ctx:   ctx,
		event: event,
	}:
		b.logger.Debug("event enqueued",
			"event", event.EventName(),
			"queue_length", len(b.queue),
		)
		return nil

	case <-b.ctx.Done():
		return b.ctx.Err()
	}
}

func (b *AsyncEventBus) Shutdown() {
	b.cancel()
	close(b.queue)
	b.wg.Wait()
	b.logger.Info("async event bus shutdown complete")
}
