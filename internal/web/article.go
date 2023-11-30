package web

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"github.com/rui-cs/webook/internal/domain"
	"github.com/rui-cs/webook/internal/service"
	ijwt "github.com/rui-cs/webook/internal/web/jwt"
	"github.com/rui-cs/webook/pkg/ginx"
	"github.com/rui-cs/webook/pkg/logger"
	"golang.org/x/sync/errgroup"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc     service.ArticleService
	l       logger.LoggerV1
	intrSvc service.InteractiveService
	biz     string
}

func NewArticleHandler(svc service.ArticleService,
	l logger.LoggerV1, intrSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		l:       l,
		biz:     "article",
		intrSvc: intrSvc,
	}
}

func (a *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	// 修改
	//g.PUT("/")
	// 新增
	//g.POST("/")
	// g.DELETE("/a_id")

	//g.POST("/edit", h.Edit)
	//g.POST("/withdraw", h.Withdraw)
	//g.POST("/publish", h.Publish)
	g.POST("/edit", ginx.WrapBodyAndToken[ArticleReq, ijwt.UserClaims](a.Edit))
	g.POST("/withdraw", ginx.WrapBodyAndToken[WithdrawReq, ijwt.UserClaims](a.Withdraw))
	g.POST("/publish", ginx.WrapBodyAndToken[ArticleReq, ijwt.UserClaims](a.Publish))

	// 创作者的查询接口
	// 这个是获取数据的接口，理论上来说（遵循 RESTful 规范），应该是用 GET 方法
	// GET localhost/articles => List 接口
	g.POST("/list", ginx.WrapBodyAndToken[ListReq, ijwt.UserClaims](a.List))
	g.GET("/detail/:id", ginx.WrapToken[ijwt.UserClaims](a.Detail))

	pub := g.Group("/pub")
	pub.GET("/:id", ginx.WrapToken[ijwt.UserClaims](a.PubDetail))

	//pub.GET("/:id", a.PubDetail, func(ctx *gin.Context) {
	//	// 增加阅读计数。
	//	//go func() {
	//	//	// 开一个 goroutine，异步去执行
	//	//	er := a.intrSvc.IncrReadCnt(ctx, a.biz, art.Id)
	//	//	if er != nil {
	//	//		a.l.Error("增加阅读计数失败",
	//	//			logger.Int64("aid", art.Id),
	//	//			logger.Error(err))
	//	//	}
	//	//}()
	//})
	// 点赞是这个接口，取消点赞也是这个接口
	// RESTful 风格
	//pub.POST("/like/:id", ginx.WrapBodyAndToken[LikeReq,
	//	ijwt.UserClaims](h.Like))
	pub.POST("/like", ginx.WrapBodyAndToken[LikeReq,
		ijwt.UserClaims](a.Like))
	//pub.POST("/cancel_like", ginx.WrapBodyAndToken[LikeReq,
	//	ijwt.UserClaims](h.Like))
}

func (a *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq, uc ijwt.UserClaims) (ginx.Result, error) {
	// 检测输入，跳过这一步
	// 调用 svc 的代码
	id, err := a.svc.Save(ctx, req.toDomain(uc.Id))
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, fmt.Errorf("保存帖子失败 %v", err)
	}

	return ginx.Result{Data: id}, nil
}

func (a *ArticleHandler) Withdraw(ctx *gin.Context, req WithdrawReq, uc ijwt.UserClaims) (ginx.Result, error) {
	// 检测输入，跳过这一步
	// 调用 svc 的代码
	err := a.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: uc.Id,
		},
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, fmt.Errorf("保存帖子失败 %v", err)
	}

	return ginx.Result{Msg: "OK"}, nil
}

func (a *ArticleHandler) Publish(ctx *gin.Context, req ArticleReq, uc ijwt.UserClaims) (ginx.Result, error) {
	id, err := a.svc.Publish(ctx, req.toDomain(uc.Id))
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, fmt.Errorf("发表帖子失败 %v", err)
	}

	return ginx.Result{Data: id}, nil
}

