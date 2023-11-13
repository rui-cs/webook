package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

//go:generate mockgen -source=./interactive.go -package=daomocks -destination=mocks/interactive.mock.go InteractiveDAO
type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error)
	DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

// biz = a, bizid = 1
// biz = a, bizid = 1
// biz = a, bizid = 1
// biz = a bizId = 2
// biz = a bizId = 2
// biz = a bizId = 2
// biz = a bizId = 2
func (dao *GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	// 可以用 map 合并吗？
	// 看情况。如果一批次里面，biz 和 bizid 都相等的占很多，那么就map 合并，性能会更好
	// 不然你合并了没有效果

	// 为什么快？
	// A：十条消息调用十次 IncrReadCnt，
	// B 就是批量
	// 事务本身的开销，A 是 B 的十倍
	// 刷新 redolog, undolog, binlog 到磁盘，A 是十次，B 是一次
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDAO := NewGORMInteractiveDAO(tx)
		for i := range bizs {
			err := txDAO.IncrReadCnt(ctx, bizs[i], ids[i])
			if err != nil {
				// 记个日志就拉到
				// 也可以 return err
				return err
			}
		}
		return nil
	})
}

func (dao *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).
		Where("biz=? AND biz_id = ? AND uid = ? AND status = ?",
			biz, bizId, uid, 1).First(&res).Error
	return res, err
}

func (dao *GORMInteractiveDAO) GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).
		Where("biz=? AND biz_id = ? AND uid = ?", biz, bizId, uid).First(&res).Error
	return res, err
}

// InsertCollectionBiz 和 InsertLikeInfo 能不能抽取出来，
// 适当的重复（复制-粘贴）要比强行抽象要更加好一点
//func (dao *GORMInteractiveDAO) common(ctx context.Context,
//	biz any, column string, intr Interactive) error {
//	now := time.Now().UnixMilli()
//	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
//		err := dao.db.WithContext(ctx).Create(&biz).Error
//		if err != nil {
//			return err
//		}
//		return tx.Clauses(clause.OnConflict{
//			DoUpdates: clause.Assignments(map[string]any{
//				column:  gorm.Expr(fmt.Sprintf("`%s`+1", column)),
//				"utime": now,
//			}),
//		}).Create(&intr).Error
//	})
//}
//func (dao *GORMInteractiveDAO) InsertCollectionBizV1(ctx context.Context,
//	cb UserCollectionBiz) error {
//	now := time.Now().UnixMilli()
//	return dao.common(ctx, cb, "collect_cnt", Interactive{
//		CollectCnt: 1,
//		Ctime:      now,
//		Utime:      now,
//		Biz:        cb.Biz,
//		BizId:      cb.BizId,
//	})
//}

// InsertCollectionBiz 插入收藏记录，并且更新计数
func (dao *GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context,
	cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.Utime = now
	cb.Ctime = now
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 插入收藏项目
		err := dao.db.WithContext(ctx).Create(&cb).Error
		if err != nil {
			return err
		}
		// 这边就是更新数量
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"collect_cnt": gorm.Expr("`collect_cnt`+1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
			Biz:        cb.Biz,
			BizId:      cb.BizId,
		}).Error
	})
}

func (dao *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context,
	biz string, bizId, uid int64) error {
	// 一把梭
	// 同时记录点赞，以及更新点赞计数
	// 首先你需要一张表来记录，谁点给什么资源点了赞
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先准备插入点赞记录
		// 有没有可能已经点赞过了？
		// 我要不要校验一下，这里必须是没有点赞过
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Biz:    biz,
			BizId:  bizId,
			Uid:    uid,
			Status: 1,
			Ctime:  now,
			Utime:  now,
		}).Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			// MySQL 不写
			//Columns:
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("like_cnt + 1"),
				"utime":    time.Now().UnixMilli(),
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   bizId,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
}

func (dao *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	now := time.Now().UnixMilli()
	// 控制事务超时
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 两个操作
		// 一个是软删除点赞记录
		// 一个是减点赞数量
		err := tx.Model(&UserLikeBiz{}).
			Where("biz=? AND biz_id = ? AND uid = ?", biz, bizId, uid).
			Updates(map[string]any{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).
			// 这边命中了索引，然后没找到，所以不会加锁
			Where("biz=? AND biz_id = ?", biz, bizId).
			Updates(map[string]any{
				"utime":    now,
				"like_cnt": gorm.Expr("like_cnt-1"),
			}).Error
	})
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{
		db: db,
	}
}

// IncrReadCnt 是一个插入或者更新语义
func (dao *GORMInteractiveDAO) IncrReadCnt(ctx context.Context,
	biz string, bizId int64) error {
	// DAO 要怎么实现？表结构该怎么设计？
	//var intr Interactive
	//err := dao.db.
	//	Where("biz_id =? AND biz = ?", bizId, biz).
	//	First(&intr).Error
	// 两个 goroutine 过来，你查询到 read_cnt 都是 10
	//if err != nil {
	//	return err
	//}
	// 都变成了 11
	//cnt := intr.ReadCnt + 1
	//// 最终变成 11
	//dao.db.Where("biz_id =? AND biz = ?", bizId, biz).Updates(map[string]any{
	//	"read_cnt": cnt,
	//})

	// update a = a + 1
	// 数据库帮你解决并发问题
	// 有一个没考虑到，就是，我可能根本没这一行
	// 事实上这里是一个 upsert 的语义
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		// MySQL 不写
		//Columns:
		DoUpdates: clause.Assignments(map[string]any{
			"read_cnt": gorm.Expr("read_cnt + 1"),
			"utime":    time.Now().UnixMilli(),
		}),
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

func (dao *GORMInteractiveDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).
		Where("biz = ? AND biz_id = ?", biz, bizId).
		First(&res).Error

	//fmt.Println("-----------")
	//fmt.Println(err)
	//fmt.Println("-----------")
	return res, err
}

