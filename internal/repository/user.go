package repository

import (
	"context"

	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository/dao"
)

type UserRepository struct {
	dao *dao.UserDAO
}

func (ur *UserRepository) Create(ctx context.Context, u domain.User) error {
	return ur.dao.Insert(ctx, dao.User{Email: u.Email, Password: u.Password})
}
