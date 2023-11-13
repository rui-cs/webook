package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func WrapReq[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			zap.L().Error(res.Msg, zap.Error(err))
		}

		ctx.JSON(http.StatusOK, res)
	}
}

func WrapToken[C jwt.Claims](fn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		res, err := fn(ctx, c)
		if err != nil {
			zap.L().Error(res.Msg, zap.Error(err),
				zap.String("path", ctx.Request.URL.Path),
				zap.String("route", ctx.FullPath()))
		}

		ctx.JSON(http.StatusOK, res)
	}
}

func WrapReqAndToken[T any, C jwt.Claims](fn func(ctx *gin.Context, req T, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		val, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		res, err := fn(ctx, req, c)
		if err != nil {
			zap.L().Error(res.Msg, zap.Error(err),
				zap.String("path", ctx.Request.URL.Path),
				zap.String("route", ctx.FullPath()))
		}

		ctx.JSON(http.StatusOK, res)
	}
}

type Result struct {
	Code int    `json:"code"` // 业务错误码
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
