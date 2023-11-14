package web

import (
	"errors"
	"fmt"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rui-cs/webook/config"
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/service"
	"github.com/rui-cs/webook/pkg/ginx"
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
	ug.POST("/login_sms", ginx.WrapReq[LoginReq](u.LoginSMS))

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
			ug.POST("/edit", ginx.WrapReqAndToken[EditReq, *UserClaims](u.EditJWT))
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

	fn := func(ctx *gin.Context, req Req) (ginx.Result, error) {
		if req.Phone == "" {
			return ginx.Result{Code: 4, Msg: "输入错误"}, nil
		}

		err := u.codeSvc.Send(ctx, biz, req.Phone)
		switch {
		case err == nil:
			return ginx.Result{Msg: "发送成功"}, nil
		case errors.Is(err, service.ErrCodeSendTooMany):
			return ginx.Result{Msg: "发送太频繁，请稍后再试"}, err
		case errors.Is(err, service.ErrCodeOperationTooMany):
			fmt.Println("----------------------------------操作太频繁，请稍后再试---------------------------")
			return ginx.Result{Msg: "操作太频繁，请稍后再试"}, err
		default:
			return ginx.Result{Code: 5, Msg: "系统错误"}, err
		}
	}

	ginx.WrapReq[Req](fn)(ctx)
}

type LoginReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

func (u *UserHandler) LoginSMS(ctx *gin.Context, req LoginReq) (ginx.Result, error) {
	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)

	if errors.Is(err, service.ErrCodeOperationTooMany) {
		fmt.Println("----------------------------------操作太频繁，请稍后再试---------------------------")
		return ginx.Result{Msg: "操作太频繁，请稍后再试"}, err
	}

	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	if !ok {
		return ginx.Result{Code: 4, Msg: "验证码有误"}, err
	}

	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	if err = u.setJWTToken(ctx, user.Id); err != nil {
		return ginx.Result{Msg: "系统错误"}, err
	}

	return ginx.Result{Msg: "登录成功"}, err
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

	fn := func(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {
		ok, err := u.emailExp.MatchString(req.Email)
		if err != nil {
			return ginx.Result{Msg: "系统错误"}, err
		}

		if !ok {
			return ginx.Result{Msg: "邮箱格式错误"}, nil
		}

		if req.Password != req.ConfirmedPassword {
			return ginx.Result{Msg: "两次输入密码不一致"}, nil
		}

		ok, err = u.passExp.MatchString(req.Password)
		if err != nil {
			return ginx.Result{Msg: "系统错误"}, err
		}

		if !ok {
			return ginx.Result{Msg: "密码格式错误，密码必须大于8位，包含数字、特殊字符"}, nil
		}

		if err = u.svc.SignUp(ctx, domain.User{
			Email:    req.Email,
			Password: req.Password,
		}); err != nil {
			if errors.Is(err, service.ErrUserDuplicateEmail) {
				return ginx.Result{Msg: "邮箱冲突"}, err
			}

			return ginx.Result{Msg: "系统错误"}, err
		}

		return ginx.Result{Msg: "注册成功"}, err
	}

	ginx.WrapReq[SignUpReq](fn)(ctx)
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	fn := func(ctx *gin.Context, req LoginReq) (ginx.Result, error) {
		user, err := u.svc.Login(ctx, req.Email, req.Password)
		if errors.Is(err, service.ErrInvalidUserOrPassword) {
			return ginx.Result{Msg: err.Error()}, err
		}

		if err != nil {
			return ginx.Result{Msg: "系统错误"}, err
		}

		s := sessions.Default(ctx)
		s.Set(userID, int(user.Id))
		s.Options(sessions.Options{MaxAge: 60})

		if err = s.Save(); err != nil {
			return ginx.Result{Msg: "系统错误"}, err
		}

		return ginx.Result{Msg: "登录成功"}, nil
	}

	ginx.WrapReq[LoginReq](fn)(ctx)
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	fn := func(ctx *gin.Context, req LoginReq) (ginx.Result, error) {
		user, err := u.svc.Login(ctx, req.Email, req.Password)
		if errors.Is(err, service.ErrInvalidUserOrPassword) {
			return ginx.Result{Msg: err.Error()}, err
		}

		if err != nil {
			return ginx.Result{Msg: "系统错误"}, err
		}

		err = u.setJWTToken(ctx, user.Id)
		if err != nil {
			return ginx.Result{Msg: "系统错误"}, err
		}

		return ginx.Result{Msg: "登录成功"}, nil
	}

	ginx.WrapReq[LoginReq](fn)(ctx)
}

