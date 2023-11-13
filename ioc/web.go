package ioc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	redisSession "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rui-cs/webook/config"
	"github.com/rui-cs/webook/internal/web"
	ijwt "github.com/rui-cs/webook/internal/web/jwt"
	"github.com/rui-cs/webook/internal/web/middleware"
	midLogger "github.com/rui-cs/webook/pkg/ginx/middlewares/logger"
	"github.com/rui-cs/webook/pkg/logger"
)

func InitWebServer(middleHdls []gin.HandlerFunc,
	userHdl *web.UserHandler,
	oauth2WechatHdl *web.OAuth2WechatHandler,
	articleHdl *web.ArticleHandler) *gin.Engine {

	server := gin.Default()

	//gin.SetMode(gin.ReleaseMode)

	server.Use(middleHdls...)
	userHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)

	return server
}

func InitMiddlewares(client redis.Cmdable, jwtHdl ijwt.Handler, l logger.LoggerV1) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHandler(),

		midLogger.NewBuilder(func(ctx context.Context, al *midLogger.AccessLog) {
			l.Debug("HTTP请求", logger.Field{Key: "al", Value: al})
		}).AllowReqBody(true).AllowRespBody().Build(),

		middleware.NewJWTLoginMiddlewareBuilder(jwtHdl).Build(),
		//IgnorePath("/users/refresh_token").
		//IgnorePath("/users/login").
		//IgnorePath("/users/login_sms/code/send").
		//IgnorePath("/users/login_sms").
		//IgnorePath("/oauth2/wechat/authurl").
		//IgnorePath("/oauth2/wechat/callback").
		//IgnorePath("/users/signup").Build(),

		//ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHandler() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost")
		},
		AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
		MaxAge:           12 * time.Hour,
	})
}

// 登录态验证方式一 : session
func addCheckSessionMiddleware(server *gin.Engine) {
	// 中间件验证session
	//c := cookie.NewStore([]byte("secret")) // cookie-based
	c, _ := redisSession.NewStore(10, "tcp", fmt.Sprintf("%s:%s", config.Config.RCg.Addr, config.Config.RCg.Port), "", []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"), []byte("0Pf2r0wZBpXVXlQNdpwCXN4ncnlnZSc3"))
	server.Use(sessions.Sessions("ssid", c)) // 提取session
	l := &middleware.LoginMiddlewareBuilder{}
	server.Use(l.CheckLogin()) // 执行登录校验
}

// 登录态验证方式二 : JWT
//func addJWTMiddleware(server *gin.Engine) {
//	l := middleware.NewLoginJWTMiddlewareBuilder()
//	server.Use(middleware.NewLoginJWTMiddlewareBuilder().IgnorePath("/users/login").IgnorePath("/users/signup").Build())
//}
