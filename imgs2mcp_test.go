package imgs2mcp

import (
	"reflect"
	"testing"

	"github.com/takanoriyanagitani/go-docker-images-mcp/images"
)

var testImages = []images.ImageSummary{
	{ID: "img1", RepoTags: []string{"ubuntu:22.04"}, Created: 1700000000, Size: 50 * 1024 * 1024},  // 50MB
	{ID: "img2", RepoTags: []string{"alpine:latest"}, Created: 1600000000, Size: 5 * 1024 * 1024},  // 5MB
	{ID: "img3", RepoTags: []string{"ubuntu:20.04"}, Created: 1650000000, Size: 150 * 1024 * 1024}, // 150MB
	{ID: "img4", RepoTags: []string{"golang:1.21"}, Created: 1720000000, Size: 300 * 1024 * 1024},  // 300MB
}

func TestApplyFilters(t *testing.T) {
	testCases := []struct {
		name     string
		input    ImagesInput
		expected []images.ImageSummary
	}{
		{
			name:     "No filters",
			input:    ImagesInput{},
			expected: testImages,
		},
		{
			name:     "Filter by name prefix (ubuntu)",
			input:    ImagesInput{NameStartsWith: "ubuntu"},
			expected: []images.ImageSummary{testImages[0], testImages[2]},
		},
		{
			name:     "Filter by name prefix (none)",
			input:    ImagesInput{NameStartsWith: "nginx"},
			expected: []images.ImageSummary{},
		},
		{
			name:     "Filter by time created since",
			input:    ImagesInput{CreatedSinceUnix: 1690000000},
			expected: []images.ImageSummary{testImages[0], testImages[3]},
		},
		{
			name:     "Filter by min size (100MB)",
			input:    ImagesInput{MinSizeMB: 100},
			expected: []images.ImageSummary{testImages[2], testImages[3]},
		},
		{
			name:     "Filter by max size (100MB)",
			input:    ImagesInput{MaxSizeMB: 100},
			expected: []images.ImageSummary{testImages[0], testImages[1]},
		},
		{
			name:     "Filter by limit (2)",
			input:    ImagesInput{Limit: 2},
			expected: []images.ImageSummary{testImages[0], testImages[1]},
		},
		{
			name:     "Filter by limit only (1)",
			input:    ImagesInput{Limit: 1},
			expected: []images.ImageSummary{testImages[0]},
		},
		{
			name: "Combined filters (ubuntu, >100MB)",
			input: ImagesInput{
				NameStartsWith: "ubuntu",
				MinSizeMB:      100,
			},
			expected: []images.ImageSummary{testImages[2]},
		},
		{
			name: "Combined filters (all filters, one match)",
			input: ImagesInput{
				NameStartsWith:   "golang",
				CreatedSinceUnix: 1710000000,
				MinSizeMB:        200,
				MaxSizeMB:        400,
			},
			expected: []images.ImageSummary{testImages[3]},
		},
		{
			name:     "Limit more than results",
			input:    ImagesInput{Limit: 10},
			expected: testImages,
		},
		{
			name:     "No matches",
			input:    ImagesInput{NameStartsWith: "fedora"},
			expected: []images.ImageSummary{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := applyFilters(testImages, tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected:\n%#v\nGot:\n%#v", tc.expected, result)
			}
		})
	}
}
