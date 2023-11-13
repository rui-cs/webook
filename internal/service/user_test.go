package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository"
	repomocks "github.com/rui-cs/webook/internal/repository/mocks"
	"github.com/rui-cs/webook/pkg/logger"
	"go.uber.org/mock/gomock"
)

func TestUserService_Login(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		email    string
		password string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{
					Email:    "123@qq.com",
					Password: "$2a$10$MN9ZKKIbjLZDyEpCYW19auY7mvOG9pcpiIcUUoZZI6pA6OmKZKOVi",
					Phone:    "15612322323",
					Ctime:    now,
				}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",
			wantUser: domain.User{
				Email:    "123@qq.com",
				Phone:    "15612322323",
				Password: "$2a$10$MN9ZKKIbjLZDyEpCYW19auY7mvOG9pcpiIcUUoZZI6pA6OmKZKOVi",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "123@qq.com",
			password: "hello#world123",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, errors.New("mock db error"))
				return repo
			},
			email:    "123@qq.com",
			password: "",
			wantUser: domain.User{},
			wantErr:  errors.New("mock db error"),
		},
		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{
					Email:    "123@qq.com",
					Password: "$2a$10$MN9ZKKIbjLZDyEpCYW19auY7mvOG9pcpiIcUUoZZI6pA6OmKZKOVi",
					Phone:    "12321232323",
					Ctime:    now,
				}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl), &logger.NopLogger{})
			u, err := svc.Login(context.Background(), tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}
