package event

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type AppendRecord struct {
	Seq int64
	Event
}

type AppendBus struct {
	mu       sync.RWMutex
	seq      int64
	records  []AppendRecord
	handlers map[string][]Handler
}

func NewAppendBus() *AppendBus {
	return &AppendBus{
		records:  make([]AppendRecord, 0),
		handlers: make(map[string][]Handler),
	}
}

func (b *AppendBus) Subscribe(eventType string, fn Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], fn)
}

func (b *AppendBus) Publish(ctx context.Context, e Event) error {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}

	b.mu.Lock()

	b.seq++

	record := AppendRecord{
		Seq:   b.seq,
		Event: e,
	}

	b.records = append(b.records, record)

	handlers := append([]Handler(nil), b.handlers[e.Type]...)

	b.mu.Unlock()

	for _, fn := range handlers {
		if err := fn(ctx, e); err != nil {
			slog.Error(
				"append bus: handler error",
				"event_type", e.Type,
				"entity_id", e.EntityID,
				"err", err,
			)
		}
	}

	return nil
}

func (b *AppendBus) Records() []AppendRecord {
	b.mu.RLock()
	defer b.mu.RUnlock()

	records := make([]AppendRecord, len(b.records))
	copy(records, b.records)

	return records
}

func (b *AppendBus) RecordsByEntityID(entityID string) []AppendRecord {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]AppendRecord, 0)

	for _, record := range b.records {
		if record.EntityID == entityID {
			result = append(result, record)
		}
	}

	return result
}

var _ Publisher = (*AppendBus)(nil)
var _ Subscriber = (*AppendBus)(nil)
