package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"    // 等待中
	TaskStatusProcessing TaskStatus = "processing" // 处理中
	TaskStatusCompleted  TaskStatus = "completed"  // 已完成
	TaskStatusFailed     TaskStatus = "failed"     // 失败
)

// Task 翻译任务模型
type Task struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Status        TaskStatus         `bson:"status" json:"status"`
	SourceLang    string             `bson:"source_lang" json:"source_lang"`
	TargetLang    string             `bson:"target_lang" json:"target_lang"`
	SourceContent string             `bson:"source_content" json:"source_content"`
	ResultContent string             `bson:"result_content,omitempty" json:"result_content,omitempty"`
	Error         string             `bson:"error,omitempty" json:"error,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	SourceLang    string `json:"source_lang" binding:"required"`
	TargetLang    string `json:"target_lang" binding:"required"`
	SourceContent string `json:"source_content" binding:"required"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	ID        string     `json:"id"`
	Status    TaskStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Error     string     `json:"error,omitempty"`
}
