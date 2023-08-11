package service

import (
	"context"

	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {
	return svc.repo.Create(ctx, u)
}
