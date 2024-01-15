//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/rui-cs/webook/internal/events/article"
	"github.com/rui-cs/webook/internal/repository"
	articleRepo "github.com/rui-cs/webook/internal/repository/article"
	"github.com/rui-cs/webook/internal/repository/cache"
	"github.com/rui-cs/webook/internal/repository/dao"
	articleDao "github.com/rui-cs/webook/internal/repository/dao/article"
	"github.com/rui-cs/webook/internal/service"
	"github.com/rui-cs/webook/internal/web"
	ijwt "github.com/rui-cs/webook/internal/web/jwt"
	"github.com/rui-cs/webook/ioc"
)

//func InitWebServer() *gin.Engine {
//	wire.Build(
//		ioc.InitDB, ioc.InitRedis,
//		ioc.InitLogger,
//		//logger.NewZapLogger,
//
//		dao.NewUserDAO,
//		dao.NewGORMArticleDAO,
//
//		cache.NewUserCache,
//		cache.NewRedisCodeCache,
//		//cache.NewRedisEncryptCodeCache,
//		//cache.NewMemoryCodeCache,ioc.InitCache,
//
//		repository.NewUserRepository,
//		repository.NewCodeRepository,
//		article.NewArticleRepository,
//
//		service.NewUserService,
//		//service.NewCodeService,
//		service.NewFixedCodeService,
//		service.NewArticleService,
//
//		ioc.InitSMSService,
//		ioc.InitWechatService,
//
//		web.NewUserHandler,
//		web.NewOAuth2WechatHandler,
//		web.NewArticleHandler,
//		//ioc.NewWechatHandlerConfig,
//		ijwt.NewRedisJWTHandler,
//
//		ioc.InitWebServer,
//		ioc.InitMiddlewares,
//	)
//
//	return new(gin.Engine)
//}

//func InitWebServer() *gin.Engine {
//	wire.Build(
//		// 最基础的第三方依赖
//		ioc.InitDB, ioc.InitRedis,
//		ioc.InitLogger,
//
//		// 初始化 DAO
//		dao.NewUserDAO,
//
//		cache.NewUserCache,
//		cache.NewRedisCodeCache,
//
//		repository.NewUserRepository,
//		repository.NewCodeRepository,
//
//		service.NewUserService,
//		service.NewCodeService,
//		// 直接基于内存实现
//		ioc.InitSMSService,
//		ioc.InitWechatService,
//
//		web.NewUserHandler,
//		web.NewOAuth2WechatHandler,
//		//ioc.NewWechatHandlerConfig,
//		ijwt.NewRedisJWTHandler,
//		// 你中间件呢？
//		// 你注册路由呢？
//		// 你这个地方没有用到前面的任何东西
//		//gin.Default,
//
//		ioc.InitWebServer,
//		ioc.InitMiddlewares,
//
//		web.NewArticleHandler,
//		service.NewArticleServiceV1,
//		article.NewArticleRepository,
//		dao.NewGORMArticleDAO,
//	)
//	return new(gin.Engine)
//}

func InitWebServer() *App {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.InitCache,
		//ioc.NewSyncProducer,

		// consumer
		article.NewInteractiveReadEventBatchConsumer,
		//article.NewKafkaProducer,

		// 初始化 DAO
		dao.NewUserDAO,
		articleDao.NewGORMArticleDAO,
		dao.NewGORMInteractiveDAO,
		dao.NewHotListDao,

		cache.NewRedisInteractiveCache,
		cache.NewInteractiveCacheHotList,
		cache.NewUserCache,
		//cache.NewCodeCache,
		cache.NewRedisCodeCache,
		cache.NewRedisArticleCache,
		cache.NewRedisHotListCache,
		cache.NewLocalHotListCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedInteractiveRepository,
		articleRepo.NewArticleRepository,
		//repository.NewHotListRepo,
		ioc.InitHotListRepo,

		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		service.NewInteractiveService,
		service.NewHotListService,

		// 直接基于内存实现
		ioc.InitSMSService,
		ioc.InitWechatService,

		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		web.NewHotListHandler,
		//ioc.NewWechatHandlerConfig,
		ijwt.NewRedisHandler,
		// 你中间件呢？
		// 你注册路由呢？
		// 你这个地方没有用到前面的任何东西
		//gin.Default,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
		// 组装我这个结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
