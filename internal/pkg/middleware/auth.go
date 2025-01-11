package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xmualex2023/i18n-translation/internal/pkg/auth"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "Bearer"
	authorizationPayloadKey = "authorization_payload"
)

func AuthMiddleware(maker *auth.JWTMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 7 || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供有效的认证令牌"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		claims, err := maker.VerifyToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID.Hex())
		c.Set("claims", claims)

		c.Next()
	}
}

// GetCurrentUser 获取当前用户ID
func GetCurrentUser(ctx *gin.Context) (*auth.Claims, bool) {
	payload, exists := ctx.Get(authorizationPayloadKey)
	if !exists {
		return nil, false
	}

	claims, ok := payload.(*auth.Claims)
	return claims, ok
}
