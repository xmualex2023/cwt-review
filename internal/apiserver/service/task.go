package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/xmualex2023/i18n-translation/internal/apiserver/model"
	"github.com/xmualex2023/i18n-translation/internal/pkg/queue"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidTask  = errors.New("invalid task")
)

// CreateTask create translation task
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

// ExecuteTranslation execute translation task
func (s *Service) ExecuteTranslation(ctx context.Context, taskID string) error {
	id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("invalid task id: %s, error: %w", taskID, err)
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task, id: %s, error: %w", taskID, err)
	}

	// update task status to processing
	task.Status = model.TaskStatusProcessing
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task status, id: %s, error: %w", taskID, err)
	}

	// create translation task and enqueue
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
		task.Error = fmt.Sprintf("failed to enqueue, error: %v", err)
		if err := s.repo.UpdateTask(ctx, task); err != nil {
			return fmt.Errorf("failed to update task status, id: %s, error: %w", taskID, err)
		}
		return fmt.Errorf("failed to enqueue, id: %s, error: %w", taskID, err)
	}

	return nil
}

// GetTaskStatus get task status
func (s *Service) GetTaskStatus(ctx context.Context, taskID string) (*model.TaskResponse, error) {
	id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, ErrInvalidTask
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.TaskResponse{
		ID:        task.ID.Hex(),
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
		Error:     task.Error,
	}, nil
}

// HandleTranslationTask handle translation task
func (s *Service) HandleTranslationTask(ctx context.Context, t queue.Task) error {
	task, ok := t.(*model.TranslationTask)
	if !ok {
		return fmt.Errorf("invalid task type, taskID: %v", t.GetID())
	}

	// get task from db
	id, err := primitive.ObjectIDFromHex(task.ID)
	if err != nil {
		return fmt.Errorf("invalid task id, taskID: %v, error: %w", task.ID, err)
	}

	dbTask, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task, id: %s, error: %w", task.ID, err)
	}

	// execute translation
	translatedText, err := s.translator.Translate(ctx, task.SourceContent, task.SourceLang, task.TargetLang)
	if err != nil {
		dbTask.Status = model.TaskStatusFailed
		dbTask.Error = err.Error()
	} else {
		dbTask.Status = model.TaskStatusCompleted
		dbTask.ResultContent = translatedText
	}

	// update task status
	return s.repo.UpdateTask(ctx, dbTask)
}

// GetTranslation get translation result
func (s *Service) GetTranslation(ctx context.Context, taskID string) (string, error) {
	id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return "", fmt.Errorf("invalid task id, taskID: %v, error: %w", taskID, err)
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get task, id: %s, error: %w", taskID, err)
	}
	if task == nil {
		return "", fmt.Errorf("task not found, id: %s", taskID)
	}

	if task.Status != model.TaskStatusCompleted {
		return "", fmt.Errorf("task not completed, id: %s", taskID)
	}

	if task.ResultContent == "" {
		return "", fmt.Errorf("translation result not found, id: %s", taskID)
	}

	return task.ResultContent, nil
}
