package imgs2mcp

import (
	"context"
	"net/http"
	"strings"

	"github.com/docker/docker/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/takanoriyanagitani/go-docker-images-mcp/images"
)

// ImagesInput defines the parameters for the "images" tool.
type ImagesInput struct {
	// NameStartsWith filters images where at least one tag has this prefix.
	NameStartsWith string `json:"nameStartsWith,omitempty"`
	// CreatedSinceUnix filters images created after this Unix timestamp.
	CreatedSinceUnix int64 `json:"createdSinceUnix,omitempty"`
	// MinSizeMB filters images smaller than this size in megabytes.
	MinSizeMB int64 `json:"minSizeMb,omitempty"`
	// MaxSizeMB filters images larger than this size in megabytes.
	MaxSizeMB int64 `json:"maxSizeMb,omitempty"`
	// Limit restricts the number of images returned.
	Limit int `json:"limit,omitempty"`
}

// ImagesOutput defines the output of the "images" tool.
type ImagesOutput struct {
	// Result is the list of Docker image summaries that match the filter criteria.
	Result []images.ImageSummary `json:"result"`
}

// applyFilters is a pure function that filters a slice of images based on the input criteria.
func applyFilters(imgs []images.ImageSummary, input ImagesInput) []images.ImageSummary {
	if input.NameStartsWith == "" && input.CreatedSinceUnix == 0 && input.MinSizeMB == 0 && input.MaxSizeMB == 0 {
		if input.Limit <= 0 {
			return imgs
		}
	}

	filtered := make([]images.ImageSummary, 0)
	for _, img := range imgs {
		if input.NameStartsWith != "" {
			nameMatch := false
			for _, tag := range img.RepoTags {
				if strings.HasPrefix(tag, input.NameStartsWith) {
					nameMatch = true
					break
				}
			}
			if !nameMatch {
				continue
			}
		}

		if input.CreatedSinceUnix > 0 {
			if img.Created < input.CreatedSinceUnix {
				continue
			}
		}

		if input.MinSizeMB > 0 {
			if img.Size < (input.MinSizeMB * 1024 * 1024) {
				continue
			}
		}

		if input.MaxSizeMB > 0 {
			if img.Size > (input.MaxSizeMB * 1024 * 1024) {
				continue
			}
		}

		filtered = append(filtered, img)
	}

	if input.Limit > 0 && len(filtered) > input.Limit {
		filtered = filtered[:input.Limit]
	}

	return filtered
}

// NewServer creates a new MCP server and returns it as an http.Handler and the docker client.
func NewServer(dockerHost string) (http.Handler, *client.Client, error) {
	var opts []client.Opt
	if dockerHost != "" {
		opts = append(opts, client.WithHost("unix://"+dockerHost))
	} else {
		opts = append(opts, client.FromEnv)
	}
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, nil, err
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "docker-images",
		Version: "v1.0.0",
	}, nil)

	imagesTool := func(ctx context.Context, req *mcp.CallToolRequest, input ImagesInput) (
		*mcp.CallToolResult,
		ImagesOutput,
		error,
	) {
		imgs, err := images.GetImages(ctx, cli)
		if err != nil {
			return nil, ImagesOutput{}, err
		}

		filtered := applyFilters(imgs, input)

		return nil, ImagesOutput{Result: filtered}, nil
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "images",
		Description: "List Docker images",
	}, imagesTool)

	handler := mcp.NewStreamableHTTPHandler(
		func(req *http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{Stateless: true},
	)

	return handler, cli, nil
}
