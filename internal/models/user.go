package models

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// UserMongo 用户模型 (MongoDB版本，向后兼容)
type UserMongo struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username       string             `bson:"username" json:"username"`
	Email          string             `bson:"email" json:"email"`
	HashedPassword string             `bson:"hashedPassword" json:"hashedPassword,omitempty"`
	Roles          []string           `bson:"roles" json:"roles"`
	Created        time.Time          `bson:"created" json:"created"`
	LastLogin      time.Time          `bson:"lastLogin" json:"lastLogin"`
}

// CreateUser 创建用户
func CreateUser(db *mongo.Database, username, email, password string, roles []string) (*UserMongo, error) {
	// 检查用户名或邮箱是否已存在
	var existingUser UserMongo
	err := db.Collection("user").FindOne(context.Background(), bson.M{"$or": []bson.M{{"username": username}, {"email": email}}}).Decode(&existingUser)
	if err == nil {
		return nil, fmt.Errorf("用户名或邮箱已存在")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	user := &UserMongo{
		Username:       username,
		Email:          email,
		HashedPassword: string(hashed),
		Roles:          roles,
		Created:        time.Now(),
		LastLogin:      time.Now(),
	}

	result, err := db.Collection("user").InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(db *mongo.Database, username string) (*UserMongo, error) {
	var user UserMongo
	err := db.Collection("user").FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func GetUserByEmail(db *mongo.Database, email string) (*UserMongo, error) {
	var user UserMongo
	err := db.Collection("user").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ValidateUser 验证用户凭证（用户名或邮箱+密码）
func ValidateUser(db *mongo.Database, identifier, password string) (*UserMongo, error) {
	var user UserMongo
	filter := bson.M{"$or": []bson.M{{"username": identifier}, {"email": identifier}}}
	err := db.Collection("user").FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("用户名/邮箱或密码错误")
		}
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)) != nil {
		return nil, fmt.Errorf("用户名/邮箱或密码错误")
	}
	// 更新最后登录时间
	_, err = db.Collection("user").UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"lastLogin": time.Now()}},
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePassword 更新用户密码
func UpdatePassword(db *mongo.Database, username, oldPassword, newPassword string) error {
	user, err := GetUserByUsername(db, username)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(oldPassword)) != nil {
		return fmt.Errorf("原密码错误")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("新密码加密失败: %v", err)
	}
	_, err = db.Collection("user").UpdateOne(
		context.Background(),
		bson.M{"username": username},
		bson.M{"$set": bson.M{"hashedPassword": string(hashed)}},
	)
	return err
}
