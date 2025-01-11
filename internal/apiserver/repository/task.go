package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/xmualex2023/i18n-translation/internal/apiserver/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const taskCollection = "tasks"

// CreateTask create translation task
func (r *Repository) CreateTask(ctx context.Context, task *model.Task) error {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	collection := r.db.Collection(taskCollection)
	_, err := collection.InsertOne(ctx, task)
	return err
}

// GetTask get task info
func (r *Repository) GetTask(ctx context.Context, taskID primitive.ObjectID) (*model.Task, error) {
	collection := r.db.Collection(taskCollection)

	var task model.Task
	err := collection.FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("task not found, id: %s", taskID.Hex())
	}
	return &task, err
}

// TODO: 这里可以优化，插入需要更新的字段即可
// UpdateTask update task status
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
