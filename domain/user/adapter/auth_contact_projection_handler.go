package adapter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	userv1 "todoe/gen/user/v1"
	"todoe/internal/event"
)

type ContactUpdatedEventPayload struct {
	UserID string `json:"user_id"`
}

func NewAuthContactUpdatedProjectionHandler(
	repo *AuthCredentialMongoRepository,
	userClient userv1.UserServiceClient,
) func(context.Context, event.Event) error {
	return func(ctx context.Context, e event.Event) error {
		payload, ok := e.Payload.(ContactUpdatedEventPayload)
		if !ok {
			return fmt.Errorf("unexpected payload type %T", e.Payload)
		}

		if payload.UserID == "" {
			return fmt.Errorf("missing user_id in contact updated event")
		}

		grpcCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		userResp, err := userClient.GetUser(grpcCtx, &userv1.GetUserRequest{
			UserId: payload.UserID,
		})
		if err != nil {
			return fmt.Errorf("get user via grpc: %w", err)
		}

		fmt.Println(userResp)

		result := repo.UpdateContact(ctx, userResp.UserId, userResp.Email)
		if result.IsError() {
			return result.Error()
		}

		slog.Info("auth credential synced from user service",
			"user_id", userResp.UserId,
			"email", userResp.Email,
		)

		return nil
	}
}
