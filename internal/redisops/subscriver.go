package redisops

import (
	"context"
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
)

func SubscribeExpiredKeys(ctx context.Context, rdb *redis.Client, handler func(containerID string)) {
	pubsub := rdb.PSubscribe(ctx, "__keyevent@0__:expired")
	ch := pubsub.Channel()

	log.Println("Subscribed to Redis expiration events...")

	go func() {
		for msg := range ch {
			key := msg.Payload
			if strings.HasPrefix(key, "container:") {
				id := strings.TrimPrefix(key, "container:")
				handler(id)
			}
		}
	}()
}
