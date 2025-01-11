package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xmualex2023/i18n-translation/internal/pkg/llm"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 只处理第一个错误
		err := c.Errors.Last()
		if err == nil {
			return
		}

		// 根据错误类型返回不同的状态码
		var httpCode int
		var message string

		switch err.Err {
		case llm.ErrInvalidResponse:
			httpCode = http.StatusBadGateway
			message = "translation service response invalid"
		case llm.ErrAPIError:
			httpCode = http.StatusServiceUnavailable
			message = "translation service temporarily unavailable"
		default:
			httpCode = http.StatusInternalServerError
			message = err.Error()
		}

		c.JSON(httpCode, gin.H{
			"error": message,
		})
	}
}
