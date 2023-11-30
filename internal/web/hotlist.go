package web

import (
	"github.com/gin-gonic/gin"
	"github.com/rui-cs/webook/internal/service"
	ijwt "github.com/rui-cs/webook/internal/web/jwt"
	"github.com/rui-cs/webook/pkg/ginx"
)

var _ handler = (*HotListHandler)(nil)

type HotListHandler struct {
	repo service.HotListService
}

func NewHotListHandler(repo service.HotListService) *HotListHandler {
	return &HotListHandler{repo: repo}
}

func (h *HotListHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/hotlist")

	g.POST("/liketopn", ginx.WrapBodyAndToken[LikeTopNReq, ijwt.UserClaims](h.GetLikeTopN))
}

type LikeTopNReq struct {
	Bizs []string `json:"bizs"`
}

func (h *HotListHandler) GetLikeTopN(ctx *gin.Context, req LikeTopNReq, uc ijwt.UserClaims) (ginx.Result, error) {
	res, err := h.repo.GetLikeTopN(req.Bizs)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{Data: res}, nil
}
