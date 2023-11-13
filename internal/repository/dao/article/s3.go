package article

import (
	"bytes"
	"context"
	"strconv"
	"time"

	_ "github.com/aws/aws-sdk-go"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"github.com/rui-cs/webook/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var statusPrivate = domain.ArticleStatusPrivate.ToUint8()

type S3DAO struct {
	oss *s3.S3
	// 通过组合 GORMArticleDAO 来简化操作
	// 当然在实践中，你是不太会有组合的机会
	// 你操作制作库总是一样的
	// 你就是操作线上库的时候不一样
	GORMArticleDAO
	bucket *string
}

// NewOssDAO 因为组合 GORMArticleDAO 是一个内部实现细节
// 所以这里要直接传入 DB
func NewOssDAO(oss *s3.S3, db *gorm.DB) ArticleDAO {
	return &S3DAO{
		oss: oss,
		// 你也可以考虑利用依赖注入来传入。
		// 但是事实上这个很少变，所以你可以延迟到必要的时候再注入
		bucket: ekit.ToPtr[string]("webook-1314583317"),
		GORMArticleDAO: GORMArticleDAO{
			db: db,
		},
	}
}

func (o *S3DAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 保存制作库
	// 保存线上库，并且把 content 上传到 OSS
	//
	var (
		id = art.Id
	)
	// 制作库流量不大，并发不高，你就保存到数据库就可以
	// 当然，有钱或者体量大，就还是考虑 OSS
	err := o.db.Transaction(func(tx *gorm.DB) error {
		var err error
		now := time.Now().UnixMilli()
		// 制作库
		txDAO := NewGORMArticleDAO(tx)
		if id == 0 {
			id, err = txDAO.Insert(ctx, art)
		} else {
			err = txDAO.UpdateById(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		publishArt := PublishedArticleV1{
			Id:       art.Id,
			Title:    art.Title,
			AuthorId: art.AuthorId,
			Status:   art.Status,
			Ctime:    now,
			Utime:    now,
		}
		// 线上库不保存 Content,要准备上传到 OSS 里面
		return tx.Clauses(clause.OnConflict{
			// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":  art.Title,
				"utime":  now,
				"status": art.Status,
				// 要参与 SQL 运算的
			}),
		}).Create(&publishArt).Error
	})
	// 说明保存到数据库的时候失败了
	if err != nil {
		return 0, err
	}
	// 接下来就是保存到 OSS 里面
	// 你要有监控，你要有重试，你要有补偿机制
	_, err = o.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      o.bucket,
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	return id, err
}

func (o *S3DAO) SyncStatus(ctx context.Context, author, id int64, status uint8) error {
	panic("implement me")
}