// Interactive 正常来说，一张主表和与它有关联关系的表会共用一个DAO，
// 所以我们就用一个 DAO 来操作
// 假如说我要查找点赞数量前 100 的，
// SELECT * FROM
// (SELECT biz, biz_id, COUNT(*) as cnt FROM `interactives` GROUP BY biz, biz_id) ORDER BY cnt LIMIT 100
// 实时查找，性能贼差，上面这个语句，就是全表扫描，
// 高性能，我不要求准确性
// 面试标准答案：用 zset
// 但是，面试标准答案不够有特色，烂大街了
// 你可以考虑别的方案
// 1. 定时计算
// 1.1 定时计算 + 本地缓存
// 2. 优化版的 zset，定时筛选 + 实时 zset 计算
// 还要别的方案你们也可以考虑
type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 业务标识符
	// 同一个资源，在这里应该只有一行
	// 也就是说我要在 bizId 和 biz 上创建联合唯一索引
	// 1. bizId, biz。优先选择这个，因为 bizId 的区分度更高
	// 2. biz, bizId。如果有 WHERE biz = xx 这种查询条件（不带 bizId）的，就只能这种
	//
	// 联合索引的列的顺序：查询条件，区分度
	// 这个名字无所谓
	BizId int64 `gorm:"uniqueIndex:biz_id_type"`
	// 我这里biz 用的是 string，有些公司枚举使用的是 int 类型
	// 0-article
	// 1- xxx
	// 默认是 BLOB/TEXT 类型
	Biz string `gorm:"uniqueIndex:biz_id_type;type:varchar(128)"`
	// 这个是阅读计数
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

// InteractiveV1 对写更友好
// Interactive 对读更加友好
type InteractiveV1 struct {
	Id    int64 `gorm:"primaryKey,autoIncrement"`
	BizId int64
	Biz   string
	// 这个是阅读计数
	Cnt int64
	// 阅读数/点赞数/收藏数
	CntType string
	Ctime   int64
	Utime   int64
}

// UserLikeBiz 命名无能，用户点赞的某个东西
type UserLikeBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`

	// 我在前端展示的时候，
	// WHERE uid = ? AND biz_id = ? AND biz = ?
	// 来判定你有没有点赞
	// 这里，联合顺序应该是什么？

	// 要分场景
	// 1. 如果你的场景是，用户要看看自己点赞过那些，那么 Uid 在前
	// WHERE uid =?
	// 2. 如果你的场景是，我的点赞数量，需要通过这里来比较/纠正
	// biz_id 和 biz 在前
	// select count(*) where biz = ? and biz_id = ?
	Biz   string `gorm:"uniqueIndex:uid_biz_id_type;type:varchar(128)"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_id_type"`

	// 谁的操作
	Uid int64 `gorm:"uniqueIndex:uid_biz_id_type"`

	Ctime int64
	Utime int64
	// 如果这样设计，那么，取消点赞的时候，怎么办？
	// 我删了这个数据
	// 你就软删除
	// 这个状态是存储状态，纯纯用于软删除的，业务层面上没有感知
	// 0-代表删除，1 代表有效
	Status uint8

	// 有效/无效
	//Type string
}

// Collection 收藏夹
type Collection struct {
	Id   int64  `gorm:"primaryKey,autoIncrement"`
	Name string `gorm:"type=varchar(1024)"`
	Uid  int64  `gorm:""`

	Ctime int64
	Utime int64
}

// UserCollectionBiz 收藏的东西
type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 收藏夹 ID
	// 作为关联关系中的外键，我们这里需要索引
	Cid   int64  `gorm:"index"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	// 这算是一个冗余，因为正常来说，
	// 只需要在 Collection 中维持住 Uid 就可以
	Uid   int64 `gorm:"uniqueIndex:biz_type_id_uid"`
	Ctime int64
	Utime int64
}

// 假如说我有一个需求，需要查询到收藏夹的信息，和收藏夹里面的资源
// SELECT c.id as cid , c.name as cname, uc.biz_id as biz_id, uc.biz as biz
// FROM `collection` as c JOIN `user_collection_biz` as uc
// ON c.id = uc.cid
// WHERE c.id IN (1,2,3)

type CollectionItem struct {
	Cid   int64
	Cname string
	BizId int64
	Biz   string
}

func (dao *GORMInteractiveDAO) GetItems() ([]CollectionItem, error) {
	// 不记得构造 JOIN 查询
	var items []CollectionItem
	err := dao.db.Raw("", 1, 2, 3).Find(&items).Error
	return items, err
}
