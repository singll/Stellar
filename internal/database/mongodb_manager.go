package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/StellarServer/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDBManager manages MongoDB connections
type MongoDBManager struct {
	client   *mongo.Client
	database *mongo.Database
	config   config.MongoDBConfig
}

// NewMongoDBManager creates a new MongoDB manager
func NewMongoDBManager(cfg config.MongoDBConfig) (*MongoDBManager, error) {
	manager := &MongoDBManager{
		config: cfg,
	}

	client, err := manager.connect()
	if err != nil {
		return nil, err
	}

	manager.client = client
	manager.database = client.Database(cfg.Database)

	return manager, nil
}

// connect establishes MongoDB connection
func (m *MongoDBManager) connect() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := m.config.URI
	if m.config.Username != "" && m.config.Password != "" {
		uri = m.addAuthToURI(m.config.URI, m.config.Username, m.config.Password)
	}

	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(m.config.MaxPoolSize).
		SetMinPoolSize(m.config.MinPoolSize).
		SetMaxConnIdleTime(time.Duration(m.config.MaxIdleTimeMS) * time.Millisecond)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

// addAuthToURI adds authentication to MongoDB URI
func (m *MongoDBManager) addAuthToURI(uri, user, pass string) string {
	if !strings.HasPrefix(uri, "mongodb://") {
		return uri
	}
	uri = strings.TrimPrefix(uri, "mongodb://")
	return "mongodb://" + user + ":" + pass + "@" + uri
}

// GetDatabase returns the MongoDB database instance
func (m *MongoDBManager) GetDatabase() *mongo.Database {
	return m.database
}

// GetClient returns the MongoDB client instance
func (m *MongoDBManager) GetClient() *mongo.Client {
	return m.client
}

// Collection returns a collection from the database
func (m *MongoDBManager) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

// Health checks MongoDB connection health
func (m *MongoDBManager) Health() error {
	if m.client == nil {
		return fmt.Errorf("MongoDB client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.client.Ping(ctx, readpref.Primary())
}

// Close closes MongoDB connection
func (m *MongoDBManager) Close() error {
	if m.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := m.client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
			return err
		}
	}
	return nil
}

// Transaction executes a function within a MongoDB transaction
func (m *MongoDBManager) Transaction(fn func(sessCtx mongo.SessionContext) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})

	return err
}