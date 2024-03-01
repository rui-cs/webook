package errs

const (
	// CommonInvalidInput 任何模块都可以使用的表达输入错误
	CommonInvalidInput   = 400001
	CommonInternalServer = 500001
)

// 用户模块
const (
	// UserInvalidInput 用户模块输入错误，这是一个含糊的错误
	UserInvalidInput        = 401001
	UserInternalServerError = 501001
	// UserInvalidOrPassword 用户不存在或者密码错误，这个你要小心，
	// 防止有人跟你过不去
	UserInvalidOrPassword = 401002
)

const (
	ArticleInvalidInput        = 402001
	ArticleInternalServerError = 502001
)

var (
	// UserInvalidInputV1 这个东西是你 DEBUG 用的，不是给 C 端用户用的
	UserInvalidInputV1 = Code{
		Number: 401001,
		Msg:    "用户输入错误",
	}
)

type Code struct {
	Number int
	Msg    string
}
