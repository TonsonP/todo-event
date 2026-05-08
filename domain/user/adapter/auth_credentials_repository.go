package adapter

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AuthCredentialMongoRepository struct {
	clientIO    mo.IOEither[*mongo.Client]
	once        sync.Once
	cached      mo.Either[error, *mongo.Client]
	initialized atomic.Bool
}

func NewAuthCredentialMongoRepository(clientIO mo.IOEither[*mongo.Client]) *AuthCredentialMongoRepository {
	return &AuthCredentialMongoRepository{clientIO: clientIO}
}

func (r *AuthCredentialMongoRepository) getClient() mo.Either[error, *mongo.Client] {
	r.once.Do(func() {
		r.cached = r.clientIO.Run()
		r.initialized.Store(true)
	})
	return r.cached
}

func (r *AuthCredentialMongoRepository) db() (*mongo.Database, error) {
	either := r.getClient()
	if either.IsLeft() {
		return nil, either.MustLeft()
	}
	return either.MustRight().Database("todoe"), nil
}

func (r *AuthCredentialMongoRepository) UpdateContact(ctx context.Context, userID, email string) mo.Result[struct{}] {
	db, err := r.db()
	if err != nil {
		return mo.Err[struct{}](err)
	}

	filter := bson.D{{Key: "user_id", Value: userID}}

	update := bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "email", Value: email},
			{Key: "updated_at", Value: time.Now()},
		},
	}}

	result, err := db.Collection("auth_credentials").UpdateOne(ctx, filter, update)
	if err != nil {
		return mo.Err[struct{}](err)
	}

	if result.MatchedCount == 0 {
		return mo.Err[struct{}](fmt.Errorf("auth credential not found for user_id %s", userID))
	}

	return mo.Ok(struct{}{})
}
