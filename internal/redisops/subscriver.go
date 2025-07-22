package redisops

import (
	"context"
	"strings"

	clog "docker-server-mgr/utils/log" //custom log

	"github.com/redis/go-redis/v9"
)

func SubscribeExpiredKeys(ctx context.Context, rdb *redis.Client, handler func(containerID string)) {
	pubsub := rdb.PSubscribe(ctx, "__keyevent@0__:expired")
	ch := pubsub.Channel()

	clog.Debug("Subscribed to Redis expiration events...")

	for {
		select {
		case <-ctx.Done():
			clog.Error("Redis subscription cancelled")
			return
		case msg, ok := <-ch:
			if !ok {
				clog.Error("Redis channel closed")
				return
			}
			key := msg.Payload

			if strings.HasPrefix(key, "container:") {
				id := strings.TrimPrefix(key, "container:")
				if id == "" {
					clog.Warn("Received invalid container ID from Redis")
					continue
				}
				clog.Info("Container expired", "containerID", id)
				handler(id)
			}
		}
	}
}
