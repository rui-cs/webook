package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type RedisEncryptCodeCache struct {
	client redis.Cmdable
}

func NewRedisEncryptCodeCache(client redis.Cmdable) CodeCache {
	return &RedisEncryptCodeCache{client: client}
}

func (c *RedisEncryptCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	// 加密
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	code = string(hash)

	// 存储
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}

	switch res {
	case 0: // no problem
		return nil
	case -1:
		return ErrCodeSendTooMany
	default:
		return errors.New("系统错误")
	}
}

func (c *RedisEncryptCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	cnt, err := c.client.Get(ctx, c.key(biz, phone)+":cnt").Int()
	if err != nil {
		return false, err
	}

	if cnt <= 0 {
		return false, ErrCodeVerifyTooManyTimes
	}

	code, err := c.client.Get(ctx, c.key(biz, phone)).Result()
	if err != nil {
		return false, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(code), []byte(inputCode)); err != nil {
		if err0 := c.client.Decr(ctx, c.key(biz, phone)+":cnt").Err(); err0 != nil {
			fmt.Println(err0)
		}

		return false, nil
	}

	c.client.Del(ctx, c.key(biz, phone))
	c.client.Del(ctx, c.key(biz, phone)+":cnt")

	return true, nil
}

func (c *RedisEncryptCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
