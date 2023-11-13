.PHONY: mock
mock:
	@mockgen -source=../../internal/service/user.go -package=svcmocks -destination=../../internal/service/mocks/user.mock.go
	@mockgen -source=../../internal/service/code.go -package=svcmocks -destination=../../internal/service/mocks/code.mock.go
	@mockgen -source=../../internal/service/article.go -package=svcmocks -destination=../../internal/service/mocks/article.mock.go

	@mockgen -source=../../internal/repository/user_with_cache.go -package=svcmocks -destination=../../internal/repository/mocks/user_with_cache.mock.go
	@mockgen -source=../../internal/repository/code.go -package=svcmocks -destination=../../internal/repository/mocks/code.mock.go

	@mockgen -source=../../internal/repository/dao/user.go -package=svcmocks -destination=../../internal/repository/dao/mocks/user.mock.go

	@mockgen -source=../../internal/repository/cache/user.go -package=svcmocks -destination=../../internal/repository/cache/mocks/user.mock.go
	@mockgen -source=../../internal/repository/cache/codecache.go -package=svcmocks -destination=../../internal/repository/cache/mocks/codecache.mock.go

	@mockgen -source=../../internal/repository/article/article.go -package=svcmocks -destination=../../internal/repository/article/mocks/article.mock.go
	@mockgen -source=../../internal/repository/article/article_author.go -package=svcmocks -destination=../../internal/repository/article/mocks/article_author.mock.go
	@mockgen -source=../../internal/repository/article/article_reader.go -package=svcmocks -destination=../../internal/repository/article/mocks/article_reader.mock.go

	@mockgen -package=redismocks -destination=../../internal/repository/cache/redismocks/cmd.mock.go github.com/redis/go-redis/v9 Cmdable

	@mockgen -source=../../pkg/ratelimit/limiter.go -package=limitmocks -destination=../../pkg/ratelimit/mocks/ratelimit.mock.go

	@go mod tidy