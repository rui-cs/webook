package repository

import (
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository/cache"
)

type HotListRepo interface {
	GetLikeTopN(bizs []string) (map[string][]domain.HotList, error)
}

type CachedHotListRepo struct {
	cache cache.HotListCache
}

func NewHotListRepo(cache cache.HotListCache) HotListRepo {
	return &CachedHotListRepo{cache: cache}
}

func (c *CachedHotListRepo) GetLikeTopN(bizs []string) (map[string][]domain.HotList, error) {
	return c.cache.GetLikeTopN(bizs)
}
