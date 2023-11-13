package ratelimit

import "context"

type Limiter interface {
	// bool : 是否触发了限流，true 代表需要限流
	// error : 限流器本身是否有错误
	Limit(ctx context.Context, key string) (bool, error)
}
