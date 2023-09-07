package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	cmap "github.com/orcaman/concurrent-map/v2"
)

var ErrCodeOperationTooMany = errors.New("操作频繁，请稍后再试")

type MemoryCodeCache struct {
	cache        *ristretto.Cache
	keyOperating cmap.ConcurrentMap[string, struct{}] // 正在操作的key
	lock         sync.Mutex
}

func NewMemoryCodeCache(cache *ristretto.Cache) CodeCache {
	return &MemoryCodeCache{
		cache:        cache,
		keyOperating: cmap.New[struct{}](),
		lock:         sync.Mutex{},
	}
}

type MemCode struct {
	code string
	cnt  int
}

func (c *MemoryCodeCache) getLock(biz, phone string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.keyOperating.Has(c.key(biz, phone)) {
		return ErrCodeOperationTooMany
	}

	c.keyOperating.Set(c.key(biz, phone), struct{}{})

	return nil
}

func (c *MemoryCodeCache) releaseLock(biz, phone string) {
	c.keyOperating.Remove(c.key(biz, phone))
}

func (c *MemoryCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	if err := c.getLock(biz, phone); err != nil {
		return err
	}

	defer c.releaseLock(biz, phone)

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
	// 离上一次发送还没多长时间
	if ttl > 9*60*time.Second { // 离过期时间超过9分钟
		return ErrCodeSendTooMany
	}

	// 离上次发送过去了一段时间。离过期时间不到9分钟，再设置一遍，相当于重新发送
	if ok0 := c.cache.SetWithTTL(c.key(biz, phone), MemCode{
		code: code,
		cnt:  3,
	}, 1, 600*time.Second); !ok0 {
		return ErrUnknownForCode
	}

	return nil
}

func (c *MemoryCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	if err := c.getLock(biz, phone); err != nil {
		return false, err
	}

	defer c.releaseLock(biz, phone)

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
