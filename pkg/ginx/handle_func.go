package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func WrapReq[T any](fn func(ctx *gin.Context, req T /*, uc jwt.UserClaims*/) (Result, error)) func(*gin.Context) {
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

//
//type Result struct {
//	Code int    `json:"code"` // 业务错误码
//	Msg  string `json:"msg"`
//	Data any    `json:"data"`
//}
