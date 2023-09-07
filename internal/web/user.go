package web

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rui-cs/webook/config"
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/service"
)

const (
	emailRegex = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$" // todo 正则表达式
	passRegex  = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct {
	svc      service.UserService
	codeSvc  service.CodeService
	emailExp *regexp.Regexp
	passExp  *regexp.Regexp
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	emailExp := regexp.MustCompile(emailRegex, regexp.None)
	passExp := regexp.MustCompile(passRegex, regexp.None)

	return &UserHandler{
		svc:      svc,
		codeSvc:  codeSvc,
		emailExp: emailExp,
		passExp:  passExp,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")

	ug.POST("/signup", u.SignUp)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginSMS)

	var routerGroup = map[int]func(){
		config.CheckSession: func() {
			ug.POST("/login", u.Login)
			ug.GET("/profile", u.Profile)
			ug.POST("/edit", u.Edit)
			ug.POST("/logout", u.Logout)
		},
		config.JWT: func() {
			ug.POST("/login", u.LoginJWT)
			ug.GET("/profile", u.ProfileJWT)
			ug.POST("/edit", u.EditJWT)
			ug.POST("/logout", u.LogoutJWT)
		},
	}

	routerGroup[config.Config.LoginCheckType]()
}

const biz = "login"

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "输入错误"})
		return
	}

	err := u.codeSvc.Send(ctx, biz, req.Phone)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, Result{Msg: "发送成功"})
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, Result{Msg: "发送太频繁，请稍后再试"})
	case errors.Is(err, service.ErrCodeOperationTooMany):
		fmt.Println("----------------------------------操作太频繁，请稍后再试---------------------------")
		ctx.JSON(http.StatusOK, Result{Msg: "操作太频繁，请稍后再试"})
	default:
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
	}
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)

	if errors.Is(err, service.ErrCodeOperationTooMany) {
		fmt.Println("----------------------------------操作太频繁，请稍后再试---------------------------")
		ctx.JSON(http.StatusOK, Result{Msg: "操作太频繁，请稍后再试"})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证码有误"})
		return
	}

	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}

	if err = u.setJWTToken(ctx, user.Id); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "登录成功")

	return
}

func (u *UserHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(config.Config.ValidTime)))},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0")) // todo
	if err != nil {
		return err
	}

	ctx.Header("x-jwt-token", tokenStr)

	return nil
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

	if err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	}); err != nil {
		if errors.Is(err, service.ErrUserDuplicateEmail) {
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
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
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
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	err = u.setJWTToken(ctx, user.Id)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

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

	user, err := u.svc.Profile(ctx, int64(id.(int)))
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

	user, err := u.svc.Profile(ctx, claims.Uid)
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

	err := u.svc.Edit(ctx, int64(id.(int)), req.Name, req.Birthday, req.Resume)
	if errors.Is(err, service.ErrUserDuplicateName) {
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

	err := u.svc.Edit(ctx, claims.Uid, req.Name, req.Birthday, req.Resume)
	if errors.Is(err, service.ErrUserDuplicateName) {
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

func (u *UserHandler) LogoutJWT(ctx *gin.Context) {
	// todo
}
