package cache

import (
	context "context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	client redis.UniversalClient
}

func NewRedisCache(client redis.UniversalClient) Cache {
	return &redisClient{client: client}
}

func (r *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *redisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (r *redisClient) Close() error {
	return r.client.Close()
}

// increments and returns new version key
func (r *redisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *redisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

func (r *redisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

func (r *redisClient) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

func (r *redisClient) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

func (r *redisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

func (r *redisClient) LRem(ctx context.Context, key string, count int64, value interface{}) error {
	return r.client.LRem(ctx, key, count, value).Err()
}

func (r *redisClient) AcquireLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	lockID := uuid.New().String()
	lockKey := fmt.Sprintf("lock:%s", key)
	success, err := r.client.SetNX(ctx, lockKey, lockID, ttl).Result()

	if err != nil {
		slog.Error("failed to acquire lock", "error", err)
		return "", false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if success {
		// TEST: uncomment if testing only
		// slog.Debug("Lock acquired successfully",
		// 	"key", key,
		// 	"lockID", lockID,
		// 	"ttl", ttl,
		// )
		return lockID, true, nil
	}

	// TEST: uncomment if testing only
	// slog.Debug("Lock already held by another instance", "key", key)
	return "", false, nil

}

// Lua 腳本：原子性地檢查並刪除鎖
// 只有當鎖的值等於 lockID 時才刪除
const releaseLockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL",KEYS[1])
else 
	return 0
end
`

func (r *redisClient) ReleaseLock(ctx context.Context, key string, lockID string) error {
	lockKey := fmt.Sprintf("lock:%s", key)
	result, err := r.client.Eval(ctx, releaseLockScript, []string{lockKey}, lockID).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	del, ok := result.(int64)
	if !ok {
		return fmt.Errorf("unexpected result type from lua script: %T", ok)
	}

	if del == 0 {
		// 鎖不存在或已被其他實例持有
		slog.Warn("Attempted to release lock that is not held or already expired",
			"key", key,
			"lockID", lockID,
		)
		return fmt.Errorf("lock not held or already expired")
	}

	// TEST: uncomment if testing only
	// slog.Debug("Lock released successfully",
	// 	"key", key,
	// 	"lockID", lockID,
	// )
	return nil
}
