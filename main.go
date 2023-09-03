package main

import (
	"fmt"

	"github.com/rui-cs/webook/config"
)

func main() {
	server := InitWebServer()

	err := server.Run(fmt.Sprintf(":%s", config.Config.ServerPort))
	if err != nil {
		panic(err)
	}
}
