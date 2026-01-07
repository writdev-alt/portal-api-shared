package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	cloudsqlmysql "cloud.google.com/go/cloudsqlconn/mysql/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config represents database configuration
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	Instance string // Cloud SQL instance connection name
}

// GetConfigFromEnv loads database config from environment variables
func GetConfigFromEnv() Config {
	return Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		Username: getEnv("DB_USERNAME", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		Database: getEnv("DB_DATABASE", ""),
		Instance: getEnv("CLOUD_SQL_INSTANCE", ""),
	}
}

// Initialize initializes database connection
func Initialize(config Config) (*gorm.DB, error) {
	ctx := context.Background()

	var dsn string
	var err error

	if config.Instance != "" {
		// Use Cloud SQL Connector
		dsn, err = getCloudSQLDSN(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Cloud SQL DSN: %w", err)
		}
	} else {
		// Use regular MySQL connection
		dsn = getRegularDSN(config)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// getCloudSQLDSN creates DSN using Cloud SQL Connector
func getCloudSQLDSN(ctx context.Context, config Config) (string, error) {
	// Register the Cloud SQL driver with dialer option
	cloudsqlmysql.RegisterDriver("cloudsql-mysql", cloudsqlconn.WithDefaultDialOptions())

	// Build DSN with Cloud SQL driver
	dsn := fmt.Sprintf("%s:%s@cloudsql-mysql(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Instance, config.Database)

	return dsn, nil
}

// getRegularDSN builds regular MySQL connection string
func getRegularDSN(config Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
