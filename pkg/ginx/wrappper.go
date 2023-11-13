package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rui-cs/webook/pkg/logger"
)

// 这个东西，放到你们的 ginx 插件库里面去
// 技术含量不是很高，但是绝对有技巧

// L 使用包变量
var L logger.LoggerV1

//func NewWrapperLogger(l logger.LoggerV1) {
//	L = l
//}

func WrapToken[C jwt.Claims](fn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 执行一些东西
		val, ok := ctx.Get("users")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 下半段的业务逻辑从哪里来？
		// 我的业务逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, c)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
		// 再执行一些东西
	}
}

func WrapBodyAndToken[Req any, C jwt.Claims](fn func(ctx *gin.Context, req Req, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			return
		}

		val, ok := ctx.Get("users")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 下半段的业务逻辑从哪里来？
		// 我的业务逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, req, c)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyV1[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		// 下半段的业务逻辑从哪里来？
		// 我的业务逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, req)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[T any](l logger.LoggerV1, fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		// 下半段的业务逻辑从哪里来？
		// 我的业务逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, req)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			l.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}
