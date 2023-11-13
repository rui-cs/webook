package ioc

import (
	"os"

	"github.com/rui-cs/webook/internal/service/oauth2/wechat"
	//"github.com/rui-cs/webook/internal/web"
	logger2 "github.com/rui-cs/webook/pkg/logger"
)

func InitWechatService(l logger2.LoggerV1) wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_ID ")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_SECRET")
	}
	// 692jdHsogrsYqxaUK9fgxw
	return wechat.NewService(appId, appKey, l)
}

//func NewWechatHandlerConfig() web.WechatHandlerConfig {
//	return web.WechatHandlerConfig{
//		Secure: false,
//	}
//}
