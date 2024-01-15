package dao

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type HotListDao interface {
	GetBizList() ([]string, error)
	GetHotListByBiz(biz string) ([]Interactive, error)
}

type GORMHotListDao struct {
	db *gorm.DB
}

func (d *GORMHotListDao) GetBizList() ([]string, error) {
	var bizs []string
	if err := d.db.Model(&Interactive{}).Distinct().Pluck("biz", &bizs).Error; err != nil { //select distinct biz from interactives
		return nil, errors.Wrap(err, "CachedHotListRepo.Preload error.")
	}

	return bizs, nil
}

const threshold = 400000

func (d *GORMHotListDao) GetHotListByBiz(biz string) ([]Interactive, error) {
	//	SELECT biz, biz_id, like_cnt FROM `interactives` where biz = 'article' and like_cnt > 0 ORDER BY like_cnt DESC LIMIT 100
	var res []Interactive
	if err := d.db.Model(&Interactive{}).Where("biz = ? and like_cnt >= ?", biz, threshold).Order("like_cnt DESC").Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func NewHotListDao(db *gorm.DB) HotListDao {
	return &GORMHotListDao{db: db}
}
