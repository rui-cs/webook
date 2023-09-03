package ioc

import (
	"fmt"

	"github.com/rui-cs/webook/config"
	"github.com/rui-cs/webook/internal/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf("root:%s@tcp(%s:%s)/webook?charset=utf8&parseTime=True&loc=Local",
		config.Config.DCfg.Pass, config.Config.DCfg.Addr, config.Config.DCfg.Port)

	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(err)
	}

	if config.Config.GormDebug {
		db = db.Debug()
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}

	return db
}
