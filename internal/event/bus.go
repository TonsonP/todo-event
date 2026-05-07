package event

import (
	"context"
	"time"
)

type Event struct {
	Type      string
	EntityID  string
	Payload   any
	CreatedAt time.Time
}

type Handler func(context.Context, Event) error

type Publisher interface {
	Publish(ctx context.Context, e Event) error
}

type Subscriber interface {
	Subscribe(eventType string, fn Handler)
}
