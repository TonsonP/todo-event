package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/domain/task/domain"
	"todoe/domain/task/port"
	"todoe/internal/event"
)

var (
	ErrInvalidTitle  = errors.New("title must not be empty")
	ErrInvalidStatus = errors.New("invalid status")
	ErrTaskNotFound  = errors.New("task not found")
)

type Service struct {
	writeBus port.Publisher
	readBus  port.ReadBus
}

var _ port.UseCase = (*Service)(nil)

func NewService(writeBus port.Publisher, readBus port.ReadBus) *Service {
	return &Service{
		writeBus: writeBus,
		readBus:  readBus,
	}
}

func validateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return ErrInvalidTitle
	}

	return nil
}

func validateStatus(s domain.Status) error {
	switch s {
	case domain.StatusPending, domain.StatusInProgress, domain.StatusDone:
		return nil
	default:
		return ErrInvalidStatus
	}
}

func (s *Service) CreateTask(ctx context.Context, title string) mo.Result[domain.Task] {
	if err := validateTitle(title); err != nil {
		return mo.Err[domain.Task](err)
	}

	now := time.Now().UTC()

	task := domain.Task{
		ID:        bson.NewObjectID(),
		Title:     title,
		Status:    domain.StatusPending,
		CreatedAt: now,
	}

	err := s.writeBus.Publish(ctx, event.Event{
		Type:     domain.EventCreated,
		EntityID: task.ID.Hex(),
		Payload: domain.TaskCreatedPayload{
			Title: task.Title,
		},
		CreatedAt: now,
	})
	if err != nil {
		return mo.Err[domain.Task](err)
	}

	return mo.Ok(task)
}

func (s *Service) ListTasks(ctx context.Context) mo.Result[[]domain.Task] {
	records := s.readBus.All()

	tasks := make([]domain.Task, 0, len(records))

	for _, record := range records {
		task, ok := record.State.(domain.Task)
		if !ok {
			return mo.Err[[]domain.Task](
				fmt.Errorf("invalid task state for entity_id %s", record.EntityID),
			)
		}

		tasks = append(tasks, task)
	}

	return mo.Ok(tasks)
}

func (s *Service) GetTask(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task] {
	record, ok := s.readBus.Get(id.Hex())
	if !ok {
		return mo.Err[domain.Task](ErrTaskNotFound)
	}

	task, ok := record.State.(domain.Task)
	if !ok {
		return mo.Err[domain.Task](
			fmt.Errorf("invalid task state for entity_id %s", id.Hex()),
		)
	}

	return mo.Ok(task)
}

func (s *Service) ChangeStatus(
	ctx context.Context,
	id bson.ObjectID,
	status domain.Status,
) mo.Result[domain.Task] {
	if err := validateStatus(status); err != nil {
		return mo.Err[domain.Task](err)
	}

	current, ok := s.readBus.Get(id.Hex())
	if !ok {
		return mo.Err[domain.Task](ErrTaskNotFound)
	}

	if _, ok := current.State.(domain.Task); !ok {
		return mo.Err[domain.Task](
			fmt.Errorf("invalid task state for entity_id %s", id.Hex()),
		)
	}

	err := s.writeBus.Publish(ctx, event.Event{
		Type:     domain.EventStatusChanged,
		EntityID: id.Hex(),
		Payload: domain.StatusChangedPayload{
			Status: status,
		},
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return mo.Err[domain.Task](err)
	}

	record, ok := s.readBus.Get(id.Hex())
	if !ok {
		return mo.Err[domain.Task](ErrTaskNotFound)
	}

	task, ok := record.State.(domain.Task)
	if !ok {
		return mo.Err[domain.Task](
			fmt.Errorf("invalid task state for entity_id %s", id.Hex()),
		)
	}

	return mo.Ok(task)
}
