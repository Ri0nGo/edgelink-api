package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const oauthOpenIDKeyPrefix = "edgelink:oauth:openid:"

var authExcludePaths = map[string]struct{}{
	"GET /api/edgelink/oauth/info":           {},
	"POST /api/edgelink/oauth/token":         {},
	"POST /api/edgelink/oauth/refresh_token": {},
	"GET /api/edgelink/oauth/userinfo":       {},
}

func Auth(redis redis.Cmdable) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := authExcludePaths[c.Request.Method+" "+c.FullPath()]; ok {
			c.Next()
			return
		}

		if redis == nil {
			abortUnauthorized(c)
			return
		}

		token := strings.TrimSpace(c.GetHeader("Authorization"))
		if token == "" {
			abortUnauthorized(c)
			return
		}

		openID, err := redis.Get(c.Request.Context(), oauthOpenIDKeyPrefix+token).Result()
		if err != nil || strings.TrimSpace(openID) == "" {
			abortUnauthorized(c)
			return
		}

		c.Set("openid", openID)
		c.Next()
	}
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code": http.StatusUnauthorized,
		"msg":  "未登录或登录已过期",
		"data": nil,
	})
}
