package config

import (
	"context"
	"fmt"
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
			fmt.Println("Initializing Redis in CLUSTER mode")
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
			fmt.Printf("Cluster nodes: %v\n", config.Addrs)

		// ========== Sentinel 模式 ==========
		case "sentinel":
			fmt.Println("Initializing Redis in SENTINEL mode")
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

			fmt.Printf("Sentinel nodes: %v, MasterName: %s\n", config.Addrs, config.MasterName)

		default: // standalone
			fmt.Println("Initializing Redis in STANDALONE mode")

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

			fmt.Printf("Standalone address: %s, DB: %d\n", config.Addrs[0], config.DB)
		}

		// test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if pingErr := globalClient.Ping(ctx).Err(); pingErr != nil {
			err = fmt.Errorf("redis ping failed: %w", pingErr)
			return
		}

		fmt.Println("Redis connection successful!")
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
