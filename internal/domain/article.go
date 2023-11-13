package domain

import "time"

// Article 可以同时表达线上库和制作库的概念吗？
// 可以同时表达，作者眼中的 Article 和读者眼中的 Article 吗？
type Article struct {
	Id      int64
	Title   string
	Content string
	// Author 要从用户来
	Author Author
	Status ArticleStatus
	Ctime  time.Time
	Utime  time.Time

	// 做成这样，就应该在 service 或者 repository 里面完成构造
	// 设计成这个样子，就认为 Interactive 是 Article 的一个属性（值对象）
	// Intr Interactive
	//
}

func (a Article) Abstract() string {
	// 摘要我们取前几句。
	// 要考虑一个中文问题
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	// 英文怎么截取一个完整的单词，我的看法是……不需要纠结，就截断拉到
	// 词组、介词，往后找标点符号
	return string(cs[:100])
}

type ArticleStatus uint8

const (
	// ArticleStatusUnknown 为了避免零值之类的问题
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

func (s ArticleStatus) NonPublished() bool {
	return s != ArticleStatusPublished
}

func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusPrivate:
		return "private"
	case ArticleStatusUnpublished:
		return "unpublished"
	case ArticleStatusPublished:
		return "published"
	default:
		return "unknown"
	}
}

// ArticleStatusV1 如果你的状态很复杂，有很多行为（就是你要搞很多方法），状态里面需要一些额外字段
// 就用这个版本
type ArticleStatusV1 struct {
	Val  uint8
	Name string
}

var (
	ArticleStatusV1Unknown = ArticleStatusV1{Val: 0, Name: "unknown"}
)

type ArticleStatusV2 string

// Author 在帖子这个领域内，是一个值对象
type Author struct {
	Id   int64
	Name string
}

//type AuthorV1 struct {
//	articles []Article
//}
