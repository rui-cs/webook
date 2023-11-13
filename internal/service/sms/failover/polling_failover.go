package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"

	"github.com/rui-cs/webook/internal/service/sms"
)

type FailoverSMSService struct {
	// 装饰器模式
	svcs []sms.Service
	idx  uint64
}

func NewFailoverSMSService(svcs []sms.Service) sms.Service {
	return &FailoverSMSService{svcs: svcs}
}

// 第一种实现：直接轮询
func (f *FailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tpl, args, numbers...)
		if err == nil { // 发送成功
			return nil
		}

		// 输出日志，做监控
		log.Println(err)
	}

	return errors.New("全部服务商都失败了")
}

// 第二种实现：起始 svc 是轮询的
func (f *FailoverSMSService) SendV1(ctx context.Context, tpl string, args []string, numbers ...string) error {
	// 取下一个节点做起始节点
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))

	for i := idx; i < idx+length; i++ {
		svc := f.svcs[int(i%length)]
		err := svc.Send(ctx, tpl, args, numbers...)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return err
		default:
			// 输出日志
		}
	}

	return errors.New("全部服务商都失败了")
}