func (a *ArticleHandler) List(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	res, err := a.svc.List(ctx, uc.Id, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	// 在列表页，不显示全文，只显示一个"摘要"
	// 比如说，简单的摘要就是前几句话
	// 强大的摘要是 AI 帮你生成的
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVO](res,
			func(idx int, src domain.Article) ArticleVO {
				return ArticleVO{
					Id:       src.Id,
					Title:    src.Title,
					Abstract: src.Abstract(),
					Status:   src.Status.ToUint8(),
					// 这个列表请求，不需要返回内容
					//Content: src.Content,
					// 这个是创作者看自己的文章列表，也不需要这个字段
					//Author: src.Author
					Ctime: src.Ctime.Format(time.DateTime),
					Utime: src.Utime.Format(time.DateTime),
				}
			}),
	}, nil
}

func (a *ArticleHandler) PubDetail(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		//a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, fmt.Errorf("前端输入的 ID 不对,%v", err)
	}

	var eg errgroup.Group
	var art domain.Article
	eg.Go(func() error {
		art, err = a.svc.GetPublishedById(ctx, id)
		return err
	})

	var intr domain.Interactive
	eg.Go(func() error {
		// 要在这里获得这篇文章的计数
		//uc := ctx.MustGet("users").(ijwt.UserClaims)
		// 这个地方可以容忍错误
		intr, err = a.intrSvc.Get(ctx, a.biz, id, uc.Id)
		// 这种是容错的写法
		//if err != nil {
		//	// 记录日志
		//}
		//return nil
		return err
	})

	// 在这儿等，要保证前面两个
	err = eg.Wait()
	if err != nil {
		// 代表查询出错了
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	// 增加阅读计数。
	go func() {
		// 开一个 goroutine，异步去执行
		er := a.intrSvc.IncrReadCnt(ctx, a.biz, art.Id)
		if er != nil {
			a.l.Error("增加阅读计数失败",
				logger.Int64("aid", art.Id),
				logger.Error(err))
		}
	}()

	// ctx.Set("art", art)

	// 这个功能是不是可以让前端，主动发一个 HTTP 请求，来增加一个计数？
	return ginx.Result{Data: ArticleVO{
		Id:      art.Id,
		Title:   art.Title,
		Status:  art.Status.ToUint8(),
		Content: art.Content,
		// 要把作者信息带出去
		Author:     art.Author.Name,
		Ctime:      art.Ctime.Format(time.DateTime),
		Utime:      art.Utime.Format(time.DateTime),
		Liked:      intr.Liked,
		Collected:  intr.Collected,
		LikeCnt:    intr.LikeCnt,
		ReadCnt:    intr.ReadCnt,
		CollectCnt: intr.CollectCnt,
	}}, nil
}

func (a *ArticleHandler) Detail(ctx *gin.Context, usr ijwt.UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	art, err := a.svc.GetById(ctx, id)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("获得文章信息失败", logger.Error(err))
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	// 这是不借助数据库查询来判定的方法
	if art.Author.Id != usr.Id {
		//ctx.JSON(http.StatusOK)
		// 如果公司有风控系统，这个时候就要上报这种非法访问的用户了。
		//a.l.Error("非法访问文章，创作者 ID 不匹配",
		//	logger.Int64("uid", usr.Id))
		return ginx.Result{
			Code: 4,
			// 也不需要告诉前端究竟发生了什么
			Msg: "输入有误",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", usr.Id)
	}
	return ginx.Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			// 不需要这个摘要信息
			//Abstract: art.Abstract(),
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			//Author: art.Author
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		},
	}, nil
}

func (a *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		err = a.intrSvc.Like(ctx, a.biz, req.Id, uc.Id)
	} else {
		err = a.intrSvc.CancelLike(ctx, a.biz, req.Id, uc.Id)
	}

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}
