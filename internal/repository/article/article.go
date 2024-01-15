package article

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/repository"
	"github.com/rui-cs/webook/internal/repository/cache"
	dao "github.com/rui-cs/webook/internal/repository/dao/article"
	"github.com/rui-cs/webook/pkg/logger"
	"gorm.io/gorm"
)

// repository 还是要用来操作缓存和DAO
// 事务概念应该在 DAO 这一层

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// Sync 存储并同步数据
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error
	//FindById(ctx context.Context, id int64) domain.Article
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

type CachedArticleRepository struct {
	dao      dao.ArticleDAO
	userRepo repository.UserRepository

	// v1 操作两个 DAO
	readerDAO dao.ReaderDAO
	authorDAO dao.AuthorDAO

	// 耦合了 DAO 操作的东西
	// 正常情况下，如果你要在 repository 层面上操作事务
	// 那么就只能利用 db 开始事务之后，创建基于事务的 DAO
	// 或者，直接去掉 DAO 这一层，在 repository 的实现中，直接操作 db
	db *gorm.DB

	cache cache.ArticleCache
	l     logger.LoggerV1
}

func NewArticleRepository(dao dao.ArticleDAO, l logger.LoggerV1, cache cache.ArticleCache, userRepo repository.UserRepository) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		l:        l,
		cache:    cache,
		userRepo: userRepo,
	}
}

func (c *CachedArticleRepository) GetPublishedById(
	ctx context.Context, id int64) (domain.Article, error) {
	// 读取线上库数据，如果你的 Content 被你放过去了 OSS 上，你就要让前端去读 Content 字段
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 你在这边要组装 user 了，适合单体应用
	usr, err := c.userRepo.FindById(ctx, art.AuthorId)
	res := domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id:   usr.Id,
			Name: usr.Name,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
	return res, nil
}

func (c *CachedArticleRepository) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	data, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return c.toDomain(data), nil
}

func (c *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 你在这个地方，集成你的复杂的缓存方案
	// 你只缓存这一页
	if offset == 0 && limit <= 100 {
		data, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				c.preCache(ctx, data)
			}()
			//return data[:limit], err
			return data, err
		}
	}
	res, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})
	// 回写缓存的时候，可以同步，也可以异步
	go func() {
		err := c.cache.SetFirstPage(ctx, uid, data)
		c.l.Error("回写缓存失败", logger.Error(err))
		//fmt.Println(err)
		c.preCache(ctx, data)
	}()
	return data, nil
}

func (c *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, id, author, uint8(status))
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		c.cache.DelFirstPage(ctx, art.Author.Id)
		er := c.cache.SetPub(ctx, art)
		if er != nil {
			// 不需要特别关心
			// 比如说输出 WARN 日志
		}
	}
	return id, err
}

//func (c *CachedArticleRepository) SyncV2_1(ctx context.Context, art domain.Article) (int64, error) {
//	// 谁在控制事务，是 repository，还是DAO在控制事务？
//	c.dao.Transaction(ctx, func(txDAO dao.ArticleDAO) error {
//
//	})
//}

// SyncV2 尝试在 repository 层面上解决事务问题
// 确保保存到制作库和线上库同时成功，或者同时失败
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	// 开启了一个事务
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()
	// 利用 tx 来构建 DAO
	author := dao.NewAuthorDAO(tx)
	reader := dao.NewReaderDAO(tx)

	var (
		id  = art.Id
		err error
	)
	artn := c.toEntity(art)
	// 应该先保存到制作库，再保存到线上库
	if id > 0 {
		err = author.UpdateById(ctx, artn)
	} else {
		id, err = author.Insert(ctx, artn)
	}
	if err != nil {
		// 执行有问题，要回滚
		//tx.Rollback()
		return id, err
	}
	// 操作线上库了，保存数据，同步过来
	// 考虑到，此时线上库可能有，可能没有，你要有一个 UPSERT 的写法
	// INSERT or UPDATE
	// 如果数据库有，那么就更新，不然就插入
	err = reader.UpsertV2(ctx, dao.PublishedArticle(artn))
	// 执行成功，直接提交
	tx.Commit()
	return id, err

}

func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	artn := c.toEntity(art)
	// 应该先保存到制作库，再保存到线上库
	if id > 0 {
		err = c.authorDAO.UpdateById(ctx, artn)
	} else {
		id, err = c.authorDAO.Insert(ctx, artn)
	}
	if err != nil {
		return id, err
	}
	// 操作线上库了，保存数据，同步过来
	// 考虑到，此时线上库可能有，可能没有，你要有一个 UPSERT 的写法
	// INSERT or UPDATE
	// 如果数据库有，那么就更新，不然就插入
	err = c.readerDAO.Upsert(ctx, artn)
	return id, err
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 清空缓存
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()

	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   uint8(art.Status),
	})
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		// 清空缓存
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()

	return c.dao.UpdateById(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   uint8(art.Status),
	})
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   uint8(art.Status),
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, data []domain.Article) {
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := c.cache.Set(ctx, data[0])
		if err != nil {
			c.l.Error("提前预加载缓存失败", logger.Error(err))
		}
	}
}
