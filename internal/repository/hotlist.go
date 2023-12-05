package repository

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository/cache"
	"github.com/rui-cs/webook/internal/repository/dao"
)

type HotListRepo interface {
	GetLikeTopN(bizs []string) (map[string][]domain.HotList, error)
}

type CachedHotListRepo struct {
	cache cache.HotListCache
	dao   dao.HotListDao
}

func NewHotListRepo(cache cache.HotListCache, dao dao.HotListDao) HotListRepo {
	return &CachedHotListRepo{cache: cache, dao: dao}
}

func NewCachedHotListRepo(cache cache.HotListCache, dao dao.HotListDao) *CachedHotListRepo {
	return &CachedHotListRepo{cache: cache, dao: dao}
}

func (r *CachedHotListRepo) GetLikeTopN(bizs []string) (map[string][]domain.HotList, error) {
	return r.cache.GetLikeTopN(bizs)
}

func (r *CachedHotListRepo) Preload() {
	bizs, err := r.dao.GetBizList()
	if err != nil {
		return
	}
	//fmt.Println("bizs : ", bizs)

	for i := range bizs {
		hotList, err := r.dao.GetHotListByBiz(bizs[i])
		if err != nil {
			fmt.Println("r.dao.GetHotListByBiz error : ", err)
		}
		//fmt.Println(hotList)
		if err := r.cache.SaveHotListToCache(bizs[i], hotList); err != nil {
			fmt.Println("r.cache.SaveHotListToCache error : ", err)
		}
	}
}

func (r *CachedHotListRepo) AddHotListCron() {
	c := cron.New()
	c.AddFunc("0 0 2 * * *", func() {
		fmt.Println("hotlist cron. time : ", time.Now())
		r.Preload()
	})
	c.Start()
}
