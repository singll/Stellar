package models

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	pkgerrors "github.com/StellarServer/internal/pkg/errors"
	"github.com/StellarServer/internal/pkg/logger"
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
		logger.Error("CreateUser user already exists", map[string]interface{}{"username": username, "email": email})
		return nil, pkgerrors.NewAppError(pkgerrors.CodeConflict, "用户名或邮箱已存在", 409)
	} else if err != mongo.ErrNoDocuments {
		logger.Error("CreateUser check existing user failed", map[string]interface{}{"username": username, "email": email, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "检查用户是否存在")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("CreateUser password hash failed", map[string]interface{}{"username": username, "error": err})
		return nil, pkgerrors.WrapError(err, pkgerrors.CodeInternalError, "密码加密失败", 500)
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
		logger.Error("CreateUser insert user failed", map[string]interface{}{"username": username, "error": err})
		return nil, pkgerrors.WrapDatabaseError(err, "创建用户")
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(db *mongo.Database, username string) (*UserMongo, error) {
	var user UserMongo
	err := db.Collection("user").FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		logger.Error("GetUserByUsername failed", map[string]interface{}{"username": username, "error": err})
		if err == mongo.ErrNoDocuments {
			return nil, pkgerrors.NewNotFoundError("user not found")
		}
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeInternalError, "GetUserByUsername failed", 500, err)
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func GetUserByEmail(db *mongo.Database, email string) (*UserMongo, error) {
	var user UserMongo
	err := db.Collection("user").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		logger.Error("GetUserByEmail failed", map[string]interface{}{"email": email, "error": err})
		if err == mongo.ErrNoDocuments {
			return nil, pkgerrors.NewNotFoundError("user not found")
		}
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeInternalError, "GetUserByEmail failed", 500, err)
	}
	return &user, nil
}

// ValidateUser 验证用户凭据
func ValidateUser(db *mongo.Database, identifier, password string) (*UserMongo, error) {
	var user UserMongo
	filter := bson.M{"$or": []bson.M{{"username": identifier}, {"email": identifier}}}
	err := db.Collection("user").FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		logger.Error("ValidateUser failed", map[string]interface{}{"identifier": identifier, "error": err})
		if err == mongo.ErrNoDocuments {
			return nil, pkgerrors.NewInvalidCredentialsError()
		}
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeInternalError, "ValidateUser failed", 500, err)
	}

	// 检查密码是否匹配
	// 支持两种加密方式：MD5（向后兼容）和bcrypt（新用户）
	passwordValid := false

	// 首先尝试bcrypt验证（新用户）
	if len(user.HashedPassword) > 0 {
		if bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)) == nil {
			passwordValid = true
		}
	}

	// 如果bcrypt验证失败，尝试MD5验证（向后兼容）
	if !passwordValid {
		// 检查是否是MD5格式（32位十六进制）
		if len(user.HashedPassword) == 32 {
			// 计算输入密码的MD5
			hashedInput := fmt.Sprintf("%x", md5.Sum([]byte(password)))
			if hashedInput == user.HashedPassword {
				passwordValid = true

				// 如果是MD5密码，自动升级为bcrypt
				go func() {
					if newHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err == nil {
						_, updateErr := db.Collection("user").UpdateOne(
							context.Background(),
							bson.M{"_id": user.ID},
							bson.M{"$set": bson.M{"hashedPassword": string(newHash)}},
						)
						if updateErr != nil {
							logger.Error("Failed to upgrade password to bcrypt", map[string]interface{}{
								"userID": user.ID.Hex(),
								"error":  updateErr,
							})
						} else {
							logger.Info("Password upgraded to bcrypt", map[string]interface{}{
								"userID": user.ID.Hex(),
							})
						}
					}
				}()
			}
		}
	}

	if !passwordValid {
		logger.Warn("ValidateUser password mismatch", map[string]interface{}{"identifier": identifier})
		return nil, pkgerrors.NewInvalidCredentialsError()
	}

	// 更新最后登录时间
	_, err = db.Collection("user").UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"lastLogin": time.Now()}},
	)
	if err != nil {
		logger.Error("Update lastLogin failed", map[string]interface{}{"userID": user.ID.Hex(), "error": err})
		return nil, pkgerrors.NewAppErrorWithCause(pkgerrors.CodeInternalError, "Update lastLogin failed", 500, err)
	}
	return &user, nil
}

// UpdatePassword 更新用户密码
func UpdatePassword(db *mongo.Database, username, oldPassword, newPassword string) error {
	user, err := GetUserByUsername(db, username)
	if err != nil {
		logger.Error("UpdatePassword get user failed", map[string]interface{}{"username": username, "error": err})
		return pkgerrors.WrapError(err, pkgerrors.CodeNotFound, "用户不存在", 404)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(oldPassword)) != nil {
		logger.Error("UpdatePassword old password incorrect", map[string]interface{}{"username": username})
		return pkgerrors.NewAppError(pkgerrors.CodeValidationFailed, "原密码错误", 400)
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("UpdatePassword new password hash failed", map[string]interface{}{"username": username, "error": err})
		return pkgerrors.WrapError(err, pkgerrors.CodeInternalError, "新密码加密失败", 500)
	}
	_, err = db.Collection("user").UpdateOne(
		context.Background(),
		bson.M{"username": username},
		bson.M{"$set": bson.M{"hashedPassword": string(hashed)}},
	)
	if err != nil {
		logger.Error("UpdatePassword update password failed", map[string]interface{}{"username": username, "error": err})
		return pkgerrors.WrapDatabaseError(err, "更新用户密码")
	}
	return nil
}
