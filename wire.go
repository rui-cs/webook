//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/rui-cs/webook/internal/repository"
	"github.com/rui-cs/webook/internal/repository/cache"
	"github.com/rui-cs/webook/internal/repository/dao"
	"github.com/rui-cs/webook/internal/service"
	"github.com/rui-cs/webook/internal/web"
	"github.com/rui-cs/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,

		dao.NewUserDAO,

		cache.NewUserCache,
		cache.NewRedisCodeCache,
		//cache.NewRedisEncryptCodeCache,
		//cache.NewMemoryCodeCache,ioc.InitCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		service.NewUserService,
		//service.NewCodeService,
		service.NewFixedCodeService,

		ioc.InitSMSService,
		web.NewUserHandler,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)

	return new(gin.Engine)
}
