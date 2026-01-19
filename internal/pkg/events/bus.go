package events

import (
	"bytecourses/internal/domain"
	"context"
	"log/slog"
	"sync"
)

type EventHandler func(ctx context.Context, event domain.Event) error

type EventBus interface {
	Subscribe(eventName string, handler EventHandler)
	Publish(ctx context.Context, event domain.Event) error
}

var (
    _ EventBus = (*InMemoryEventBus)(nil)
)

type InMemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
	logger   *slog.Logger
}

func NewInMemoryEventBus(logger *slog.Logger) *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger,
	}
}

func (b *InMemoryEventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventName] = append(b.handlers[eventName], handler)
	b.logger.Debug("event handler registered",
		"event", eventName,
		"handlers_count", len(b.handlers[eventName]),
	)
}

func (b *InMemoryEventBus) Publish(ctx context.Context, event domain.Event) error {
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
