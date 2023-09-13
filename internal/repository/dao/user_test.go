package dao

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/assert/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGormUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(t *testing.T) *sql.DB
		ctx     context.Context
		user    User
		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				res := sqlmock.NewResult(3, 1)
				// 正则表达式 : 只要是 INSERT 到 users 的语句
				mock.ExpectExec("INSERT INTO `users`.*").WillReturnResult(res)
				require.NoError(t, err)
				return mockDB
			},
			user: User{Email: sql.NullString{String: "123@qq.com", Valid: true}},
		},
		{
			name: "邮箱冲突",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				mock.ExpectExec("INSERT INTO `users`.*").WillReturnError(&mysql.MySQLError{Number: 1062})
				require.NoError(t, err)
				return mockDB
			},
			user:    User{},
			wantErr: ErrUserDuplicateEmail,
		},
		{
			name: "数据库错误",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				mock.ExpectExec("INSERT INTO `users`.*").WillReturnError(errors.New("数据库错误"))
				require.NoError(t, err)
				return mockDB
			},
			user:    User{},
			wantErr: errors.New("数据库错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{Conn: tc.mock(t), SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true})
			assert.Equal(t, err, nil)
			d := NewUserDAO(db)
			u := tc.user
			err = d.Insert(tc.ctx, u)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
