package main

import (
	"errors"
	"fmt"

	"github.com/rui-cs/webook/config"
	"go.uber.org/zap"
)

func main() {
	initLogger()

	server := InitWebServer()

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
