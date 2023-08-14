package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rui-cs/webook/internal/web"
)

type LoginJWTMiddlewareBuilder struct {
	Path []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePath(path string) *LoginJWTMiddlewareBuilder {
	l.Path = append(l.Path, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, p := range l.Path {
			if ctx.Request.URL.Path == p {
				return
			}
		}

		authorization := ctx.GetHeader("Authorization")
		if authorization == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		segs := strings.Split(authorization, " ") // strings.SplitN

		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(segs[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"), nil
		})

		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()

		if claims.ExpiresAt.Sub(now) < time.Second*50 { // 超过10s了
			claims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Minute))

			var tokenStr string
			if tokenStr, err = token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0")); err != nil {
				log.Println("jwt 续约失败", err) // 记录日志
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			ctx.Header("x-jwt-token", tokenStr)
		}

		ctx.Set("claims", claims)
	}
}
