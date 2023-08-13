package service

import (
	"context"
	"errors"

	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrUserDuplicateName     = repository.ErrUserDuplicateName
	ErrInvalidUserOrPassword = errors.New("账号或密码错误")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hash)

	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	user, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword // 模糊错误信息
	}
	if err != nil {
		return domain.User{}, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// todo 后续添加日志
		return domain.User{}, ErrInvalidUserOrPassword
	}

	return user, nil
}
func (svc *UserService) Edit(ctx context.Context, id int, name string, birthday WebookTime, resume string) error {
	tmp, _ := birthday.MarshalJSON()

	return svc.repo.EditByID(ctx, id, name, string(tmp), resume)
}

func (svc *UserService) Profile(ctx context.Context, id int) (domain.User, error) {
	return svc.repo.Profile(ctx, id)
}
