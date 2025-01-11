package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xmualex2023/i18n-translation/internal/apiserver/model"
	"github.com/xmualex2023/i18n-translation/internal/pkg/middleware"
)

// CreateTask 创建翻译任务
func (c *Controller) CreateTask(ctx *gin.Context) {
	var req model.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	claims, exists := middleware.GetCurrentUser(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	resp, err := c.svc.CreateTask(ctx.Request.Context(), &req, claims.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// ExecuteTranslation 执行翻译
func (c *Controller) ExecuteTranslation(ctx *gin.Context) {
	taskID := ctx.Param("taskID")

	if err := c.svc.ExecuteTranslation(ctx.Request.Context(), taskID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "翻译任务已开始执行"})
}

// GetTaskStatus 获取任务状态
func (c *Controller) GetTaskStatus(ctx *gin.Context) {
	taskID := ctx.Param("taskID")

	resp, err := c.svc.GetTaskStatus(ctx.Request.Context(), taskID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DownloadTranslation 下载翻译结果
func (c *Controller) DownloadTranslation(ctx *gin.Context) {
	taskID := ctx.Param("taskID")

	content, err := c.svc.GetTranslation(ctx.Request.Context(), taskID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"result": content,
	})
}
