package ioc

import (
	"github.com/rui-cs/webook/internal/repository"
	"github.com/rui-cs/webook/internal/repository/cache"
	"github.com/rui-cs/webook/internal/repository/dao"
)

func InitHotListRepo(cache cache.HotListCache, localCache cache.HotListCacheLocal, dao dao.HotListDao) repository.HotListRepo {
	r := repository.NewCachedHotListRepo(cache, localCache, dao)
	r.Preload()
	r.AddHotListCron()

	return r
}
