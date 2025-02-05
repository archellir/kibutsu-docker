package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"

	apitypes "kibutsu/api/types"
	"kibutsu/docker"
)

type ComposeHandler struct {
	client *client.Client
}

func NewComposeHandler(client *client.Client) *ComposeHandler {
	return &ComposeHandler{client: client}
}

func (h *ComposeHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Filter containers with compose label
	f := filters.NewArgs()
	f.Add("label", "com.docker.compose.project")

	containers, err := h.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list compose projects: %v", err), http.StatusInternalServerError)
		return
	}

	// Group containers by project
	projects := make(map[string][]apitypes.ContainerResponse)
	for _, c := range containers {
		projectName := c.Labels["com.docker.compose.project"]
		if projectName == "" {
			continue
		}

		inspect, err := h.client.ContainerInspect(ctx, c.ID)
		if err != nil {
			continue
		}

		created, _ := time.Parse(time.RFC3339Nano, inspect.Created)
		projects[projectName] = append(projects[projectName], apitypes.ContainerResponse{
			ID:      c.ID,
			Name:    strings.TrimPrefix(inspect.Name, "/"),
			Image:   c.Image,
			Status:  c.Status,
			Created: created,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *ComposeHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/compose/projects/")
	name = strings.Split(name, "/")[0]

	config, err := h.loadComposeFile(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load compose file: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (h *ComposeHandler) ProjectUp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/compose/projects/")
	name = strings.Split(name, "/")[0]

	config, err := h.loadComposeFile(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load compose file: %v", err), http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	if err := h.startProject(ctx, name, config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start project: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ComposeHandler) ProjectDown(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/compose/projects/")
	name = strings.Split(name, "/")[0]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	f := filters.NewArgs()
	f.Add("label", fmt.Sprintf("com.docker.compose.project=%s", name))

	containers, err := h.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list project containers: %v", err), http.StatusInternalServerError)
		return
	}

	for _, c := range containers {
		timeout := 30
		if err := h.client.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
			continue
		}
		if err := h.client.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true}); err != nil {
			continue
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ComposeHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/compose/projects/")
	name = strings.Split(name, "/")[0]

	config, err := h.loadComposeFile(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load compose file: %v", err), http.StatusNotFound)
		return
	}

	services := make([]string, 0, len(config.Services))
	for service := range config.Services {
		services = append(services, service)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (h *ComposeHandler) GetProjectLogs(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/compose/projects/")
	name = strings.Split(name, "/")[0]

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	f := filters.NewArgs()
	f.Add("label", fmt.Sprintf("com.docker.compose.project=%s", name))

	containers, err := h.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list project containers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	for _, c := range containers {
		serviceName := c.Labels["com.docker.compose.service"]
		fmt.Fprintf(w, "=== %s ===\n", serviceName)

		options := container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       "100",
			Timestamps: true,
		}

		logs, err := h.client.ContainerLogs(ctx, c.ID, options)
		if err != nil {
			continue
		}
		io.Copy(w, logs)
		logs.Close()
		fmt.Fprintln(w)
	}
}

func (h *ComposeHandler) ScaleService(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/compose/projects/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	projectName := parts[0]
	serviceName := parts[2]

	var scaleReq struct {
		Replicas int `json:"replicas"`
	}
	if err := json.NewDecoder(r.Body).Decode(&scaleReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	if err := h.scaleService(ctx, projectName, serviceName, scaleReq.Replicas); err != nil {
		http.Error(w, fmt.Sprintf("Failed to scale service: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ComposeHandler) loadComposeFile(project string) (*apitypes.ComposeConfig, error) {
	path := filepath.Join("compose", project, "docker-compose.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config apitypes.ComposeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (h *ComposeHandler) startProject(ctx context.Context, project string, config *apitypes.ComposeConfig) error {
	composeProject, err := docker.NewComposeProject(h.client, project, config)
	if err != nil {
		return fmt.Errorf("failed to create compose project: %w", err)
	}

	if err := composeProject.Up(ctx); err != nil {
		return fmt.Errorf("failed to start project: %w", err)
	}

	return nil
}

func (h *ComposeHandler) scaleService(ctx context.Context, project, service string, replicas int) error {
	config, err := h.loadComposeFile(project)
	if err != nil {
		return fmt.Errorf("failed to load compose file: %w", err)
	}

	composeProject, err := docker.NewComposeProject(h.client, project, config)
	if err != nil {
		return fmt.Errorf("failed to create compose project: %w", err)
	}

	if err := composeProject.Scale(ctx, service, replicas); err != nil {
		return fmt.Errorf("failed to scale service: %w", err)
	}

	return nil
}
