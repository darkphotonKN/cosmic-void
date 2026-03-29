package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) Cache {
	t.Helper()
	// 使用 miniredis（內存 Redis，不需要真實 Redis）
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	cacheService := NewRedisCache(client)

	// 在測試結束時執行，不是在 helper 結束時
	// 先關閉 cache（內部會關閉 client）
	// 再關閉 miniredis 服務器
	t.Cleanup(func() {
		cacheService.Close()
		mr.Close()
	})

	return cacheService
}

func TestAcquireLock_Basic(t *testing.T) {
	cacheService := setupTestRedis(t)
	// defer cacheService.Close()

	ctx := context.Background()
	key := "test:resource"
	ttl := 10 * time.Second
	lockID, acquired, err := cacheService.AcquireLock(ctx, key, ttl)

	assert.NoError(t, err)
	assert.True(t, acquired)
	assert.NotEmpty(t, lockID)

	val, err := cacheService.Get(ctx, fmt.Sprintf("lock:%s", key))
	assert.NoError(t, err)
	assert.Equal(t, lockID, val)
}

func TestAcquireLock_AlreadyHeld(t *testing.T) {
	cacheService := setupTestRedis(t)
	// defer cacheService.Close()

	ctx := context.Background()
	key := "test:resource"
	ttl := 10 * time.Second
	lockID, acquired, err := cacheService.AcquireLock(ctx, key, ttl)
	assert.NoError(t, err)
	assert.True(t, acquired)
	assert.NotEmpty(t, lockID)

	// -- 模擬同個key重複取多次狀況
	lockID2, acquired2, err2 := cacheService.AcquireLock(ctx, key, ttl)
	assert.NoError(t, err2)
	assert.False(t, acquired2, "鎖已被持有，不應該再次獲取成功")
	assert.Empty(t, lockID2, "未獲取到鎖時，lockID 應為空")

}

func TestReleaseLock_Basic(t *testing.T) {
	cacheService := setupTestRedis(t)
	ctx := context.Background()
	key := "test:resource"
	ttl := 10 * time.Second
	lockID, acquired, err := cacheService.AcquireLock(ctx, key, ttl)
	assert.NoError(t, err)
	assert.True(t, acquired)
	assert.NotEmpty(t, lockID)

	// Del
	err = cacheService.ReleaseLock(ctx, key, lockID)
	assert.NoError(t, err, "Lock should be released successfully")

	lockKey := fmt.Sprintf("locked:%s", key)
	exists, err := cacheService.Exists(ctx, lockKey)
	assert.NoError(t, err)
	assert.False(t, exists, "Lock should no longer exist after release")

	// 第二次：應該可以再次獲取
	lockID2, acquired2, err2 := cacheService.AcquireLock(ctx, key, ttl)
	require.NoError(t, err2)
	require.True(t, acquired2)
	require.NotEqual(t, lockID, lockID2, "New lock should have different ID")

	// clear
	cacheService.ReleaseLock(ctx, key, lockID2)

}
