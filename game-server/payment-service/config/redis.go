package config

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

var globalRedisClient redis.UniversalClient

func InitRedis(addr, password string, db int) error {
	globalRedisClient = redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        password,
		DB:              db,
		PoolSize:        10,
		MinIdleConns:    3,
		ConnMaxIdleTime: 5 * time.Minute,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := globalRedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	slog.Info("Payment service Redis connected", "address", addr)
	return nil
}

func GetRedisClient() redis.UniversalClient {
	if globalRedisClient == nil {
		panic("Redis client is not initialized. Call InitRedis first.")
	}
	return globalRedisClient
}

func CloseRedis() error {
	if globalRedisClient != nil {
		return globalRedisClient.Close()
	}
	return nil
}
