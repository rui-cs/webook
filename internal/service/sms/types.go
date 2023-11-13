package sms

import "context"

type Service interface {
	// Send biz 很含糊的业务
	Send(ctx context.Context, biz string, args []string, numbers ...string) error
	//SendV1(ctx context.Context, tpl string, args []NamedArg, numbers ...string) error
	// 调用者需要知道实现者需要什么类型的参数，是 []string，还是 map[string]string
	//SendV2(ctx context.Context, tpl string, args any, numbers ...string) error
}
