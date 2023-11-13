package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/rui-cs/webook/config"
	"github.com/rui-cs/webook/ioc"
	"github.com/rui-cs/webook/pkg/ginx"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

func main() {
	initViperV1()

	initLogger()

	ginx.L = ioc.InitLogger()

	app := InitWebServer()
	// Consumer 在我设计下，类似于 Web，或者 GRPC 之类的，是一个顶级入口
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	server := app.web
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "你好，你来了")
	})

	err := server.Run(fmt.Sprintf(":%s", config.Config.ServerPort))

	if err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	zap.L().Info("before replace")
	// 不用replace，直接用zap.L()，什么都打不出来
	zap.ReplaceGlobals(logger)
	zap.L().Info("hello,ok!")

	type Demo struct {
		Name string `json:"name"`
	}

	zap.L().Info("这是实验参数",
		zap.Error(errors.New("this is an error")),
		zap.Int64("id", 123),
		zap.Any("a struct", Demo{Name: "hello"}))
}

func initViperV1() {
	cfile := pflag.String("config",
		"config/local.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	// 实时监听配置变更
	viper.WatchConfig()
	// 只能告诉你文件变了，不能告诉你，文件的哪些内容变了
	viper.OnConfigChange(func(in fsnotify.Event) {
		// 比较好的设计，它会在 in 里面告诉你变更前的数据，和变更后的数据
		// 更好的设计是，它会直接告诉你差异。
		fmt.Println(in.Name, in.Op)
		fmt.Println(viper.GetString("db.dsn"))
	})
	//viper.SetDefault("db.mysql.dsn",
	//	"root:root@tcp(localhost:3306)/mysql")
	//viper.SetConfigFile("config/dev.yaml")
	//viper.KeyDelimiter("-")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViperReader() {
	viper.SetConfigType("yaml")
	cfg := `
db.mysql:
  dsn: "root:root@tcp(localhost:13316)/webook"

redis:
  addr: "localhost:6379"
`
	err := viper.ReadConfig(bytes.NewReader([]byte(cfg)))
	if err != nil {
		panic(err)
	}
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3",
		// 通过 webook 和其他使用 etcd 的区别出来
		"http://127.0.0.1:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	err = viper.WatchRemoteConfig()
	if err != nil {
		panic(err)
	}
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println(in.Name, in.Op)
	})
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	viper.SetDefault("db.mysql.dsn",
		"root:root@tcp(localhost:3306)/mysql")
	// 配置文件的名字，但是不包含文件扩展名
	// 不包含 .go, .yaml 之类的后缀
	viper.SetConfigName("dev")
	// 告诉 viper 我的配置用的是 yaml 格式
	// 现实中，有很多格式，JSON，XML，YAML，TOML，ini
	viper.SetConfigType("yaml")
	// 当前工作目录下的 config 子目录
	viper.AddConfigPath("./config")
	//viper.AddConfigPath("/tmp/config")
	//viper.AddConfigPath("/etc/webook")
	// 读取配置到 viper 里面，或者你可以理解为加载到内存里面
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	//otherViper := viper.New()
	//otherViper.SetConfigName("myjson")
	//otherViper.AddConfigPath("./config")
	//otherViper.SetConfigType("json")
}
