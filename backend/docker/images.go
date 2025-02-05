package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"

	apitypes "kibutsuapi/api/types"
)

// ImageManager handles Docker image operations
type ImageManager struct {
	client *client.Client
}

// NewImageManager creates a new image manager
func NewImageManager(client *client.Client) *ImageManager {
	return &ImageManager{client: client}
}

// List returns a list of Docker images
func (m *ImageManager) List(ctx context.Context, filterArgs filters.Args) ([]apitypes.ImageInfo, error) {
	images, err := m.client.ImageList(ctx, image.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	result := make([]apitypes.ImageInfo, 0, len(images))
	for _, img := range images {
		result = append(result, apitypes.ImageInfo{
			ID:          img.ID,
			ParentID:    img.ParentID,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Created:     time.Unix(img.Created, 0),
			Size:        img.Size,
			SharedSize:  img.SharedSize,
			VirtualSize: img.Size + img.SharedSize,
			Labels:      img.Labels,
		})
	}
	return result, nil
}

// Pull pulls a Docker image with progress reporting
func (m *ImageManager) Pull(ctx context.Context, ref string, progressCh chan<- apitypes.PullProgress) error {
	reader, err := m.client.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	decoder := json.NewDecoder(reader)
	for {
		var event apitypes.PullProgress
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading pull progress: %w", err)
		}
		select {
		case progressCh <- event:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// Remove removes a Docker image
func (m *ImageManager) Remove(ctx context.Context, id string, force, pruneChildren bool) error {
	_, err := m.client.ImageRemove(ctx, id, image.RemoveOptions{
		Force:         force,
		PruneChildren: pruneChildren,
	})
	if err != nil {
		return fmt.Errorf("failed to remove image: %w", err)
	}
	return nil
}

// Tag creates a new tag for an existing image
func (m *ImageManager) Tag(ctx context.Context, source, target string) error {
	if err := m.client.ImageTag(ctx, source, target); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}
	return nil
}

// GetHistory returns the history of an image
func (m *ImageManager) GetHistory(ctx context.Context, id string) ([]apitypes.ImageHistory, error) {
	history, err := m.client.ImageHistory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get image history: %w", err)
	}

	result := make([]apitypes.ImageHistory, len(history))
	for i, h := range history {
		result[i] = apitypes.ImageHistory{
			ID:        h.ID,
			Created:   time.Unix(h.Created, 0),
			CreatedBy: h.CreatedBy,
			Size:      h.Size,
			Comment:   h.Comment,
			Tags:      h.Tags,
		}
	}
	return result, nil
}

// GetSystemInfo returns Docker system information
func (m *ImageManager) GetSystemInfo(ctx context.Context) (*apitypes.SystemInfo, error) {
	info, err := m.client.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system info: %w", err)
	}

	return &apitypes.SystemInfo{
		ID:                info.ID,
		Name:              info.Name,
		ServerVersion:     info.ServerVersion,
		OperatingSystem:   info.OperatingSystem,
		Architecture:      info.Architecture,
		NCPU:              info.NCPU,
		MemTotal:          info.MemTotal,
		DockerRootDir:     info.DockerRootDir,
		Debug:             info.Debug,
		Images:            info.Images,
		Containers:        info.Containers,
		ContainersRunning: info.ContainersRunning,
		ContainersPaused:  info.ContainersPaused,
		ContainersStopped: info.ContainersStopped,
	}, nil
}

// GetDiskUsage returns disk usage information
func (m *ImageManager) GetDiskUsage(ctx context.Context) (*apitypes.DiskUsage, error) {
	usage, err := m.client.DiskUsage(ctx, types.DiskUsageOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	return &apitypes.DiskUsage{
		LayersSize: usage.LayersSize,
		Images:     convertImageDiskUsage(usage.Images),
		Containers: convertContainerDiskUsage(usage.Containers),
		Volumes:    convertVolumeDiskUsage(usage.Volumes),
	}, nil
}

func convertImageDiskUsage(images []*image.Summary) []apitypes.ImageDiskUsage {
	result := make([]apitypes.ImageDiskUsage, len(images))
	for i, img := range images {
		result[i] = apitypes.ImageDiskUsage{
			ID:          img.ID,
			ParentID:    img.ParentID,
			RepoTags:    img.RepoTags,
			Size:        img.Size,
			SharedSize:  img.SharedSize,
			VirtualSize: img.Size + img.SharedSize,
			Labels:      img.Labels,
		}
	}
	return result
}

func convertContainerDiskUsage(containers []*types.Container) []apitypes.ContainerDiskUsage {
	result := make([]apitypes.ContainerDiskUsage, len(containers))
	for i, c := range containers {
		result[i] = apitypes.ContainerDiskUsage{
			ID:         c.ID,
			Names:      c.Names,
			Image:      c.Image,
			SizeRw:     c.SizeRw,
			SizeRootFs: c.SizeRootFs,
		}
	}
	return result
}

func convertVolumeDiskUsage(volumes []*volume.Volume) []apitypes.VolumeDiskUsage {
	result := make([]apitypes.VolumeDiskUsage, len(volumes))
	for i, v := range volumes {
		result[i] = apitypes.VolumeDiskUsage{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			Size:       v.UsageData.Size,
		}
	}
	return result
}
