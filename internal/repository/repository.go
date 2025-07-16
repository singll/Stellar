package repository

import (
	"fmt"

	"github.com/StellarServer/internal/database"
	"github.com/StellarServer/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// Repository provides access to storage
type Repository struct {
	db     *database.DB
	User   models.UserRepository
	// Add other repositories here
	// Project ProjectRepository
	// Asset   AssetRepository
}

// NewRepository creates a new repository
func NewRepository(db *database.DB) *Repository {
	repo := &Repository{
		db: db,
	}

	// Initialize repositories based on available databases
	if db.GORM != nil {
		repo.User = models.NewSQLUserRepository(db.GORM)
	}
	// Add MongoDB repositories for backward compatibility if needed

	return repo
}

// Transaction executes a function within a database transaction
func (r *Repository) Transaction(fn func(*Repository) error) error {
	if r.db.GORM != nil {
		return r.db.GORM.Transaction(func(tx *gorm.DB) error {
			txRepo := &Repository{
				db:   r.db,
				User: models.NewSQLUserRepository(tx),
			}
			return fn(txRepo)
		})
	}

	// For MongoDB, use the MongoDB transaction if available
	if r.db.MongoDB != nil {
		return r.db.MongoDB.Transaction(func(sessCtx mongo.SessionContext) error {
			return fn(r) // TODO: Implement MongoDB transaction support
		})
	}

	return fmt.Errorf("no database connection available for transaction")
}

// Health checks the health of the repository connections
func (r *Repository) Health() map[string]error {
	return r.db.Health()
}

// Close closes the repository connections
func (r *Repository) Close() error {
	return r.db.Close()
}