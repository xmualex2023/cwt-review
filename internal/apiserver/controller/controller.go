package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/xmualex2023/i18n-translation/internal/apiserver/service"
)

// Controller interface
type IController interface {
	// user related
	Register(ctx *gin.Context)
	Login(ctx *gin.Context)
	RefreshToken(ctx *gin.Context)

	// task related
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
