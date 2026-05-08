package port

import (
	"context"

	"github.com/samber/mo"

	"todoe/domain/user/domain"
	"todoe/internal/event"
)

type UseCase interface {
	Register(ctx context.Context, name, email string) mo.Result[domain.User]
	VerifyEmail(ctx context.Context, id, token string) mo.Result[domain.User]
	RecordCreditScore(ctx context.Context, id string, score int, approved bool) mo.Result[domain.User]
	CompleteProfile(ctx context.Context, id, bio string) mo.Result[domain.User]
	GetUser(ctx context.Context, id string) mo.Result[domain.User]
	ListActivatedUsers(ctx context.Context) mo.Result[[]domain.User]
	UpdateContact(ctx context.Context, id, name, email, bio string) mo.Result[domain.User]
	GetUserHistory(ctx context.Context, id string) mo.Result[[]domain.UserEvent]
}

type Repository interface {
	Upsert(ctx context.Context, user domain.User) mo.Result[struct{}]
	Append(ctx context.Context, aggregateID, eventType string, payload any) mo.Result[struct{}]
	FindByID(ctx context.Context, id string) mo.Result[domain.User]
	FindByEmail(ctx context.Context, email string) mo.Result[domain.User]
	FindActivated(ctx context.Context) mo.Result[[]domain.User]
	FindEvents(ctx context.Context, aggregateID string) mo.Result[[]domain.UserEvent]
}

type Publisher interface {
	Publish(ctx context.Context, e event.Event)
}
