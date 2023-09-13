package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rui-cs/webook/internal/repository/cache/redismocks"
	"go.uber.org/mock/gomock"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		ctx     context.Context
		biz     string
		phone   string
		code    string
		wantErr error
	}{
		{
			name: "验证码设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:152"}, []any{"123456"}).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "152",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "redis错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redis.NewCmd(context.Background())
				res.SetErr(errors.New("mock redis error"))
				cmd := redismocks.NewMockCmdable(ctrl)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{"phone_code:login:152"},
					[]any{"123456"},
				).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "152",
			code:    "123456",
			wantErr: errors.New("mock redis error"),
		},
		{
			name: "发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-1))
				cmd := redismocks.NewMockCmdable(ctrl)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:152"}, []any{"123456"}).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "152",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-10))
				cmd := redismocks.NewMockCmdable(ctrl)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:152"}, []any{"123456"}).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "152",
			code:    "123456",
			wantErr: errors.New("系统错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := NewRedisCodeCache(tc.mock(ctrl))
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
