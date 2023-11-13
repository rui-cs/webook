package startup

import (
	"github.com/rui-cs/webook/internal/service/oauth2/wechat"
	"github.com/rui-cs/webook/pkg/logger"
)

func InitPhantomWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("", "", l)
}
