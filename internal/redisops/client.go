package redisops

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"docker-server-mgr/config"
	clog "docker-server-mgr/utils/log" //custom log
)

func NewRedisClient(redisConfig *config.DBConfig) *redis.Client {

	addr := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		clog.Fatal("❌ Redis 연결 실패: %v", err)
	}

	clog.Debug("✅ Redis 연결 성공")
	return rdb
}
