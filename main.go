package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rui-cs/webook/internal/web"
)

func main() {
	server := initWebServer()

	user := web.NewUserHandler()
	user.RegisterRoutes(server)

	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	return server
}
