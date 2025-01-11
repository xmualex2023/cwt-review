package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/xmualex2023/i18n-translation/internal/apiserver/service"
)

// Controller 接口定义
type IController interface {
	// 用户相关
	Register(ctx *gin.Context)
	Login(ctx *gin.Context)
	RefreshToken(ctx *gin.Context)

	// 任务相关
	CreateTask(ctx *gin.Context)
	ExecuteTranslation(ctx *gin.Context)
	GetTaskStatus(ctx *gin.Context)
	DownloadTranslation(ctx *gin.Context)
}

type Controller struct {
	svc *service.Service
}

func NewController(svc *service.Service) IController {
	return &Controller{
		svc: svc,
	}
}
