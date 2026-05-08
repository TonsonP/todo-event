package grpcadapter

import (
	"context"

	userport "todoe/domain/user/port"
	userv1 "todoe/gen/user/v1"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	useCase userport.UseCase
}

func NewServer(useCase userport.UseCase) *Server {
	return &Server{
		useCase: useCase,
	}
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	result := s.useCase.GetUser(ctx, req.UserId)
	if result.IsError() {
		return nil, result.Error()
	}

	user := result.MustGet()

	return &userv1.GetUserResponse{
		UserId: user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Bio:    user.Bio,
		Status: string(user.Status),
	}, nil
}
