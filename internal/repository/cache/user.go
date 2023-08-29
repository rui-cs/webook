package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rui-cs/webook/internal/domain"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

/*
A 用到了 B，B 一定是接口 => 这个是保证面向接口
A 用到了 B，B 一定是 A 的字段 => 规避包变量、包方法，都非常缺乏扩展性
A 用到了 B，A 绝对不初始化 B，而是外面注入 => 保持依赖注入(DI, Dependency Injection)和依赖反转(IOC)
*/
func NewUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (c *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}

func (c *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := c.key(id)

	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}

	var u domain.User
	err = json.Unmarshal(val, &u)

	return u, err
}

func (c *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := c.key(u.Id)
	return c.client.Set(ctx, key, val, c.expiration).Err()
}
