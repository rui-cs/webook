package repository

import (
	"context"

	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
	ErrUserDuplicateName  = dao.ErrUserDuplicateName
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(d *dao.UserDAO) *UserRepository {
	return &UserRepository{dao: d}
}

func (ur *UserRepository) Create(ctx context.Context, u domain.User) error {
	return ur.dao.Insert(ctx, dao.User{Email: u.Email, Password: u.Password})
}

func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := ur.dao.FindByEmail(ctx, email)
	//if err == dao.ErrUserNotFound { 无需判断，上层直接判断
	//	return domain.User{}, ErrUserNotFound
	//}
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{Id: user.Id, Email: user.Email, Password: user.Password}, nil
}

func (ur *UserRepository) EditByID(ctx context.Context, id int, name, birthday, resume string) error {
	return ur.dao.EditByID(ctx, id, name, birthday, resume)
}

func (ur *UserRepository) Profile(ctx context.Context, id int) (domain.User, error) {
	user, err := ur.dao.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Name:     user.Name,
		Birthday: user.Birthday,
		Resume:   user.Resume,
	}, nil
}
