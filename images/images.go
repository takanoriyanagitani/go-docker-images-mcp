package images

import (
	"context"

	"github.com/docker/docker/api/types/image"
)

// ImageSummary provides a lean summary of a Docker image.
type ImageSummary struct {
	ID       string   `json:"id"`
	RepoTags []string `json:"repoTags"`
	Created  int64    `json:"created"`
	Size     int64    `json:"size"`
}

// GetMockImages returns a hardcoded list of ImageSummary structs for testing.
func GetMockImages() ([]ImageSummary, error) {
	return []ImageSummary{
		{
			ID:       "sha256:f1b3f28a5259",
			RepoTags: []string{"ubuntu:latest", "ubuntu:22.04"},
			Created:  1678886400,
			Size:     72957747,
		},
		{
			ID:       "sha256:c3c3c3c3c3c3",
			RepoTags: []string{"golang:1.21-alpine"},
			Created:  1678880000,
			Size:     389934592,
		},
		{
			ID:       "sha256:a1a1a1a1a1a1",
			RepoTags: []string{"alpine:latest"},
			Created:  1678870000,
			Size:     5592324,
		},
	}, nil
}

// DockerClient defines the interface for Docker client operations needed by this package.
type DockerClient interface {
	ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error)
}

// GetImages retrieves a list of Docker images using the Docker client.
func GetImages(ctx context.Context, cli DockerClient) ([]ImageSummary, error) {
	imagesList, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}

	var summaries []ImageSummary
	for _, img := range imagesList {
		summaries = append(summaries, ImageSummary{
			ID:       img.ID,
			RepoTags: img.RepoTags,
			Created:  img.Created,
			Size:     img.Size,
		})
	}

	return summaries, nil
}
