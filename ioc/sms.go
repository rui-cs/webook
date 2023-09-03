package ioc

import (
	"github.com/rui-cs/webook/internal/service/sms"
	"github.com/rui-cs/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
