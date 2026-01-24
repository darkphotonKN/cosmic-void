package config

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Mode string // "standalone", "sentinel", "cluster"

	Addrs    []string
	Password string
	DB       int

	// connection pool settings
	PoolSize     int
	MinIdleConns int
	MaxIdleConns int

	// Sentinel specific settings
	MasterName string // only sentinel mode need
}

var once sync.Once
var globalClient redis.UniversalClient

func InitRedis(config RedisConfig) error {
	var err error

	once.Do(func() {
		switch config.Mode {
		case "cluster":
			slog.Info("Initializing Redis in CLUSTER mode")
			globalClient = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:           config.Addrs,
				Password:        config.Password,
				PoolSize:        config.PoolSize,
				MinIdleConns:    config.MinIdleConns,
				ConnMaxIdleTime: 5 * time.Minute,
				ConnMaxLifetime: 1 * time.Hour,
				DialTimeout:     5 * time.Second,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    3 * time.Second,
				PoolTimeout:     4 * time.Second,
				MaxRedirects:    3,
				MaxRetries:      3,
				MinRetryBackoff: 8 * time.Millisecond,
				MaxRetryBackoff: 512 * time.Millisecond,
				RouteByLatency:  false,
				RouteRandomly:   false,
			})
			slog.Info("Cluster nodes", "addresses", config.Addrs)

		// ========== Sentinel 模式 ==========
		case "sentinel":
			slog.Info("Initializing Redis in SENTINEL mode")
			globalClient = redis.NewFailoverClient(&redis.FailoverOptions{
				// specify sentinel settings
				MasterName:       config.MasterName,
				SentinelAddrs:    config.Addrs,
				Password:         config.Password,
				SentinelPassword: config.Password,
				DB:               config.DB,

				PoolSize:        config.PoolSize,
				MinIdleConns:    config.MinIdleConns,
				ConnMaxIdleTime: 5 * time.Minute,
				ConnMaxLifetime: 1 * time.Hour,
				DialTimeout:     5 * time.Second,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    3 * time.Second,
				PoolTimeout:     4 * time.Second,
				MaxRetries:      3,
				MinRetryBackoff: 8 * time.Millisecond,
				MaxRetryBackoff: 512 * time.Millisecond,
			})

			slog.Info("Sentinel nodes", "addresses", config.Addrs, "masterName", config.MasterName)

		default: // standalone
			slog.Info("Initializing Redis in STANDALONE mode")

			if len(config.Addrs) == 0 {
				err = fmt.Errorf("standalone mode requires at least one address")
				return
			}

			globalClient = redis.NewClient(&redis.Options{
				Addr:            config.Addrs[0],
				Password:        config.Password,
				DB:              config.DB,
				PoolSize:        config.PoolSize,
				MinIdleConns:    config.MinIdleConns,
				ConnMaxIdleTime: 5 * time.Minute,
				ConnMaxLifetime: 1 * time.Hour,
				DialTimeout:     5 * time.Second,
				ReadTimeout:     3 * time.Second,
				WriteTimeout:    3 * time.Second,
				PoolTimeout:     4 * time.Second,
				MaxRetries:      3,
				MinRetryBackoff: 8 * time.Millisecond,
				MaxRetryBackoff: 512 * time.Millisecond,
			})

			slog.Info("Standalone Redis", "address", config.Addrs[0], "db", config.DB)
		}

		// test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if pingErr := globalClient.Ping(ctx).Err(); pingErr != nil {
			err = fmt.Errorf("redis ping failed: %w", pingErr)
			return
		}

		slog.Info("Redis connection successful!")
	})

	return err
}

func GetClient() redis.UniversalClient {
	if globalClient == nil {
		panic("Redis client is not initialized. Call InitRedis first.")
	}
	return globalClient
}

func CloseRedis() error {
	if globalClient != nil {
		return globalClient.Close()
	}
	return nil
}
