package cache

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rui-cs/webook/internal/domain"
)

var (
	//go:embed lua/interative_incr_cnt.lua
	luaIncrCnt string

	//go:embed lua/interative_incr_cnt_like.lua
	luaIncrLikeCnt string
)

const (
	fieldReadCnt    = "read_cnt"
	fieldCollectCnt = "collect_cnt"
	fieldLikeCnt    = "like_cnt"
)

//go:generate mockgen -source=./interactive.go -package=cachemocks -destination=mocks/interactive.mock.go InteractiveCache
type InteractiveCache interface {

	// IncrReadCntIfPresent 如果在缓存中有对应的数据，就 +1
	IncrReadCntIfPresent(ctx context.Context,
		biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context,
		biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context,
		biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	// Get 查询缓存中数据
	// 事实上，这里 liked 和 collected 是不需要缓存的
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error
}

// 方案1
// key1 => map[string]int

// 方案2
// key1_read_cnt => 10
// key1_collect_cnt => 11
// key1_like_cnt => 13

type RedisInteractiveCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

type InteractiveCacheRedis InteractiveCache

func (r *RedisInteractiveCache) IncrCollectCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId),
		fieldCollectCnt}, 1).Err()
}

func (r *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	// 拿到的结果，可能自增成功了，可能不需要自增（key不存在）
	// 你要不要返回一个 error 表达 key 不存在？
	//res, err := r.client.Eval(ctx, luaIncrCnt,
	//	[]string{r.key(biz, bizId)},
	//	// read_cnt +1
	//	"read_cnt", 1).Int()
	//if err != nil {
	//	return err
	//}
	//if res == 0 {
	// 这边一般是缓存过期了
	//	return errors.New("缓存中 key 不存在")
	//}
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		// read_cnt +1
		fieldReadCnt, 1).Err()
}

const threshold = 400000

func (r *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrLikeCnt,
		[]string{r.key(biz, bizId)},
		fieldLikeCnt, 1, fieldLikeCnt, biz, bizId, threshold).Err()
}

//func (r *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context,
//	biz string, bizId int64) error {
//	return r.client.Eval(ctx, luaIncrCnt,
//		[]string{r.key(biz, bizId)},
//		fieldLikeCnt, 1).Err()
//}

func (r *RedisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrLikeCnt,
		[]string{r.key(biz, bizId)},
		fieldLikeCnt, -1, fieldLikeCnt, biz, bizId, threshold).Err()
}

//func (r *RedisInteractiveCache) GetV1(ctx context.Context,
//	biz string, bizId int64) (map[string]string, error) {
//	//	你知道我会返回哪些 key 吗？
//	data, err := r.client.HGetAll(ctx, r.key(biz, bizId)).Result()
//	if err != nil {
//		return nil, err
//	}
//	// 你同样看不出来我会返回哪些 key
//	// 你要看完全部代码你才知道
//	return data, nil
//}

func (r *RedisInteractiveCache) Get(ctx context.Context,
	biz string, bizId int64) (domain.Interactive, error) {
	// 直接使用 HMGet，即便缓存中没有对应的 key，也不会返回 error
	//r.client.HMGet(ctx, r.key(biz, bizId),
	//	fieldCollectCnt, fieldLikeCnt, fieldReadCnt)
	// 所以你没有办法判定，缓存里面是有这个key，但是对应 cnt 都是0，还是说没有这个 key

	// 拿到 key 对应的值里面的所有的 field
	data, err := r.client.HGetAll(ctx, r.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	if len(data) == 0 {
		// 缓存不存在，系统错误，比如说你的同事，手贱设置了缓存，但是忘记任何 fields
		return domain.Interactive{}, ErrKeyNotExist
	}

	// 理论上来说，这里不可能有 error
	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)

	return domain.Interactive{
		// 懒惰的写法
		CollectCnt: collectCnt,
		LikeCnt:    likeCnt,
		ReadCnt:    readCnt,
	}, err
}

func (r *RedisInteractiveCache) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	key := r.key(biz, bizId)
	err := r.client.HMSet(ctx, key,
		fieldLikeCnt, intr.LikeCnt,
		fieldCollectCnt, intr.CollectCnt,
		fieldReadCnt, intr.ReadCnt).Err()
	if err != nil {
		return err
	}
	return r.client.Expire(ctx, key, time.Minute*15).Err()
}

func (r *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}

//func (r *RedisInteractiveCache) keyPersonal(biz string, bizId int64) string {
//	return fmt.Sprintf("interactive:personal:%s:%d:%d", biz, bizId, uid)
//}

func NewRedisInteractiveCache(client redis.Cmdable) InteractiveCacheRedis {
	return &RedisInteractiveCache{
		client: client,
	}
}
