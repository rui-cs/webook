package cache

import (
	"context"
	"fmt"

	"github.com/rui-cs/webook/internal/domain"
)

type InteractiveCacheHotList struct {
	// 装饰器模式
	cache InteractiveCacheRedis

	localHotList HotListCacheLocal
	redisHotList HotListCache
}

func NewInteractiveCacheHotList(cache InteractiveCacheRedis, localHotList HotListCacheLocal, redisHotList HotListCache) InteractiveCache {
	return &InteractiveCacheHotList{cache: cache, localHotList: localHotList, redisHotList: redisHotList}
}

func (i *InteractiveCacheHotList) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return i.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (i *InteractiveCacheHotList) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	if err := i.localHotList.IncrLikeCntIfPresent(ctx, biz, bizId); err != nil {
		fmt.Println("i.localHotList.IncrLikeCntIfPresent error : ", err)
	}
	// 这部分先不拆出来了，没有想好，先在interactive cache的lua脚本中实现功能
	//if err := i.redisHotList.IncrLikeCntIfPresent(ctx, biz, bizId); err != nil {
	//	fmt.Println("i.redisHotList.IncrLikeCntIfPresent error : ", err)
	//}

	return i.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (i *InteractiveCacheHotList) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	if err := i.localHotList.DecrLikeCntIfPresent(ctx, biz, bizId); err != nil {
		fmt.Println("i.localHotList.DecrLikeCntIfPresent error : ", err)
	}

	// 这部分先不拆出来了，没有想好，先在interactive cache的lua脚本中实现功能
	//if err := i.redisHotList.DecrLikeCntIfPresent(ctx, biz, bizId); err != nil {
	//	fmt.Println("i.redisHotList.DecrLikeCntIfPresent error : ", err)
	//}

	return i.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (i *InteractiveCacheHotList) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return i.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (i *InteractiveCacheHotList) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	return i.cache.Get(ctx, biz, bizId)
}

func (i *InteractiveCacheHotList) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	return i.cache.Set(ctx, biz, bizId, intr)
}
