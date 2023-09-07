.PHONY: mock
mock:
	@mockgen -source=../../internal/service/user.go -package=svcmocks -destination=../../internal/service/mocks/user.mock.go
	@mockgen -source=../../internal/service/code.go -package=svcmocks -destination=../../internal/service/mocks/code.mock.go
	@mockgen -source=../../internal/repository/user_with_cache.go -package=svcmocks -destination=../../internal/repository/mocks/user_with_cache.mock.go
	@mockgen -source=../../internal/repository/code.go -package=svcmocks -destination=../../internal/repository/mocks/code.mock.go
	@mockgen -source=../../internal/repository/dao/user.go -package=svcmocks -destination=../../internal/repository/dao/mocks/user.mock.go
	@mockgen -source=../../internal/repository/cache/user.go -package=svcmocks -destination=../../internal/repository/cache/mocks/user.mock.go
	@mockgen -source=../../internal/repository/cache/codecache.go -package=svcmocks -destination=../../internal/repository/cache/mocks/codecache.mock.go
	@go mod tidy
