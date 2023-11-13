package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

//func InitRedis() redis.Cmdable {
//	client := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%s", config.Config.RCg.Addr, config.Config.RCg.Port)})
//	return client
//}

func InitRedis() redis.Cmdable {
	addr := viper.GetString("redis.addr")
	redisClient := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return redisClient
}
