package main

import (
	"context"

	"docker-server-mgr/config"
	"docker-server-mgr/internal/appctx"
	"docker-server-mgr/internal/dockerops"
	"docker-server-mgr/internal/monitor"
	"docker-server-mgr/internal/mysqlops"
	"docker-server-mgr/internal/redisops"
	"docker-server-mgr/internal/server"
	"docker-server-mgr/utils"
	clog "docker-server-mgr/utils/log" //custom log
)

func main() {
	ctx := context.Background()

	clog.LogSet()

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		clog.Fatal("설정 파일 로딩 실패: %v", err)
	}

	dockerClient, err := dockerops.NewDockerClient()
	if err != nil {
		clog.Fatal("Docker client error: %v", err)
	}

	mysqlClient, err := mysqlops.MysqlConnection(&cfg.MySQL)
	if err != nil {
		clog.Fatal("MySQL client error: %v", err)
	}

	redisClient := redisops.NewRedisClient(&cfg.Redis)

	deps := &appctx.Dependencies{
		DockerClient: dockerClient,
		RedisClient:  redisClient,
		MySQLClient:  mysqlClient,
	}

	// HTTP server 리스너 시작 (listen + 요청마다 thread 생성)
	utils.SafeGoRoutineCtx(ctx, func() {
		server.StartHTTPServer(ctx, deps)
	})

	// Redis docker container TTL 만료 감시 thread
	utils.SafeGoRoutineCtx(ctx, func() {
		redisops.SubscribeExpiredKeys(ctx, redisClient, func(containerID string) {
			clog.Info("Expired container detached: %s\n", containerID)
			dockerops.RemoveContainer(ctx, deps, containerID)
		})
	})

	// 3. docker life sycle 감시 thread
	utils.SafeGoRoutineCtx(ctx, func() {
		monitor.CheckDockerStatus(ctx, deps)
	})

	// 4. docker 사용중이지 않은 image 관리 thread
	utils.SafeGoRoutineCtx(ctx, func() {
		monitor.CheckImageStatus(ctx, deps)
	})

	// main block
	select {}
}
