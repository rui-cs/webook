package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/service"
	svcmocks "github.com/rui-cs/webook/internal/service/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestEmailRegex(t *testing.T) {
	emailExp := regexp.MustCompile(emailRegex, regexp.None)
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "right email",
			email: "12223456432@qq.com",
			want:  true,
		},
		{
			name:  "wrong email",
			email: "1234567@qq",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := emailExp.MatchString(tt.email)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			if ok != tt.want {
				t.Errorf("ok != tt.want. ok : %v, tt.want : %v", ok, tt.want)
			}
		})
	}
}

func TestPassRegex(t *testing.T) {
	passExp := regexp.MustCompile(passRegex, regexp.None)
	tests := []struct {
		name string
		pass string
		want bool
	}{
		{
			name: "right password",
			//pass: "1E#3fg4et",
			pass: "Ew333W#23fget",
			want: true,
		},
		{
			name: "wrong password",
			pass: "1234",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := passExp.MatchString(tt.pass)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			if ok != tt.want {
				t.Errorf("ok != tt.want. ok : %v, tt.want : %v", ok, tt.want)
			}
		})
	}
}

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSVC := svcmocks.NewMockUserService(ctrl)
				userSVC.EXPECT().SignUp(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello#world123"}).Return(nil)
				return userSVC
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "hello#world123",
					"confirmedPassword": "hello#world123"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "参数错误，bind失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSVC := svcmocks.NewMockUserService(ctrl)
				return userSVC
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "hello#world123",
				}
			`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return svcmocks.NewMockUserService(ctrl)
			},
			reqBody: `
				{
					"email": "123qq.com",
					"password": "hello#world123",
					"confirmedPassword": "hello#world123"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "邮箱格式错误",
		},
		{
			name: "两次输入密码不一致",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return svcmocks.NewMockUserService(ctrl)
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "hello#world123",
					"confirmedPassword": "helloworld123"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "两次输入密码不一致",
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return svcmocks.NewMockUserService(ctrl)
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "hello",
					"confirmedPassword": "hello"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "密码格式错误，密码必须大于8位，包含数字、特殊字符",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSVC := svcmocks.NewMockUserService(ctrl)
				userSVC.EXPECT().SignUp(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello#world123"}).Return(service.ErrUserDuplicateEmail)
				return userSVC
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "hello#world123",
					"confirmedPassword": "hello#world123"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSVC := svcmocks.NewMockUserService(ctrl)
				userSVC.EXPECT().SignUp(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello#world123"}).Return(errors.New("随便一个error"))
				return userSVC
			},
			reqBody: `
				{
					"email": "123@qq.com",
					"password": "hello#world123",
					"confirmedPassword": "hello#world123"
				}
			`,
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			h := NewUserHandler(tc.mock(ctrl), nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name        string
		mockUserSVC func(ctrl *gomock.Controller) service.UserService
		mockCodeSVC func(ctrl *gomock.Controller) service.CodeService
		reqBody     string
		wantCode    int
		isJson      bool
		wantString  string
		wantJson    Result
	}{
		{
			name: "登录成功",
			mockUserSVC: func(ctrl *gomock.Controller) service.UserService {
				userSVC := svcmocks.NewMockUserService(ctrl)
				userSVC.EXPECT().FindOrCreate(gomock.Any(), "15612345678").Return(domain.User{}, nil)
				return userSVC
			},
			mockCodeSVC: func(ctrl *gomock.Controller) service.CodeService {
				codeSVC := svcmocks.NewMockCodeService(ctrl)
				codeSVC.EXPECT().Verify(gomock.Any(), biz, "15612345678", "234567").Return(true, nil)
				return codeSVC
			},
			reqBody: `
				{    
					"phone": "15612345678",
					"code": "234567"
				}`,
			isJson:     false,
			wantCode:   http.StatusOK,
			wantString: "登录成功",
		},
		{
			name: "参数错误，bind失败",
			mockUserSVC: func(ctrl *gomock.Controller) service.UserService {
				userSVC := svcmocks.NewMockUserService(ctrl)
				return userSVC
			},
			mockCodeSVC: func(ctrl *gomock.Controller) service.CodeService {
				codeSVC := svcmocks.NewMockCodeService(ctrl)
				return codeSVC
			},
			reqBody: `
				{    
					"phone": "15612345678",
					"code": "234567",
				}`,

			wantCode: http.StatusBadRequest,
		},
		{
			name: "Verify-系统错误",
			mockUserSVC: func(ctrl *gomock.Controller) service.UserService {
				userSVC := svcmocks.NewMockUserService(ctrl)
				return userSVC
			},
			mockCodeSVC: func(ctrl *gomock.Controller) service.CodeService {
				codeSVC := svcmocks.NewMockCodeService(ctrl)
				codeSVC.EXPECT().Verify(gomock.Any(), biz, "15612345678", "234567").Return(false, errors.New("随便一个错误"))
				return codeSVC
			},
			reqBody: `
				{    
					"phone": "15612345678",
					"code": "234567"
				}`,
			isJson:   true,
			wantCode: http.StatusOK,
			wantJson: Result{Code: 5, Msg: "系统错误"},
		},
		{
			name: "验证码有误",
			mockUserSVC: func(ctrl *gomock.Controller) service.UserService {
				userSVC := svcmocks.NewMockUserService(ctrl)
				return userSVC
			},
			mockCodeSVC: func(ctrl *gomock.Controller) service.CodeService {
				codeSVC := svcmocks.NewMockCodeService(ctrl)
				codeSVC.EXPECT().Verify(gomock.Any(), biz, "15612345678", "234567").Return(false, nil)
				return codeSVC
			},
			reqBody: `
				{    
					"phone": "15612345678",
					"code": "234567"
				}`,
			isJson:   true,
			wantJson: Result{Code: 4, Msg: "验证码有误"},
			wantCode: http.StatusOK,
		},
		{
			name: "FindOrCreate-系统错误",
			mockUserSVC: func(ctrl *gomock.Controller) service.UserService {
				userSVC := svcmocks.NewMockUserService(ctrl)
				userSVC.EXPECT().FindOrCreate(gomock.Any(), "15612345678").Return(domain.User{}, errors.New("随便一个错误"))
				return userSVC
			},
			mockCodeSVC: func(ctrl *gomock.Controller) service.CodeService {
				codeSVC := svcmocks.NewMockCodeService(ctrl)
				codeSVC.EXPECT().Verify(gomock.Any(), biz, "15612345678", "234567").Return(true, nil)
				return codeSVC
			},
			reqBody: `
				{    
					"phone": "15612345678",
					"code": "234567"
				}`,
			isJson:   true,
			wantJson: Result{Code: 5, Msg: "系统错误"},
			wantCode: http.StatusOK,
		},
		//{
		//	name:     "setJWTToken-系统错误",
		//	wantCode: http.StatusOK,
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			h := NewUserHandler(tc.mockUserSVC(ctrl), tc.mockCodeSVC(ctrl))
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if !tc.isJson {
				assert.Equal(t, tc.wantString, resp.Body.String())
			} else {
				var r Result
				err0 := json.Unmarshal(resp.Body.Bytes(), &r)
				require.NoError(t, err0)
				assert.Equal(t, tc.wantJson, r)
			}
		})
	}
}
