package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
)

type MemoryCodeCache struct {
	cache *ristretto.Cache
}

func NewMemoryCodeCache(cache *ristretto.Cache) CodeCache {
	return &MemoryCodeCache{cache: cache}
}

type MemCode struct {
	code string
	cnt  int
}

func (c *MemoryCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	ttl, ok := c.cache.GetTTL(c.key(biz, phone))
	if !ok {
		// 未设置过该key 或者 找到了但是过期了
		// 那么可以设置该验证码
		if ok0 := c.cache.SetWithTTL(c.key(biz, phone), MemCode{
			code: code,
			cnt:  3,
		}, 1, 600*time.Second); !ok0 {
			return ErrUnknownForCode
		}

		return nil
	}

	// 判断过期情况
	if ttl > 9*60*time.Second { // 离过期时间超过9分钟
		return ErrCodeSendTooMany
	}

	// 离过期时间不到9分钟，再设置一遍
	if ok0 := c.cache.SetWithTTL(c.key(biz, phone), MemCode{
		code: code,
		cnt:  3,
	}, 1, 600*time.Second); !ok0 {
		return ErrUnknownForCode
	}

	return nil
}

func (c *MemoryCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	ttl, ok := c.cache.GetTTL(c.key(biz, phone))
	if !ok {
		// 未找到这个 key 或者 已经超时
		// 返回验证不通过，没有错误
		return false, nil
	}

	value, ok := c.cache.Get(c.key(biz, phone))
	if !ok { // 没找到
		return false, ErrUnknownForCode
	}

	memCode, ok := value.(MemCode)
	if !ok {
		return false, ErrUnknownForCode
	}

	if memCode.code == inputCode && memCode.cnt >= 1 { // 验证码对上了
		c.cache.Del(c.key(biz, phone)) // 要删除这个key

		return true, nil
	}

	// 验证次数要减一
	memCode.cnt--
	if memCode.cnt <= 0 { // 没有剩余次数了
		c.cache.Del(c.key(biz, phone)) // 要删除这个key

		return false, ErrCodeVerifyTooManyTimes
	}

	if ok0 := c.cache.SetWithTTL(c.key(biz, phone), memCode, 1, ttl); !ok0 {
		return false, ErrUnknownForCode
	}

	return false, nil
}

func (c *MemoryCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
