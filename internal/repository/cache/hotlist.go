package cache

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/rui-cs/webook/internal/domain"
)

type HotListCache interface {
	GetLikeTopN(bizs []string) (map[string][]domain.HotList, error)
}

type RedisHotListCache struct {
	client redis.Cmdable
}

func (r *RedisHotListCache) GetLikeTopN(bizs []string) (map[string][]domain.HotList, error) {
	return ParseZsetToHotList(r.client, bizs)
}

func NewHotListCache(client redis.Cmdable) HotListCache {
	// 启动服务时初始化热榜数据

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
