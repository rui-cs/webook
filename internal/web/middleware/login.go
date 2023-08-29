package middleware

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rui-cs/webook/config"
)

type LoginMiddlewareBuilder struct {
}

func (l *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/users/login" || ctx.Request.URL.Path == "/users/signup" {
			return
		}

		s := sessions.Default(ctx)

		if s.Get("userID") == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		gob.Register(time.Time{}) // todo

		now := time.Now()
		updateTime := s.Get("update_time")

		if updateTime != nil {
			if _, ok := updateTime.(time.Time); !ok {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		if updateTime == nil || now.Sub(updateTime.(time.Time)) > time.Second*10 {
			s.Set("update_time", now)
			s.Options(sessions.Options{MaxAge: 60 * config.Config.ValidTime})
			err := s.Save()
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}
	}
}
