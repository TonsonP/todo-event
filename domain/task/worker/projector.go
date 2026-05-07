package worker

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/domain/task/domain"
	"todoe/internal/event"
)

const EventTaskLatest = "task.latest"

type ReadModelRepository interface {
	UpsertReadModel(ctx context.Context, task domain.Task) error
}

type TaskProjector struct {
	readBus *event.UpsertBus
	repo    ReadModelRepository
}

func NewTaskProjector(readBus *event.UpsertBus, repo ReadModelRepository) *TaskProjector {
	return &TaskProjector{
		readBus: readBus,
		repo:    repo,
	}
}

func (p *TaskProjector) Register(writeBus event.Subscriber) {
	writeBus.Subscribe(domain.EventCreated, p.OnTaskCreated)
	writeBus.Subscribe(domain.EventStatusChanged, p.OnTaskStatusChanged)
}

func (p *TaskProjector) OnTaskCreated(ctx context.Context, e event.Event) error {
	payload, ok := e.Payload.(domain.TaskCreatedPayload)
	if !ok {
		return fmt.Errorf("invalid payload for %s", e.Type)
	}

	id, err := bson.ObjectIDFromHex(e.EntityID)
	if err != nil {
		return err
	}

	task := domain.Task{
		ID:        id,
		Title:     payload.Title,
		Status:    domain.StatusPending,
		CreatedAt: e.CreatedAt,
	}

	if err := p.readBus.Publish(ctx, event.Event{
		Type:      EventTaskLatest,
		EntityID:  e.EntityID,
		Payload:   task,
		CreatedAt: e.CreatedAt,
	}); err != nil {
		return err
	}

	if err := p.repo.UpsertReadModel(ctx, task); err != nil {
		return err
	}

	return nil
}

func (p *TaskProjector) OnTaskStatusChanged(ctx context.Context, e event.Event) error {
	payload, ok := e.Payload.(domain.StatusChangedPayload)
	if !ok {
		return fmt.Errorf("invalid payload for %s", e.Type)
	}

	current, ok := p.readBus.Get(e.EntityID)
	if !ok {
		return fmt.Errorf("task state not found: %s", e.EntityID)
	}

	task, ok := current.State.(domain.Task)
	if !ok {
		return fmt.Errorf("invalid task state for %s", e.EntityID)
	}

	task.Status = payload.Status

	if err := p.readBus.Publish(ctx, event.Event{
		Type:      EventTaskLatest,
		EntityID:  e.EntityID,
		Payload:   task,
		CreatedAt: e.CreatedAt,
	}); err != nil {
		return err
	}

	if err := p.repo.UpsertReadModel(ctx, task); err != nil {
		return err
	}

	return nil
}
