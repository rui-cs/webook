package ratelimit

import (
	"context"
	"fmt"

	"github.com/rui-cs/webook/internal/service/sms"
	"github.com/rui-cs/webook/pkg/ratelimit"
)

var errLimited = fmt.Errorf("触发了限流")

type RateLimitSMSService struct {
	svc     sms.Service       // 这个是接口 // svc是被装饰者，也是最终业务逻辑的执行者
	limiter ratelimit.Limiter // 这个还是接口
	// 装饰器模式
	// 这样可以保持面向接口和依赖注入
}

func (r *RateLimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := r.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，你的下游很坑的时候，
		// 可以不限：你的下游很强，业务可用性要求很高，尽量容错策略
		// 包一下这个错误
		return fmt.Errorf("短信服务判断限流出现问题：%w", err)
	}

	if limited {
		return errLimited
	}

	err = r.svc.Send(ctx, tpl, args, numbers...)
	return err
}

func NewRateLimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}
