package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.GET("/profile", u.Profile)
	ug.POST("/edit", u.Edit)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	ctx.String(http.StatusOK, "signup")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	ctx.String(http.StatusOK, "login")
}

func (u *UserHandler) Profile(ctx *gin.Context) {
	ctx.String(http.StatusOK, "profile")
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	ctx.String(http.StatusOK, "edit")
}
