package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	user.Created = time.Now()
	user.LastLogin = time.Now()
	result, err := r.db.Collection("user").InsertOne(ctx, user)
	if err != nil {
		return err
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	var user User
	err := r.db.Collection("user").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.db.Collection("user").FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.Collection("user").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id primitive.ObjectID, hashedPassword string) error {
	_, err := r.db.Collection("user").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"hashedPassword": hashedPassword}})
	return err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Collection("user").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"lastLogin": time.Now()}})
	return err
}
