package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xmualex2023/i18n-translation/internal/pkg/limiter"
)

func RateLimiter(limiter *limiter.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户标识，可以是用户ID或IP地址
		userID := c.GetString("user_id")
		if userID == "" {
			userID = c.ClientIP()
		}

		// 构造限流 key
		key := "rate_limit:" + userID

		allowed, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "速率限制服务异常"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}

		c.Next()
	}
}
