package types

import "time"

// ImageInfo represents information about a Docker image
type ImageInfo struct {
	// ID is the unique identifier of the image
	ID string `json:"id"`

	// ParentID is the ID of the parent image
	ParentID string `json:"parent_id,omitempty"`

	// RepoTags are the repository tags associated with this image
	RepoTags []string `json:"repo_tags,omitempty"`

	// RepoDigests are the repository digests associated with this image
	RepoDigests []string `json:"repo_digests,omitempty"`

	// Created is the timestamp when the image was created
	Created time.Time `json:"created"`

	// Size is the total size of the image in bytes
	Size int64 `json:"size"`

	// SharedSize is the size of shared layers in bytes
	SharedSize int64 `json:"shared_size"`

	// VirtualSize is the total size of the image including shared layers
	VirtualSize int64 `json:"virtual_size"`

	// Labels are the metadata labels associated with the image
	Labels map[string]string `json:"labels,omitempty"`
}

// ImageHistory represents a layer in the image history
type ImageHistory struct {
	// ID is the unique identifier of the layer
	ID string `json:"id"`

	// Created is the timestamp when the layer was created
	Created time.Time `json:"created"`

	// CreatedBy is the command that created the layer
	CreatedBy string `json:"created_by"`

	// Size is the size of the layer in bytes
	Size int64 `json:"size"`

	// Comment is an optional comment
	Comment string `json:"comment,omitempty"`

	// Tags are the tags associated with this layer
	Tags []string `json:"tags,omitempty"`
}

// PullProgress represents the progress of an image pull operation
type PullProgress struct {
	// Status is the current status message
	Status string `json:"status"`

	// ID is the layer ID being processed
	ID string `json:"id,omitempty"`

	// Progress is the progress bar data
	Progress string `json:"progress,omitempty"`

	// ProgressDetail contains detailed progress information
	ProgressDetail struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progressDetail"`

	// Error is set if an error occurred
	Error string `json:"error,omitempty"`
}

// ImageError represents an error that occurred during image operations
type ImageError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ImageError) Error() string {
	return e.Message
}

// Common image-related errors
var (
	ErrImageNotFound = &ImageError{
		Code:    "image_not_found",
		Message: "image not found",
	}
	ErrPullFailed = &ImageError{
		Code:    "pull_failed",
		Message: "failed to pull image",
	}
)
