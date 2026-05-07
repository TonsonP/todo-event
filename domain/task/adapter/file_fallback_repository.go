package adapter

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"todoe/domain/task/domain"
)

type FileFallbackRepository struct {
	filePath string
	mu       sync.Mutex
}

func NewFileFallbackRepository(filePath string) *FileFallbackRepository {
	return &FileFallbackRepository{
		filePath: filePath,
	}
}

func (r *FileFallbackRepository) WriteTaskCreated(ctx context.Context, task domain.Task, cause error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.OpenFile(r.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	record := map[string]any{
		"type":       domain.EventCreated,
		"task":       task,
		"cause":      cause.Error(),
		"written_at": time.Now().Format(time.RFC3339),
	}

	line, err := json.Marshal(record)
	if err != nil {
		return err
	}

	_, err = file.Write(append(line, '\n'))
	return err
}
