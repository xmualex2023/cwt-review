package repository

import (
	"context"
	"time"

	"github.com/xmualex2023/i18n-translation/internal/apiserver/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const taskCollection = "tasks"

// CreateTask 创建翻译任务
func (r *Repository) CreateTask(ctx context.Context, task *model.Task) error {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	collection := r.db.Collection(taskCollection)
	_, err := collection.InsertOne(ctx, task)
	return err
}

// GetTask 获取任务信息
func (r *Repository) GetTask(ctx context.Context, taskID primitive.ObjectID) (*model.Task, error) {
	collection := r.db.Collection(taskCollection)

	var task model.Task
	err := collection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &task, err
}

// UpdateTask 更新任务状态
func (r *Repository) UpdateTask(ctx context.Context, task *model.Task) error {
	task.UpdatedAt = time.Now()

	collection := r.db.Collection(taskCollection)
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": task.ID},
		bson.M{"$set": task},
	)
	return err
}
