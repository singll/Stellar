package models

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// UserSQL represents user model for SQL databases
type UserSQL struct {
	BaseModel
	Username       string    `gorm:"uniqueIndex;not null;size:100" json:"username"`
	Email          string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	HashedPassword string    `gorm:"not null" json:"-"`
	Roles          string    `gorm:"type:text" json:"roles"` // JSON string of roles array
	LastLogin      time.Time `json:"last_login"`
	Active         bool      `gorm:"default:true" json:"active"`
}

// TableName returns the table name for UserSQL
func (UserSQL) TableName() string {
	return "users"
}

// UserMongoDB represents user model for MongoDB (backward compatibility)
type UserMongoDB struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username       string             `bson:"username" json:"username"`
	Email          string             `bson:"email" json:"email"`
	HashedPassword string             `bson:"hashedPassword" json:"hashedPassword,omitempty"`
	Roles          []string           `bson:"roles" json:"roles"`
	Created        time.Time          `bson:"created" json:"created"`
	LastLogin      time.Time          `bson:"lastLogin" json:"lastLogin"`
	Active         bool               `bson:"active" json:"active"`
}

// User interface defines common user operations
type User interface {
	GetID() string
	GetUsername() string
	GetEmail() string
	GetRoles() []string
	SetPassword(password string) error
	ValidatePassword(password string) bool
	UpdateLastLogin() error
}

// UserRepository defines user database operations
type UserRepository interface {
	Create(user User) error
	GetByID(id string) (User, error)
	GetByUsername(username string) (User, error)
	GetByEmail(email string) (User, error)
	Update(user User) error
	Delete(id string) error
	List(limit, offset int) ([]User, error)
}

// SQLUserRepository implements UserRepository for SQL databases
type SQLUserRepository struct {
	db *gorm.DB
}

// NewSQLUserRepository creates a new SQL user repository
func NewSQLUserRepository(db *gorm.DB) *SQLUserRepository {
	return &SQLUserRepository{db: db}
}

// Create creates a new user
func (r *SQLUserRepository) Create(user User) error {
	userSQL, ok := user.(*UserSQL)
	if !ok {
		return errors.New("invalid user type for SQL repository")
	}

	// Check if username or email already exists
	var existingUser UserSQL
	err := r.db.Where("username = ? OR email = ?", userSQL.Username, userSQL.Email).First(&existingUser).Error
	if err == nil {
		return errors.New("username or email already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return r.db.Create(userSQL).Error
}

// GetByID gets user by ID
func (r *SQLUserRepository) GetByID(id string) (User, error) {
	var user UserSQL
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername gets user by username
func (r *SQLUserRepository) GetByUsername(username string) (User, error) {
	var user UserSQL
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail gets user by email
func (r *SQLUserRepository) GetByEmail(email string) (User, error) {
	var user UserSQL
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates user
func (r *SQLUserRepository) Update(user User) error {
	userSQL, ok := user.(*UserSQL)
	if !ok {
		return errors.New("invalid user type for SQL repository")
	}
	return r.db.Save(userSQL).Error
}

// Delete deletes user
func (r *SQLUserRepository) Delete(id string) error {
	return r.db.Delete(&UserSQL{}, id).Error
}

// List lists users with pagination
func (r *SQLUserRepository) List(limit, offset int) ([]User, error) {
	var users []UserSQL
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := make([]User, len(users))
	for i, user := range users {
		userCopy := user // 创建副本避免指针问题
		result[i] = &userCopy
	}
	return result, nil
}

// Implement User interface for UserSQL
func (u *UserSQL) GetID() string {
	return fmt.Sprintf("%d", u.ID)
}

func (u *UserSQL) GetUsername() string {
	return u.Username
}

func (u *UserSQL) GetEmail() string {
	return u.Email
}

func (u *UserSQL) GetRoles() []string {
	// Parse JSON string to roles array
	// This is a simplified implementation
	if u.Roles == "" {
		return []string{}
	}
	return []string{u.Roles} // TODO: Implement proper JSON parsing
}

func (u *UserSQL) SetPassword(password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashed)
	return nil
}

func (u *UserSQL) ValidatePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password)) == nil
}

func (u *UserSQL) UpdateLastLogin() error {
	u.LastLogin = time.Now()
	return nil
}