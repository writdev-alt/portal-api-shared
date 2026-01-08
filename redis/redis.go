package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

type Database struct {
	*gorm.DB
}

func Setup() error {
	var client *redis.Client
	enabled := parseBoolEnv("REDIS_ENABLED", false)
	if enabled {
		host := getEnv("REDIS_HOST", "localhost")
		port := getEnv("REDIS_PORT", "6379")
		password := os.Getenv("REDIS_PASSWORD")
		db := parseIntEnv("REDIS_DB", 0)

		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: password,
			DB:       db,
		})

		if err := client.Ping(ctx).Err(); err != nil {
			return err
		}
	}

	rdb = client

	return nil
}

func IsAlive() bool {
	if rdb == nil {
		return false
	}

	return rdb.Ping(ctx).Err() == nil
}

func GetRedis() *redis.Client {
	if rdb == nil {
		panic("Redis client is not initialized. Call Setup() first.")
	}

	return rdb
}

func getEnv(key, defaultValue string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return defaultValue
}

func parseBoolEnv(key string, defaultValue bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue
	}
	if b, err := strconv.ParseBool(raw); err == nil {
		return b
	}
	return defaultValue
}

func parseIntEnv(key string, defaultValue int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue
	}
	if n, err := strconv.Atoi(raw); err == nil {
		return n
	}
	return defaultValue
}
