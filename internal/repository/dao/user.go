package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserDuplicateName  = errors.New("用户名冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (ud *UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli() // 存毫秒数
	u.Utime = now
	u.Ctime = now

	err := ud.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			return ErrUserDuplicateEmail // 邮箱冲突
		}
	}

	return err
}

func (ud *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("email = ?", email).Find(&u).Error
	return u, err
}

func (ud *UserDAO) EditByID(ctx context.Context, id int, name, birthday, resume string) error {
	err := ud.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(User{Name: sql.NullString{String: name, Valid: true}, Birthday: birthday, Resume: resume}).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			return ErrUserDuplicateName
		}
	}

	return err
}

func (ud *UserDAO) FindByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("id = ?", id).Find(&u).Error
	return u, err
}

// User 直接对应数据库表结构
// 有些人叫做 entity，有些人叫做 model，有些人叫做 PO(persistent object)
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`

	Email    sql.NullString `gorm:"unique"` // 唯一索引 全部用户唯一
	Password string
	Name     sql.NullString `gorm:"unique"` // 唯一索引 全部用户唯一
	Birthday string
	Resume   string `gorm:"type:text"`

	Ctime int64 // 创建时间，毫秒数
	Utime int64 // 更新时间，毫秒数
}
