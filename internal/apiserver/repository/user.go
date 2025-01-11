package repository

import (
	"context"
	"time"

	"github.com/xmualex2023/i18n-translation/internal/apiserver/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const userCollection = "users"

// CreateUser 创建用户
func (r *Repository) CreateUser(ctx context.Context, user *model.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	collection := r.db.Collection(userCollection)
	_, err := collection.InsertOne(ctx, user)
	return err
}

// GetUserByUsername 通过用户名获取用户
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	collection := r.db.Collection(userCollection)

	var user model.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}
