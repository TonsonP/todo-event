package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuditEntry struct {
	ID        bson.ObjectID `bson:"_id" json:"id"`
	EntityID  string        `bson:"entity_id" json:"entity_id"`
	EventType string        `bson:"event_type" json:"event_type"`
	Payload   any           `bson:"payload" json:"payload"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}
