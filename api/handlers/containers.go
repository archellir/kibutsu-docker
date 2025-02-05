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
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

	apitypes "kibutsu/api/types"
)

type ContainerResponse struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Image    string                 `json:"image"`
	Status   string                 `json:"status"`
	Created  time.Time              `json:"created"`
	Ports    []apitypes.PortMapping `json:"ports"`
	Networks []apitypes.NetworkInfo `json:"networks"`
	Mounts   []apitypes.MountInfo   `json:"mounts"`
}

type ContainerHandler struct {
	client *client.Client
}

func NewContainerHandler(client *client.Client) *ContainerHandler {
	return &ContainerHandler{client: client}
}

func (h *ContainerHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	containers, err := h.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list containers: %v", err), http.StatusInternalServerError)
		return
	}

	response := make([]apitypes.ContainerResponse, 0, len(containers))
	for _, c := range containers {
		inspect, err := h.client.ContainerInspect(ctx, c.ID)
		if err != nil {
			continue
		}

		created, err := time.Parse(time.RFC3339Nano, inspect.Created)
		if err != nil {
			created = time.Unix(0, 0)
		}

		response = append(response, apitypes.ContainerResponse{
			ID:       c.ID,
			Name:     strings.TrimPrefix(inspect.Name, "/"),
			Image:    c.Image,
			Status:   c.Status,
			Created:  created,
			Ports:    convertPorts(c.Ports),
			Networks: convertNetworks(inspect.NetworkSettings.Networks),
			Mounts:   convertMounts(inspect.Mounts),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ContainerHandler) GetContainer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/containers/")
	id = strings.Split(id, "/")[0]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	inspect, err := h.client.ContainerInspect(ctx, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Container not found: %v", err), http.StatusNotFound)
		return
	}

	created, err := time.Parse(time.RFC3339Nano, inspect.Created)
	if err != nil {
		created = time.Unix(0, 0)
	}

	response := apitypes.ContainerResponse{
		ID:       inspect.ID,
		Name:     strings.TrimPrefix(inspect.Name, "/"),
		Image:    inspect.Config.Image,
		Status:   inspect.State.Status,
		Created:  created,
		Networks: convertNetworks(inspect.NetworkSettings.Networks),
		Mounts:   convertMounts(inspect.Mounts),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ContainerHandler) StartContainer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/containers/")
	id = strings.Split(id, "/")[0]

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := h.client.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start container: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ContainerHandler) StopContainer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/containers/")
	id = strings.Split(id, "/")[0]

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	timeout := 30 * time.Second
	timeoutSeconds := int(timeout.Seconds())
	if err := h.client.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop container: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ContainerHandler) RestartContainer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/containers/")
	id = strings.Split(id, "/")[0]

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	timeout := 30 * time.Second
	timeoutSeconds := int(timeout.Seconds())
	if err := h.client.ContainerRestart(ctx, id, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to restart container: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ContainerHandler) GetContainerLogs(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/containers/")
	id = strings.Split(id, "/")[0]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "100",
		Timestamps: true,
	}

	logs, err := h.client.ContainerLogs(ctx, id, options)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get logs: %v", err), http.StatusInternalServerError)
		return
	}
	defer logs.Close()

	w.Header().Set("Content-Type", "text/plain")
	io.Copy(w, logs)
}

func (h *ContainerHandler) GetContainerStats(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/containers/")
	id = strings.Split(id, "/")[0]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	stats, err := h.client.ContainerStats(ctx, id, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}
	defer stats.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, stats.Body)
}

// Helper functions to convert Docker SDK types to our API types
func convertPorts(ports []types.Port) []apitypes.PortMapping {
	result := make([]apitypes.PortMapping, len(ports))
	for i, p := range ports {
		result[i] = apitypes.PortMapping{
			HostIP:        p.IP,
			HostPort:      p.PublicPort,
			ContainerPort: p.PrivatePort,
			Protocol:      p.Type,
		}
	}
	return result
}

func convertNetworks(networks map[string]*network.EndpointSettings) []apitypes.NetworkInfo {
	result := make([]apitypes.NetworkInfo, 0, len(networks))
	for name, net := range networks {
		result = append(result, apitypes.NetworkInfo{
			Name:      name,
			IPAddress: net.IPAddress,
			Gateway:   net.Gateway,
			Aliases:   net.Aliases,
		})
	}
	return result
}

func convertMounts(mounts []types.MountPoint) []apitypes.MountInfo {
	result := make([]apitypes.MountInfo, len(mounts))
	for i, m := range mounts {
		result[i] = apitypes.MountInfo{
			Type:        string(m.Type),
			Source:      m.Source,
			Destination: m.Destination,
			Mode:        m.Mode,
			RW:          m.RW,
		}
	}
	return result
}
