package adapter

import (
	"context"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/domain/task/domain"
	"todoe/domain/task/port"
)

type ResilientRepository struct {
	primary  port.Repository
	fallback *FileFallbackRepository
}

func NewResilientRepository(primary port.Repository, fallback *FileFallbackRepository) *ResilientRepository {
	return &ResilientRepository{
		primary:  primary,
		fallback: fallback,
	}
}

func (r *ResilientRepository) Save(ctx context.Context, task domain.Task) mo.Result[struct{}] {
	result := r.primary.Save(ctx, task)
	if !result.IsError() {
		return result
	}

	if err := r.fallback.WriteTaskCreated(ctx, task, result.Error()); err != nil {
		return mo.Err[struct{}](err)
	}

	return mo.Ok(struct{}{})
}

func (r *ResilientRepository) FindAll(ctx context.Context) mo.Result[[]domain.Task] {
	return r.primary.FindAll(ctx)
}

func (r *ResilientRepository) FindByID(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task] {
	return r.primary.FindByID(ctx, id)
}

var _ port.Repository = (*ResilientRepository)(nil)
