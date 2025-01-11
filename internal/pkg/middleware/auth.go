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
		authHeader := c.GetHeader(authorizationHeaderKey)
		if len(authHeader) < 7 || !strings.HasPrefix(authHeader, authorizationTypeBearer) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no valid auth token provided"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		claims, err := maker.VerifyToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
			c.Abort()
			return
		}

		// store user info in context
		c.Set("user_id", claims.UserID.Hex())
		c.Set("claims", claims)

		c.Next()
	}
}

// GetCurrentUser get current user id
func GetCurrentUser(ctx *gin.Context) (*auth.Claims, bool) {
	payload, exists := ctx.Get(authorizationPayloadKey)
	if !exists {
		return nil, false
	}

	claims, ok := payload.(*auth.Claims)
	return claims, ok
}
