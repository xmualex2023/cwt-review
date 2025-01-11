package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xmualex2023/i18n-translation/internal/pkg/limiter"
)

func RateLimiter(limiter *limiter.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get user identifier, can be user id or ip address
		userID := c.GetString("user_id")
		if userID == "" {
			userID = c.ClientIP()
		}

		// construct rate limit key
		key := "rate_limit:" + userID

		allowed, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit service error"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "request too frequent, please try again later"})
			c.Abort()
			return
		}

		c.Next()
	}
}
