package adapter

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/internal/audit/domain"
	"todoe/internal/event"
)

func NewAuditHandler(repo *MongoRepository) func(context.Context, event.Event) error {
	return func(ctx context.Context, e event.Event) error {
		createdAt := e.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}

		entry := domain.AuditEntry{
			ID:        bson.NewObjectID(),
			EntityID:  e.EntityID,
			EventType: e.Type,
			Payload:   e.Payload,
			CreatedAt: createdAt,
		}
		result := repo.Save(ctx, entry)
		if result.IsError() {
			return result.Error()
		}
		return nil
	}
}
