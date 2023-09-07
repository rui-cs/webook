package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository/cache"
	cachemocks "github.com/rui-cs/webook/internal/repository/cache/mocks"
	"github.com/rui-cs/webook/internal/repository/dao"
	daomocks "github.com/rui-cs/webook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserRepositoryWithCache_FindById(t *testing.T) {
	now := time.Now()
	now = time.UnixMilli(now.UnixMilli()) //去掉毫秒以外的部分

	testCase := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		ctx      context.Context
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中，查询成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, cache.ErrKeyNotExist)

				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindByID(gomock.Any(), int64(123)).Return(dao.User{
					Id:       123,
					Email:    sql.NullString{String: "123@qq.com", Valid: true},
					Phone:    sql.NullString{String: "12323232312", Valid: true},
					Password: "password",
					Ctime:    now.UnixMilli(),
					Utime:    now.UnixMilli(),
				}, nil)

				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Phone:    "12323232312",
					Password: "password",
					Ctime:    now,
				}).Return(nil)

				return d, c
			},
			ctx: context.Background(),
			id:  123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Phone:    "12323232312",
				Password: "password",
				Ctime:    now,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ud, uc := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			u, err := repo.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
			time.Sleep(time.Second) // cache 异步 set
		})
	}
}
