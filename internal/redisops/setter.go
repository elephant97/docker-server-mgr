package redisops

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func RegisterContainerTTL(ctx context.Context, rdb *redis.Client, containerID string, ttl time.Duration) error {
	key := fmt.Sprintf("container:%s", containerID)
	return rdb.Set(ctx, key, "active", ttl).Err()
}
