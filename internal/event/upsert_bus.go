package event

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type StateRecord struct {
	EntityID  string
	Type      string
	State     any
	Version   int64
	UpdatedAt time.Time
}

type UpsertBus struct {
	mu       sync.RWMutex
	records  map[string]StateRecord
	handlers map[string][]Handler
}

func NewUpsertBus() *UpsertBus {
	return &UpsertBus{
		records:  make(map[string]StateRecord),
		handlers: make(map[string][]Handler),
	}
}

func (b *UpsertBus) Subscribe(eventType string, fn Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], fn)
}

func (b *UpsertBus) Publish(ctx context.Context, e Event) error {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}

	b.mu.Lock()

	current := b.records[e.EntityID]

	next := StateRecord{
		EntityID:  e.EntityID,
		Type:      e.Type,
		State:     e.Payload,
		Version:   current.Version + 1,
		UpdatedAt: e.CreatedAt,
	}

	b.records[e.EntityID] = next

	handlers := append([]Handler(nil), b.handlers[e.Type]...)

	b.mu.Unlock()

	for _, fn := range handlers {
		if err := fn(ctx, e); err != nil {
			slog.Error(
				"upsert bus: handler error",
				"event_type", e.Type,
				"entity_id", e.EntityID,
				"err", err,
			)
		}
	}

	return nil
}

func (b *UpsertBus) Get(entityID string) (StateRecord, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	record, ok := b.records[entityID]
	return record, ok
}

func (b *UpsertBus) All() []StateRecord {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]StateRecord, 0, len(b.records))

	for _, record := range b.records {
		result = append(result, record)
	}

	return result
}

var _ Publisher = (*UpsertBus)(nil)
var _ Subscriber = (*UpsertBus)(nil)
