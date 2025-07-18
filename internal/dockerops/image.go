package dockerops

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	shptypes "docker-server-mgr/internal/dockerops/types"
	"docker-server-mgr/internal/mysqlops"

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
		log.Printf("getSavedImageInfo failed: %v", err)
		panic(err)
	}

	containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		log.Printf("Error listing containers: %v", err)
		panic(err)
	}

	usedImageIDs := make(map[string]bool)
	for _, c := range containers {
		usedImageIDs[c.ImageID] = true
	}

	images, err := dockerClient.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		log.Printf("Error get ImageList: %v", err)
		panic(err)
	}

	for _, img := range images {
		if !usedImageIDs[img.ID] {
			fmt.Printf("Deleting unused image: %s\n", img.ID[:20])
			_, err := dockerClient.ImageRemove(ctx, img.ID, types.ImageRemoveOptions{
				Force:         false,
				PruneChildren: false,
			})
			if err != nil || img.RepoTags == nil {
				fmt.Printf("⚠️ Failed to delete %s: %v\n", img.ID[:20], err)
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
		log.Printf("⚠️ Failed get Image Info in DB %v\n", err)
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
	log.Printf("Updating image %s status to %s", imageID, status)

	_, err := mysqlops.ExecQuery(mysqlClient, `
    INSERT INTO images (id, status, name, tag, last_check_time)
    VALUES (?, ?, ?, ?, NOW())
    ON DUPLICATE KEY UPDATE
        status = VALUES(status),
        last_check_time = NOW()`,
		imageID, status, imageName, imageTag)

	if err != nil {
		log.Printf("Error upsert container status in MySQL: %v", err)
	} else {
		log.Printf("Image %s status upsert to %s successfully", imageID, status)
	}
}
