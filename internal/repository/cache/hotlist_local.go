package cache

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/ecodeclub/ekit/mapx"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository/dao"
)

type LocalHotListCache struct {
	// key 是 biz
	// value 是 map[string]domain.HotList  key 是该 biz 中的 id
	cache *ristretto.Cache

	keyOperating cmap.ConcurrentMap[string, struct{}] // 正在操作的key
}

type HotListCacheLocal HotListCache // 使用别名，wire可以处理多个相同类型参数

func NewLocalHotListCache(cache *ristretto.Cache) HotListCacheLocal {
	return &LocalHotListCache{
		cache:        cache,
		keyOperating: cmap.New[struct{}](),
	}
}

func (l *LocalHotListCache) getLockByKey(key string) error {
	// 不存在返回的是true，存在返回的是false
	notExist := l.keyOperating.SetIfAbsent(key, struct{}{})
	if !notExist {
		return ErrCodeOperationTooMany
	}

	return nil
}

func (l *LocalHotListCache) releaseLockByKey(key string) {
	l.keyOperating.Remove(key)
}

func (l *LocalHotListCache) GetLikeTopN(bizs []string) (map[string][]domain.HotList, error) {
	res := make(map[string][]domain.HotList)
	for i := range bizs { // 此处可修改为并发操作
		biz := bizs[i]

		if err := l.getLockByKey(biz); err != nil {
			continue
		}

		value, ok := l.cache.Get(biz)
		if !ok { // 没找到
			//ErrUnknownForCode
			l.releaseLockByKey(biz)
			continue
		}

		hotList, ok := value.(map[string]domain.HotList)
		if !ok {
			//ErrUnknownForCode
			l.releaseLockByKey(biz)
			continue
		}

		values := mapx.Values[string, domain.HotList](hotList)

		sort.Sort(hotListSlice(values))

		if len(values) >= 100 {
			res[biz] = values[0:100]
		} else {
			res[biz] = values
		}

		l.releaseLockByKey(biz)
	}

	return res, nil
}

type hotListSlice []domain.HotList

func (h hotListSlice) Len() int {
	return len(h)
}

func (h hotListSlice) Less(i, j int) bool {
	return h[i].Cnt > h[j].Cnt
}

func (h hotListSlice) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (l *LocalHotListCache) addValue(biz string, bizId int64, likeCnt int64) {
	value, ok := l.cache.Get(biz)
	if !ok { // 没找到
		fmt.Println("没找到")
		l.cache.Set(biz, map[string]domain.HotList{fmt.Sprint(bizId): {Biz: biz, Id: fmt.Sprint(bizId), Cnt: fmt.Sprint(likeCnt)}}, 1)
		time.Sleep(10 * time.Second) //  要等待一下！！
		return
	}

	hotList, ok := value.(map[string]domain.HotList)
	if !ok {
		fmt.Println("!OK")
		// log
		return
	}

	hotList[fmt.Sprint(bizId)] = domain.HotList{Biz: biz, Id: fmt.Sprint(bizId), Cnt: fmt.Sprint(likeCnt)}

	l.cache.Set(biz, hotList, 1)

	return
}

func (l *LocalHotListCache) SaveHotListToCache(biz string, hotList []dao.Interactive) error {
	if err := l.getLockByKey(biz); err != nil {
		return err
	}

	defer l.releaseLockByKey(biz)

	for i := range hotList {
		l.addValue(biz, hotList[i].BizId, hotList[i].LikeCnt)
	}

	return nil
}

func (l *LocalHotListCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	if err := l.getLockByKey(biz); err != nil {
		return err
	}

	defer l.releaseLockByKey(biz)

	value, ok := l.cache.Get(biz)
	if !ok { // 没找到
		//l.cache.Set(biz, map[string]domain.HotList{fmt.Sprint(bizId): {Biz: biz, Id: fmt.Sprint(bizId), Cnt: "1"}}, 1)
		return nil
	}

	hotList, ok := value.(map[string]domain.HotList)
	if !ok {
		// log
		return nil
	}

	if _, hasKey := hotList[fmt.Sprint(bizId)]; hasKey {
		cnt, _ := strconv.Atoi(hotList[fmt.Sprint(bizId)].Cnt)
		hotList[fmt.Sprint(bizId)] = domain.HotList{Biz: biz, Id: fmt.Sprint(bizId), Cnt: fmt.Sprint(cnt + 1)}
	} else {
		//hotList[fmt.Sprint(bizId)] = domain.HotList{Biz: biz, Id: fmt.Sprint(bizId), Cnt: "1"}
	}

	l.cache.Set(biz, hotList, 1)

	return nil
}

func (l *LocalHotListCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	if err := l.getLockByKey(biz); err != nil {
		return err
	}

	defer l.releaseLockByKey(biz)

	value, ok := l.cache.Get(biz)
	if !ok { // 没找到
		return nil
	}

	hotList, ok := value.(map[string]domain.HotList)
	if !ok {
		// log
		return nil
	}

	if _, hasKey := hotList[fmt.Sprint(bizId)]; hasKey {
		cnt, _ := strconv.Atoi(hotList[fmt.Sprint(bizId)].Cnt)
		hotList[fmt.Sprint(bizId)] = domain.HotList{Biz: biz, Id: fmt.Sprint(bizId), Cnt: fmt.Sprint(cnt - 1)}
		l.cache.Set(biz, hotList, 1)
	}

	return nil
}
