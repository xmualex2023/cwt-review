package service

import (
	"context"
	"errors"
	"time"

	"github.com/xmualex2023/i18n-translation/internal/apiserver/model"
	"github.com/xmualex2023/i18n-translation/internal/pkg/queue"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrTaskNotFound = errors.New("任务不存在")
	ErrInvalidTask  = errors.New("无效的任务")
)

// CreateTask 创建翻译任务
func (s *Service) CreateTask(ctx context.Context, req *model.CreateTaskRequest, userID primitive.ObjectID) (*model.TaskResponse, error) {
	task := &model.Task{
		UserID:        userID,
		Status:        model.TaskStatusPending,
		SourceLang:    req.SourceLang,
		TargetLang:    req.TargetLang,
		SourceContent: req.SourceContent,
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	return &model.TaskResponse{
		ID:        task.ID.Hex(),
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}, nil
}

// ExecuteTranslation 执行翻译任务
func (s *Service) ExecuteTranslation(ctx context.Context, taskID string) error {
	id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return ErrInvalidTask
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	// 更新任务状态为处理中
	task.Status = model.TaskStatusProcessing
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return err
	}

	// 创建翻译任务并加入队列
	translationTask := &model.TranslationTask{
		ID:            task.ID.Hex(),
		UserID:        task.UserID.Hex(),
		SourceLang:    task.SourceLang,
		TargetLang:    task.TargetLang,
		SourceContent: task.SourceContent,
		CreatedAt:     time.Now(),
	}

	if err := s.queue.Enqueue(ctx, translationTask); err != nil {
		task.Status = model.TaskStatusFailed
		task.Error = "加入队列失败"
		_ = s.repo.UpdateTask(ctx, task)
		return err
	}

	return nil
}

// GetTaskStatus 获取任务状态
func (s *Service) GetTaskStatus(ctx context.Context, taskID string) (*model.TaskResponse, error) {
	id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, ErrInvalidTask
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	return &model.TaskResponse{
		ID:        task.ID.Hex(),
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
		Error:     task.Error,
	}, nil
}

// HandleTranslationTask 处理翻译任务
func (s *Service) HandleTranslationTask(ctx context.Context, t queue.Task) error {
	task, ok := t.(*model.TranslationTask)
	if !ok {
		return errors.New("无效的任务类型")
	}

	// 获取数据库中的任务
	id, err := primitive.ObjectIDFromHex(task.ID)
	if err != nil {
		return err
	}

	dbTask, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return err
	}
	if dbTask == nil {
		return ErrTaskNotFound
	}

	// 执行翻译
	translatedText, err := s.translator.Translate(ctx, task.SourceContent, task.SourceLang, task.TargetLang)
	if err != nil {
		dbTask.Status = model.TaskStatusFailed
		dbTask.Error = err.Error()
	} else {
		dbTask.Status = model.TaskStatusCompleted
		dbTask.ResultContent = translatedText
	}

	// 更新任务状态
	return s.repo.UpdateTask(ctx, dbTask)
}

// GetTranslation 获取翻译结果
func (s *Service) GetTranslation(ctx context.Context, taskID string) (string, error) {
	id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return "", ErrInvalidTask
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return "", err
	}
	if task == nil {
		return "", ErrTaskNotFound
	}

	if task.Status != model.TaskStatusCompleted {
		return "", errors.New("任务未完成")
	}

	if task.ResultContent == "" {
		return "", errors.New("翻译结果不存在")
	}

	return task.ResultContent, nil
}
