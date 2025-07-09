package main

import (
	"context"
	"log"

	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/dockerops"
	"docker-server-mgr/internal/monitor"
	"docker-server-mgr/internal/mysqlops"
	"docker-server-mgr/internal/redisops"
	"docker-server-mgr/internal/server"
)

func main() {
	ctx := context.Background()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dockerClient, err := dockerops.NewDockerClient()
	if err != nil {
		log.Fatalf("Docker client error: %v", err)
	}

	mysqlClient, err := mysqlops.MysqlConnection()
	if err != nil {
		log.Fatalf("MySQL client error: %v", err)
	}

	redisClient := redisops.NewRedisClient("128.10.30.70:6379")

	deps := &appctx.Dependencies{
		DockerClient: dockerClient,
		RedisClient:  redisClient,
		MySQLClient:  mysqlClient,
	}

	// HTTP server 리스너 시작 (listen + 요청마다 thread 생성)
	go server.StartHTTPServer(ctx, deps)

	// Redis docker container TTL 만료 감시 thread
	go redisops.SubscribeExpiredKeys(ctx, redisClient, func(containerID string) {
		log.Printf("Expired container detached: %s\n", containerID)
		go dockerops.RemoveContainer(ctx, deps, containerID)
	})

	// 3. docker life sycle 감시 thread
	go monitor.CheckDockerStatus(ctx, deps)

	// main block
	select {}
}
