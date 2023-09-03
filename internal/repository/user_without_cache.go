package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository/dao"
)

type UserRepositoryWithoutCache struct {
	dao dao.UserDAO
}

func NewUserRepositoryWithoutCache(d dao.UserDAO) UserRepository {
	return &UserRepositoryWithoutCache{dao: d}
}

func (ur *UserRepositoryWithoutCache) Create(ctx context.Context, u domain.User) error {
	return ur.dao.Insert(ctx, dao.User{Email: sql.NullString{String: u.Email, Valid: true}, Password: u.Password})
}

func (ur *UserRepositoryWithoutCache) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := ur.dao.FindByEmail(ctx, email)
	//if err == dao.ErrUserNotFound { 无需判断，上层直接判断
	//	return domain.User{}, ErrUserNotFound
	//}
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{Id: user.Id, Email: user.Email.String, Password: user.Password}, nil
}

func (ur *UserRepositoryWithoutCache) EditByID(ctx context.Context, id int64, name, birthday, resume string) error {
	return ur.dao.EditByID(ctx, id, name, birthday, resume)
}

func (ur *UserRepositoryWithoutCache) FindById(ctx context.Context, id int64) (domain.User, error) {
	user, err := ur.dao.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u := ur.entityToDomain(user)

	return u, nil
}

func (ur *UserRepositoryWithoutCache) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := ur.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}

	return ur.entityToDomain(u), err
}

func (ur *UserRepositoryWithoutCache) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Name:     u.Name.String,
		Password: u.Password,
		Birthday: u.Birthday,
		Resume:   u.Resume,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
