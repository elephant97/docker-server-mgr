package dockerops

import (
	"context"
	"database/sql"
	"strings"
	"time"

	shptypes "docker-server-mgr/internal/dockerops/types"
	"docker-server-mgr/internal/mysqlops"

	clog "docker-server-mgr/utils/log" //custom log

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func WatchImageUsingStatus(
	ctx context.Context,
	dockerClient *client.Client,
	mysqlClient *sql.DB,
) {

	mapDbImage, err := getSavedImageInfo(mysqlClient)
	if mapDbImage == nil {
		clog.Error("getSavedImageInfo failed", "err", err)
		panic(err)
	}

	containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		clog.Error("Error listing containers", "err", err)
		panic(err)
	}

	usedImageIDs := make(map[string]bool)
	for _, c := range containers {
		usedImageIDs[c.ImageID] = true
	}

	images, err := dockerClient.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		clog.Error("Error get ImageList", "err", err)
		panic(err)
	}

	for _, img := range images {
		timeSinceCreated := time.Since(time.Unix(img.Created, 0))
		// 컨테이너에서 사용중이 아니면서 이미지 생성된지 1주일이 지났으면 사용하지 않는 것으로 간주하고 삭제처리
		if !usedImageIDs[img.ID] && timeSinceCreated > 7*24*time.Hour {
			clog.Info("사용하지 않는 이미지 삭제처리", "imageID", img.ID[:20])
			_, err := dockerClient.ImageRemove(ctx, img.ID, types.ImageRemoveOptions{
				Force:         false,
				PruneChildren: false,
			})
			if err != nil {
				clog.Error("사용하지 않는 이미지 삭제에러 발생", "imageID", img.ID[:20], "err", err)
			}

			if mapDbImage[img.ID] == "" || mapDbImage[img.ID] != "deleted" {
				imageStatusUpdate(img.RepoTags, mysqlClient, img.ID, "deleted")
			}
		} else {
			if mapDbImage[img.ID] == "" || mapDbImage[img.ID] != "using" {
				imageStatusUpdate(img.RepoTags, mysqlClient, img.ID, "using")
			}
		}
	}
}

func getSavedImageInfo(
	mysqlClient *sql.DB,
) (map[string]string, error) {
	dbImageList, err := mysqlops.SelectQueryRowsToStructs[shptypes.ImageStatus](mysqlClient,
		"SELECT id, status FROM images")
	if err != nil {
		clog.Error("⚠️ Failed get Image Info in DB", "err", err)
		return nil, err
	}
	imageMap := make(map[string]string)

	for _, img := range dbImageList {
		imageMap[img.ID] = img.Status.String
	}

	return imageMap, nil
}

func imageStatusUpdate(
	repoTags []string,
	mysqlClient *sql.DB,
	imageID string,
	status string,
) {
	for _, tag := range repoTags {
		parts := strings.SplitN(tag, ":", 2)
		if len(parts) == 2 {
			imageName := parts[0]
			imageTag := parts[1]
			upsertImageStatus(mysqlClient, imageID, imageName, imageTag, status)
		} else {
			clog.Warn("repo Tag split count is over or lower then 2", "len", len(parts))
		}
	}
}

func upsertImageStatus(
	mysqlClient *sql.DB,
	imageID string,
	imageName string,
	imageTag string,
	status string,
) {
	clog.Debug("Updating image %s status to %s", imageID, status)

	_, err := mysqlops.ExecQuery(mysqlClient, `
    INSERT INTO images (id, status, name, tag, last_check_time)
    VALUES (?, ?, ?, ?, NOW())
    ON DUPLICATE KEY UPDATE
        status = VALUES(status),
        last_check_time = NOW()`,
		imageID, status, imageName, imageTag)

	if err != nil {
		clog.Error("Error upsert container status in MySQL", "err", err)
	} else {
		clog.Debug("Image status upsert successfully", "imageId", imageID, "status", status)
	}
}
