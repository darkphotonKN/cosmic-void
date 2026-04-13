package cache

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	Close() error
	Incr(ctx context.Context, key string) (int64, error)
	// Redis List 操作
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (string, error)
	LLen(ctx context.Context, key string) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	LRem(ctx context.Context, key string, count int64, value interface{}) error

	// 分佈式鎖
	// AcquireLock 獲取分佈式鎖
	// 參數:
	//   - key: 鎖的鍵名
	//   - ttl: 鎖的過期時間
	// 返回:
	//   - lockID: 鎖的唯一標識符，用於釋放鎖
	//   - acquired: 是否成功獲取鎖
	//   - err: 錯誤信息
	AcquireLock(ctx context.Context, key string, ttl time.Duration) (lockID string, acquired bool, err error)

	// ReleaseLock 釋放分佈式鎖
	// 參數:
	//   - key: 鎖的鍵名
	//   - lockID: 獲取鎖時返回的標識符
	// 返回:
	//   - err: 錯誤信息（如果鎖不存在或已被其他實例持有）
	ReleaseLock(ctx context.Context, key string, lockID string) error
}
