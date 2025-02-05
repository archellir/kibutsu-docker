package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"golang.org/x/net/websocket"

	apitypes "kibutsu/api/types"
)

type ImageHandler struct {
	client *client.Client
}

func NewImageHandler(client *client.Client) *ImageHandler {
	return &ImageHandler{client: client}
}

func (h *ImageHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Parse filter query parameters
	filterArgs := filters.NewArgs()
	if dangling := r.URL.Query().Get("dangling"); dangling != "" {
		filterArgs.Add("dangling", dangling)
	}
	if reference := r.URL.Query().Get("reference"); reference != "" {
		filterArgs.Add("reference", reference)
	}

	images, err := h.client.ImageList(ctx, image.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list images: %v", err), http.StatusInternalServerError)
		return
	}

	response := make([]apitypes.ImageInfo, 0, len(images))
	for _, img := range images {
		response = append(response, apitypes.ImageInfo{
			ID:          img.ID,
			ParentID:    img.ParentID,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Created:     time.Unix(img.Created, 0),
			Size:        img.Size,
			SharedSize:  img.SharedSize,
			VirtualSize: img.VirtualSize,
			Labels:      img.Labels,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/images/")
	id = strings.Split(id, "/")[0]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	inspect, _, err := h.client.ImageInspectWithRaw(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Image not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inspect)
}

func (h *ImageHandler) RemoveImage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/images/")
	id = strings.Split(id, "/")[0]

	force := r.URL.Query().Get("force") == "true"
	prune := r.URL.Query().Get("prune") == "true"

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	_, err := h.client.ImageRemove(ctx, id, image.RemoveOptions{
		Force:         force,
		PruneChildren: prune,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove image: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ImageHandler) PullImage(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		var pullReq struct {
			Image string `json:"image"`
			Tag   string `json:"tag"`
		}
		if err := json.NewDecoder(r.Body).Decode(&pullReq); err != nil {
			websocket.JSON.Send(ws, map[string]string{"error": "Invalid request body"})
			return
		}

		ctx := r.Context()
		ref := pullReq.Image
		if pullReq.Tag != "" {
			ref = fmt.Sprintf("%s:%s", pullReq.Image, pullReq.Tag)
		}

		reader, err := h.client.ImagePull(ctx, ref, image.PullOptions{})
		if err != nil {
			websocket.JSON.Send(ws, map[string]string{"error": fmt.Sprintf("Failed to pull image: %v", err)})
			return
		}
		defer reader.Close()

		decoder := json.NewDecoder(reader)
		for {
			var event apitypes.PullProgress
			if err := decoder.Decode(&event); err != nil {
				if err != io.EOF {
					websocket.JSON.Send(ws, map[string]string{"error": fmt.Sprintf("Error reading pull progress: %v", err)})
				}
				return
			}
			websocket.JSON.Send(ws, event)
		}
	})

	upgrader.ServeHTTP(w, r)
}

func (h *ImageHandler) GetImageHistory(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/images/")
	id = strings.Split(id, "/")[0]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	history, err := h.client.ImageHistory(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get image history: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *ImageHandler) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	info, err := h.client.Info(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get system info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (h *ImageHandler) GetSystemVersion(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	version, err := h.client.ServerVersion(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get system version: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(version)
}

func (h *ImageHandler) GetDiskUsage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	usage, err := h.client.DiskUsage(ctx, types.DiskUsageOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get disk usage: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}
