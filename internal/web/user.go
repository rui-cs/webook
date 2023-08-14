package web

import (
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/service"
)

const (
	emailRegex = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$" // todo 正则表达式
	passRegex  = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct {
	svc      *service.UserService
	emailExp *regexp.Regexp
	passExp  *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	emailExp := regexp.MustCompile(emailRegex, regexp.None)
	passExp := regexp.MustCompile(passRegex, regexp.None)

	return &UserHandler{
		svc:      svc,
		emailExp: emailExp,
		passExp:  passExp,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	ug.POST("/signup", u.SignUp)
	//ug.POST("/login", u.Login)
	ug.POST("/login", u.LoginJWT)
	//ug.GET("/profile", u.Profile)
	ug.GET("/profile", u.ProfileJWT)
	//ug.POST("/edit", u.Edit)
	ug.POST("/edit", u.EditJWT)
	ug.POST("/logout", u.Logout)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email             string `json:"email"`
		Password          string `json:"password"`
		ConfirmedPassword string `json:"confirmedPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !ok {
		ctx.String(http.StatusOK, "邮箱格式错误")
		return
	}

	if req.Password != req.ConfirmedPassword {
		ctx.String(http.StatusOK, "两次输入密码不一致")
		return
	}

	ok, err = u.passExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !ok {
		ctx.String(http.StatusOK, "密码格式错误，密码必须大于8位，包含数字、特殊字符")
		return
	}

	if err = u.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	}); err != nil {
		if err == service.ErrUserDuplicateEmail {
			ctx.String(http.StatusOK, "邮箱冲突")
			return
		}
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	s := sessions.Default(ctx)
	s.Set(userID, int(user.Id))
	s.Options(sessions.Options{MaxAge: 60})

	if err = s.Save(); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))},
		Uid:              user.Id,
		UserAgent:        ctx.Request.UserAgent(),
	})
	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0")) // todo
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	ctx.Header("x-jwt-token", tokenStr)
	ctx.String(http.StatusOK, "登录成功")
}

// 放入token的数据
type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}

const userID = "userID"

func (u *UserHandler) Profile(ctx *gin.Context) {
	s := sessions.Default(ctx)
	id := s.Get(userID)
	_, ok := id.(int)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	user, err := u.svc.Profile(ctx, id.(int))
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, ok := ctx.Get("claims")
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	user, err := u.svc.Profile(ctx, int(claims.Uid))
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Name     string             `json:"name"`
		Birthday service.WebookTime `json:"birthday"`
		Resume   string             `json:"resume"`
	}

	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	s := sessions.Default(ctx)
	id := s.Get(userID)
	_, ok := id.(int)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	err := u.svc.Edit(ctx, id.(int), req.Name, req.Birthday, req.Resume)
	if err == service.ErrUserDuplicateName {
		ctx.String(http.StatusOK, "用户名重复")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "修改成功")
}

func (u *UserHandler) EditJWT(ctx *gin.Context) {
	type EditReq struct {
		Name     string             `json:"name"`
		Birthday service.WebookTime `json:"birthday"`
		Resume   string             `json:"resume"`
	}

	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	c, ok := ctx.Get("claims")
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	err := u.svc.Edit(ctx, int(claims.Uid), req.Name, req.Birthday, req.Resume)
	if err == service.ErrUserDuplicateName {
		ctx.String(http.StatusOK, "用户名重复")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "修改成功")
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	s := sessions.Default(ctx)
	s.Options(sessions.Options{MaxAge: -1})

	err := s.Save()
	if err != nil {
		ctx.String(http.StatusOK, "登出失败")
		return
	}

	ctx.String(http.StatusOK, "登出成功")
}
