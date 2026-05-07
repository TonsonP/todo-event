package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/internal/audit/domain"
	"todoe/internal/event"
)

func NewAuditHandler(repo *MongoRepository) func(context.Context, event.Event) error {
	return func(ctx context.Context, e event.Event) error {
		entry := domain.AuditEntry{
			ID:        bson.NewObjectID(),
			EventType: e.Type,
			Payload:   e.Payload,
			CreatedAt: time.Now(),
		}
		slog.Info("audit: handling event", "event_type", e.Type)
		result := repo.Save(ctx, entry)
		if result.IsError() {
			return result.Error()
		}
		return nil
	}
}

var fallbackMu sync.Mutex

func appendAuditFallback(filePath string, entry domain.AuditEntry, mongoErr error) error {
	fallbackMu.Lock()
	defer fallbackMu.Unlock()

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	record := map[string]any{
		"id":         entry.ID.Hex(),
		"event_type": entry.EventType,
		"payload":    entry.Payload,
		"created_at": entry.CreatedAt.Format(time.RFC3339),
		"mongo_err":  mongoErr.Error(),
		"written_at": time.Now().Format(time.RFC3339),
	}

	line, err := json.Marshal(record)
	if err != nil {
		return err
	}

	if _, err := file.Write(append(line, '\n')); err != nil {
		return err
	}

	return nil
}

func NewAuditHandlerWithRetry(repo *MongoRepository, fallbackPath string) func(context.Context, event.Event) error {
	return func(ctx context.Context, e event.Event) error {
		entry := domain.AuditEntry{
			ID:        bson.NewObjectID(),
			EventType: e.Type,
			Payload:   e.Payload,
			CreatedAt: time.Now(),
		}

		slog.Info("audit: handling event", "event_type", e.Type)

		var lastErr error

		for attempt := 1; attempt <= 3; attempt++ {
			attemptCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			result := repo.Save(attemptCtx, entry)
			cancel()

			if !result.IsError() {
				slog.Info(
					"audit: saved to mongo",
					"event_type", e.Type,
					"attempt", attempt,
				)
				return nil
			}

			lastErr = result.Error()

			slog.Warn(
				"audit: mongo save failed",
				"event_type", e.Type,
				"attempt", attempt,
				"err", lastErr,
			)

			if attempt < 3 {
				time.Sleep(1 * time.Second)
			}
		}

		slog.Error(
			"audit: mongo unavailable after retries, writing fallback file",
			"event_type", e.Type,
			"err", lastErr,
		)

		if err := appendAuditFallback(fallbackPath, entry, lastErr); err != nil {
			return fmt.Errorf("mongo error: %w; fallback file error: %v", lastErr, err)
		}

		return nil
	}
}
