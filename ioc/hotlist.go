package ioc

import (
	"github.com/rui-cs/webook/internal/repository"
	"github.com/rui-cs/webook/internal/repository/cache"
	"github.com/rui-cs/webook/internal/repository/dao"
)

func InitHotListRepo(cache cache.HotListCache, dao dao.HotListDao) repository.HotListRepo {
	r := repository.NewCachedHotListRepo(cache, dao)
	r.Preload()
	r.AddHotListCron()

	return r
}
