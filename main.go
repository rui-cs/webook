package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/rui-cs/webook/internal/repository"
	"github.com/rui-cs/webook/internal/repository/dao"
	"github.com/rui-cs/webook/internal/service"
	"github.com/rui-cs/webook/internal/web"
	"github.com/rui-cs/webook/internal/web/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	server := initWebServer()

	initUser(server, db)

	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	// todo 添加跨域中间件

	// 中间件验证session
	//c := cookie.NewStore([]byte("secret")) // cookie-based
	c, _ := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"), []byte("0Pf2r0wZBpXVXlQNdpwCXN4ncnlnZSc3"))
	server.Use(sessions.Sessions("ssid", c)) // 提取session
	l := &middleware.LoginMiddlewareBuilder{}
	server.Use(l.CheckLogin()) // 执行登录校验

	return server
}

func initUser(server *gin.Engine, db *gorm.DB) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	c := web.NewUserHandler(us)
	c.RegisterRoutes(server)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:your_password@tcp(localhost:3306)/webook?charset=utf8&parseTime=True&loc=Local"))
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}

	return db
}
