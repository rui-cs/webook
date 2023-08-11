package dao

import (
	"context"

	"gorm.io/gorm"
)

type UserDAO struct {
	db *gorm.DB
}

func (ud *UserDAO) Insert(ctx context.Context, u User) error {
	return ud.db.WithContext(ctx).Create(&u).Error
}

// User 直接对应数据库表结构
// 有些人叫做 entity，有些人叫做 model，有些人叫做 PO(persistent object)
type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Email    string `gorm:"unique"` // 唯一索引 全部用户唯一
	Password string
	Ctime    int64 // 创建时间，毫秒数
	Utime    int64 // 更新时间，毫秒数
}
