package failover

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/rui-cs/webook/internal/service/sms"
)

type TimeoutFailoverSMSService struct {
	svcs []sms.Service //服务商
	idx  int32
	cnt  int32 // 连续超时的个数

	threshold int32 // 阈值 连续超时超过这个数字，就要切换
}

func NewTimeoutFailoverSMSService(svcs []sms.Service) sms.Service {
	return &TimeoutFailoverSMSService{svcs: svcs, threshold: 10}
}

// 只要连续超过 N 个请求超时了，就直接切换
func (t *TimeoutFailoverSMSService) Send(ctx context.Context,
	tpl string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)

	if cnt > t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))

		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			atomic.StoreInt32(&t.cnt, 0)
		}

		// else 出现并发别人切换成功了

		idx = atomic.LoadInt32(&t.idx)
	}

	svc := t.svcs[idx]
	err := svc.Send(ctx, tpl, args, numbers...)
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		atomic.AddInt32(&t.cnt, 1)
		return err
	case err == nil:
		//连接状态被打断了
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	default:
		// 未知错误
		// 可以考虑，换下一个，语义则是：
		// - 超时错误，可能是偶发的，我尽量再试试
		// - 非超时，我直接下一个
		return err
	}
}
