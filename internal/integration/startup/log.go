package startup

import "github.com/rui-cs/webook/pkg/logger"

func InitLog() logger.LoggerV1 {
	return logger.NewNoOpLogger()
}
