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

type UserDAO interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	Insert(ctx context.Context, u User) error
	EditByID(ctx context.Context, id int64, name, birthday, resume string) error
}

type GormUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDAO{db: db}
}

func (ud *GormUserDAO) Insert(ctx context.Context, u User) error {
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

func (ud *GormUserDAO) EditByID(ctx context.Context, id int64, name, birthday, resume string) error {
	err := ud.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(User{Name: sql.NullString{String: name, Valid: true}, Birthday: birthday, Resume: resume}).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			return ErrUserDuplicateName
		}
	}

	return err
}

func (ud *GormUserDAO) FindByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

func (ud *GormUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("phone=?", phone).First(&u).Error
	return u, err
}

func (ud *GormUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

// User 直接对应数据库表结构
// 有些人叫做 entity，有些人叫做 model，有些人叫做 PO(persistent object)
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`

	Email    sql.NullString `gorm:"unique"` // 唯一索引 全部用户唯一
	Phone    sql.NullString `gorm:"unique"` // 唯一索引 全部用户唯一
	Name     sql.NullString `gorm:"unique"` // 唯一索引 全部用户唯一
	Birthday string
	Resume   string `gorm:"type:text"`
	Password string

	Ctime int64 // 创建时间，毫秒数
	Utime int64 // 更新时间，毫秒数
}
