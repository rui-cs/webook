package ioc

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rui-cs/webook/config"
)

func InitRedis() redis.Cmdable {
	client := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%s", config.Config.RCg.Addr, config.Config.RCg.Port)})
	return client
}
