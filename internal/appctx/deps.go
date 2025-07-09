package appctx

import (
	"database/sql"

	"github.com/docker/docker/client"
	"github.com/redis/go-redis/v9"
)

type Dependencies struct {
	DockerClient *client.Client
	RedisClient  *redis.Client
	MySQLClient  *sql.DB
}
