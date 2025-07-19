package database

import (
	"fmt"
	"time"

	"github.com/StellarServer/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB holds database connections
type DB struct {
	GORM *gorm.DB
	// Keep MongoDB connection for backward compatibility
	MongoDB *MongoDBManager
	Redis   *RedisManager
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string `yaml:"type" json:"type"` // mysql, postgres, sqlite, mongodb
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"`
	Path     string `yaml:"path" json:"path"` // for sqlite
}

// NewDB creates a new database manager
func NewDB(cfg *config.Config) (*DB, error) {
	db := &DB{}

	// Initialize GORM if SQL database is configured
	if cfg.Database.Type != "" && cfg.Database.Type != "mongodb" {
		gormDB, err := initGORM(DatabaseConfig{
			Type:     cfg.Database.Type,
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			Database: cfg.Database.Database,
			Username: cfg.Database.Username,
			Password: cfg.Database.Password,
			SSLMode:  cfg.Database.SSLMode,
			Path:     cfg.Database.Path,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to initialize GORM: %w", err)
		}

		db.GORM = gormDB
	}

	// Initialize MongoDB (for backward compatibility)
	if cfg.MongoDB.URI != "" {
		mongoManager, err := NewMongoDBManager(cfg.MongoDB)
		if err != nil {
			// 记录错误但不中断程序启动
			fmt.Printf("Warning: Failed to initialize MongoDB: %v\n", err)
		} else {
			db.MongoDB = mongoManager
		}
	}

	// Initialize Redis
	if cfg.Redis.Addr != "" {
		redisManager, err := NewRedisManager(cfg.Redis)
		if err != nil {
			// 记录错误但不中断程序启动
			fmt.Printf("Warning: Failed to initialize Redis: %v\n", err)
		} else {
			db.Redis = redisManager
		}
	}

	return db, nil
}

// initGORM initializes GORM based on database type
func initGORM(cfg DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		dialector = mysql.Open(dsn)

	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
			cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port, cfg.SSLMode)
		dialector = postgres.Open(dsn)

	case "sqlite":
		dialector = sqlite.Open(cfg.Path)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// GORM configuration
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// Close closes all database connections
func (db *DB) Close() error {
	var errors []error

	// Close GORM connection
	if db.GORM != nil {
		sqlDB, err := db.GORM.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close GORM connection: %w", err))
			}
		}
	}

	// Close MongoDB connection
	if db.MongoDB != nil {
		if err := db.MongoDB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close MongoDB connection: %w", err))
		}
	}

	// Close Redis connection
	if db.Redis != nil {
		if err := db.Redis.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close Redis connection: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}

	return nil
}

// AutoMigrate runs auto migration for given models
func (db *DB) AutoMigrate(models ...interface{}) error {
	if db.GORM == nil {
		return fmt.Errorf("GORM is not initialized")
	}
	return db.GORM.AutoMigrate(models...)
}

// Health checks the health of all database connections
func (db *DB) Health() map[string]error {
	health := make(map[string]error)

	// Check GORM connection
	if db.GORM != nil {
		sqlDB, err := db.GORM.DB()
		if err != nil {
			health["gorm"] = err
		} else {
			health["gorm"] = sqlDB.Ping()
		}
	}

	// Check MongoDB connection
	if db.MongoDB != nil {
		health["mongodb"] = db.MongoDB.Health()
	}

	// Check Redis connection
	if db.Redis != nil {
		health["redis"] = db.Redis.Health()
	}

	return health
}
