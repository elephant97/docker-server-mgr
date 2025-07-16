package redisops

import (
	"context"
	"fmt"
	"log"
	"time"

	"docker-server-mgr/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(redisConfig *config.DBConfig) *redis.Client {

	addr := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Redis 연결 실패: %v", err)
	}

	log.Println("✅ Redis 연결 성공")
	return rdb
}
