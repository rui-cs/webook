package ioc

import (
	"time"

	"github.com/rui-cs/webook/config"
	"github.com/rui-cs/webook/internal/repository/dao"
	"github.com/rui-cs/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	//dsn := fmt.Sprintf("root:%s@tcp(%s:%s)/webook?charset=utf8&parseTime=True&loc=Local",
	//	config.Config.DCfg.Pass, config.Config.DCfg.Addr, config.Config.DCfg.Port)

	type Config struct {
		DSN string `yaml:"dsn"`

		// 有些人的做法
		// localhost:13316
		//Addr string
		//// localhost
		//Domain string
		//// 13316
		//Port string
		//Protocol string
		//// root
		//Username string
		//// root
		//Password string
		//// webook
		//DBName string
	}
	var cfg = Config{
		DSN: "root:ttdysmmfsl@tcp(localhost:3306)/webook?charset=utf8&parseTime=True&loc=Local",
	}
	// 看起来，remote 不支持 key 的切割
	err := viper.UnmarshalKey("db", &cfg)

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			//	慢查询阈值，只有执行时间超过这个阈值，才会使用
			//	SQL 查询必然要求命中索引，最好就是走一次磁盘 IO
			//	一次磁盘IO是不到10ms
			SlowThreshold:             time.Millisecond * 10,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  glogger.Info,
		}),
	})
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

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}

type DoSomething interface {
	DoABC() string
}

type DoSomethingFunc func() string

func (d DoSomethingFunc) DoABC() string {
	return d()
}
