package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/rui-cs/webook/config"
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
	addCORSMiddleware(server)

	var loginStatusType = map[int]func(engine *gin.Engine){
		config.CheckSession: addCheckSessionMiddleware,
		config.JWT:          addJWTMiddleware,
	}
	loginStatusType[config.Config.LoginCheckType](server)

	initUser(server, db)

	err := server.Run(fmt.Sprintf(":%s", config.Config.ServerPort))
	if err != nil {
		panic(err)
	}
}

func initDB() *gorm.DB {
	dsn := fmt.Sprintf("root:%s@tcp(%s:%s)/webook?charset=utf8&parseTime=True&loc=Local",
		config.Config.DCfg.Pass, config.Config.DCfg.Addr, config.Config.DCfg.Port)

	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}

	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	return server
}

func addCORSMiddleware(server *gin.Engine) {
	server.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost")
		},
		AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"x-jwt-token"},
		MaxAge:           12 * time.Hour,
	}))
}

// 登录态验证方式一 : session
func addCheckSessionMiddleware(server *gin.Engine) {
	// 中间件验证session
	//c := cookie.NewStore([]byte("secret")) // cookie-based
	c, _ := redis.NewStore(10, "tcp", fmt.Sprintf("%s:%s", config.Config.RCg.Addr, config.Config.RCg.Port), "", []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"), []byte("0Pf2r0wZBpXVXlQNdpwCXN4ncnlnZSc3"))
	server.Use(sessions.Sessions("ssid", c)) // 提取session
	l := &middleware.LoginMiddlewareBuilder{}
	server.Use(l.CheckLogin()) // 执行登录校验
}

// 登录态验证方式二 : JWT
func addJWTMiddleware(server *gin.Engine) {
	l := middleware.NewLoginJWTMiddlewareBuilder()
	server.Use(l.IgnorePath("/users/login").IgnorePath("/users/signup").Build())
}

func initUser(server *gin.Engine, db *gorm.DB) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	c := web.NewUserHandler(us)
	c.RegisterRoutes(server)
}