var _ jwt.Claims = (*UserClaims)(nil)

// 放入token的数据
type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}

const userID = "userID"

func (u *UserHandler) Profile(ctx *gin.Context) {
	type EmptyReq struct {
	}

	fn := func(ctx *gin.Context, req EmptyReq) (ginx.Result, error) {
		s := sessions.Default(ctx)
		id := s.Get(userID)
		_, ok := id.(int)
		if !ok {
			return ginx.Result{Msg: "系统错误"}, nil
		}

		user, err := u.svc.Profile(ctx, int64(id.(int)))
		if err != nil {
			return ginx.Result{Msg: "系统错误"}, err
		}

		return ginx.Result{Data: user}, err
	}

	ginx.WrapReq[EmptyReq](fn)(ctx)
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	type EmptyReq struct {
	}

	fn := func(ctx *gin.Context, req EmptyReq) (ginx.Result, error) {
		c, ok := ctx.Get("claims")
		if !ok {
			return ginx.Result{Msg: "系统错误"}, nil
		}

		claims, ok := c.(*UserClaims)
		if !ok {
			return ginx.Result{Msg: "系统错误"}, nil
		}

		user, err := u.svc.Profile(ctx, claims.Uid)
		if err != nil {
			return ginx.Result{Msg: "系统错误"}, err
		}

		return ginx.Result{Data: user}, nil
	}

	ginx.WrapReq[EmptyReq](fn)(ctx)
}

type EditReq struct {
	Name     string             `json:"name"`
	Birthday service.WebookTime `json:"birthday"`
	Resume   string             `json:"resume"`
}

func (u *UserHandler) Edit(ctx *gin.Context) {

	fn := func(ctx *gin.Context, req EditReq) (ginx.Result, error) {
		s := sessions.Default(ctx)
		id := s.Get(userID)
		_, ok := id.(int)
		if !ok {
			return ginx.Result{Msg: "系统错误"}, nil
		}

		err := u.svc.Edit(ctx, int64(id.(int)), req.Name, req.Birthday, req.Resume)
		if errors.Is(err, service.ErrUserDuplicateName) {
			return ginx.Result{Msg: "用户名重复"}, err
		}

		if err != nil {
			return ginx.Result{Msg: "系统错误"}, err
		}

		return ginx.Result{Msg: "修改成功"}, nil
	}

	ginx.WrapReq[EditReq](fn)(ctx)
}

func (u *UserHandler) EditJWT(ctx *gin.Context, req EditReq, uc *UserClaims) (ginx.Result, error) {
	err := u.svc.Edit(ctx, uc.Uid, req.Name, req.Birthday, req.Resume)
	if errors.Is(err, service.ErrUserDuplicateName) {
		return ginx.Result{Msg: "用户名重复"}, err
	}

	if err != nil {
		return ginx.Result{Msg: "系统错误"}, err
	}

	return ginx.Result{Msg: "修改成功"}, nil
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	type EmptyReq struct {
	}

	fn := func(ctx *gin.Context, req EmptyReq) (ginx.Result, error) {
		s := sessions.Default(ctx)
		s.Options(sessions.Options{MaxAge: -1})

		err := s.Save()
		if err != nil {
			return ginx.Result{Msg: "登出失败"}, err
		}

		return ginx.Result{Msg: "登出成功"}, nil
	}

	ginx.WrapReq[EmptyReq](fn)(ctx)
}

func (u *UserHandler) LogoutJWT(ctx *gin.Context) {
	// todo
}
