package cache

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository/dao"
)

type HotListCache interface {
	GetLikeTopN(bizs []string) (map[string][]domain.HotList, error)
	SaveHotListToCache(biz string, hotList []dao.Interactive) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
}

type RedisHotListCache struct {
	client redis.Cmdable
}

func (r *RedisHotListCache) SaveHotListToCache(biz string, hotList []dao.Interactive) error {
	zset := make([]redis.Z, len(hotList))
	for i := range hotList {
		zset[i] = redis.Z{
			Score:  float64(hotList[i].LikeCnt),
			Member: hotList[i].BizId,
		}
	}

	key := fmt.Sprintf("hotlist:biz:%s:like", biz)

	if err := r.client.Del(context.Background(), key).Err(); err != nil {
		fmt.Println("client.Del error : ", err)
	}

	if err := r.client.ZAdd(context.Background(), key, zset...).Err(); err != nil {
		fmt.Println("client.ZAdd error : ", err)
		return err
	}

	return nil
}

func (r *RedisHotListCache) GetLikeTopN(bizs []string) (map[string][]domain.HotList, error) {
	return ParseZsetToHotList(r.client, bizs)
}

//var (
//	//go:embed lua/hotlist_incr_cnt_like.lua
//	luaIncrHotListLikeCnt string
//)

func (r *RedisHotListCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	panic("RedisHotListCache.IncrLikeCntIfPresent")

	//return r.client.Eval(ctx, luaIncrLikeCnt,
	//	[]string{r.key(biz, bizId)},
	//	fieldLikeCnt, 1, fieldLikeCnt, biz, bizId, 100000).Err()
}

func (r *RedisHotListCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	panic("RedisHotListCache.DecrLikeCntIfPresent")
}

func NewRedisHotListCache(client redis.Cmdable) HotListCache {
	return &RedisHotListCache{client: client}
}

//go:embed lua/hotlist_like_bizs.lua
var luaEvalCode string

func ParseZsetToHotList(rdb redis.Cmdable, bizs []string) (map[string][]domain.HotList, error) {
	redisRes, err := rdb.Eval(context.Background(), luaEvalCode, bizs).Result()
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(redisRes)
	if err != nil {
		return nil, err
	}

	var hotListRes [][]string
	err = json.Unmarshal(b, &hotListRes)
	if err != nil {
		return nil, err
	}

	if len(hotListRes) != len(bizs) {
		// add log
		return nil, err
	}

	res := make(map[string][]domain.HotList)
	for i := range hotListRes {
		if len(hotListRes[i])%2 != 0 {
			// add log
			continue
		}

		res[bizs[i]] = make([]domain.HotList, len(hotListRes[i])/2)

		for j := 0; j < len(hotListRes[i])/2; j++ {
			res[bizs[i]][j] = domain.HotList{
				Biz:  bizs[i],
				Id:   hotListRes[i][j*2+0],
				Cnt:  hotListRes[i][j*2+1],
				Name: "",
			}

		}
	}

	return res, nil
}
