package web

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ObservabilityHandler struct {
}

func (h *ObservabilityHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("test")
	g.GET("/metric", func(ctx *gin.Context) {
		sleep := rand.Int31n(1000)
		time.Sleep(time.Millisecond * time.Duration(sleep))
		ctx.String(http.StatusOK, "OK")
	})
}
