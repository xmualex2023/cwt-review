package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/xmualex2023/i18n-translation/internal/apiserver/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const userCollection = "users"

func (r *Repository) CreateUser(ctx context.Context, user *model.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	collection := r.db.Collection(userCollection)
	_, err := collection.InsertOne(ctx, user)
	return err
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	collection := r.db.Collection(userCollection)

	var user model.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("user not found, username: %s", username)
	}
	return &user, err
}

func (r *Repository) UpdateUser(ctx context.Context, user *model.User) error {
	collection := r.db.Collection(userCollection)
	_, err := collection.UpdateOne(ctx, bson.M{"username": user.Username}, bson.M{"$set": user})
	return err
}

func (r *Repository) DeleteUser(ctx context.Context, username string) error {
	collection := r.db.Collection(userCollection)
	_, err := collection.DeleteOne(ctx, bson.M{"username": username})
	return err
}
