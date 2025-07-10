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

	for {
		select {
		case <-ctx.Done():
			log.Println("Redis subscription cancelled")
			return
		case msg, ok := <-ch:
			if !ok {
				log.Println("Redis channel closed")
				return
			}
			key := msg.Payload

			if strings.HasPrefix(key, "container:") {
				id := strings.TrimPrefix(key, "container:")
				if id == "" {
					log.Println("Received invalid container ID from Redis")
					continue
				}
				log.Printf("Container expired: %s\n", id)
				handler(id)
			}
		}
	}
}
