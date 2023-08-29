package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository/cache"
	"github.com/rui-cs/webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
	ErrUserDuplicateName  = dao.ErrUserDuplicateName
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	//FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	EditByID(ctx context.Context, id int, name, birthday, resume string) error
}

type UserRepositoryWithCache struct {
	dao   *dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(d *dao.UserDAO, c cache.UserCache) UserRepository {
	return &UserRepositoryWithCache{dao: d, cache: c}
}

func (ur *UserRepositoryWithCache) Create(ctx context.Context, u domain.User) error {
	return ur.dao.Insert(ctx, dao.User{Email: sql.NullString{String: u.Email, Valid: true}, Password: u.Password})
}

func (ur *UserRepositoryWithCache) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := ur.dao.FindByEmail(ctx, email)
	//if err == dao.ErrUserNotFound { 无需判断，上层直接判断
	//	return domain.User{}, ErrUserNotFound
	//}
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{Id: user.Id, Email: user.Email.String, Password: user.Password}, nil
}

func (ur *UserRepositoryWithCache) EditByID(ctx context.Context, id int, name, birthday, resume string) error {
	return ur.dao.EditByID(ctx, id, name, birthday, resume)
}

func (ur *UserRepositoryWithCache) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := ur.cache.Get(ctx, id)
	if err == nil { // 缓存中找到了
		return u, nil
	}

	user, err := ur.dao.FindByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = ur.entityToDomain(user)

	go func() {
		err = ur.cache.Set(ctx, u)
		if err != nil {
			fmt.Println("ur.cache.Set error : ", err)
		}
	}()

	return u, nil
}

func (ur *UserRepositoryWithCache) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Name:     u.Name.String,
		Password: u.Password,
		Birthday: u.Birthday,
		Resume:   u.Resume,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}
