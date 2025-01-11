package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xmualex2023/i18n-translation/internal/apiserver/model"
)

// Register 用户注册
func (c *Controller) Register(ctx *gin.Context) {
	var req model.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.svc.Register(ctx.Request.Context(), &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "注册成功"})
}

// Login 用户登录
func (c *Controller) Login(ctx *gin.Context) {
	var req model.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.svc.Login(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RefreshToken 刷新令牌
func (c *Controller) RefreshToken(ctx *gin.Context) {
	// 从请求头获取旧令牌
	authHeader := ctx.GetHeader("Authorization")
	if len(authHeader) < 7 || !strings.HasPrefix(authHeader, "Bearer ") {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未提供有效的认证令牌"})
		return
	}

	oldToken := authHeader[7:]
	resp, err := c.svc.RefreshToken(ctx.Request.Context(), oldToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
