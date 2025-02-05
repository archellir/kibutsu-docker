package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	apitypes "kibutsuapi/api/types"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"gopkg.in/yaml.v3"
)

type ComposeProject struct {
	Name       string
	ConfigPath string
	Config     *apitypes.ComposeConfig
	client     *client.Client
	mu         sync.RWMutex
}

type ProjectStatus struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Services  map[string]ServiceInfo `json:"services"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

type ServiceInfo struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	Replicas     int               `json:"replicas"`
	DesiredState string            `json:"desiredState"`
	Containers   []ContainerInfo   `json:"containers"`
	Labels       map[string]string `json:"labels"`
}

type ContainerInfo struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Status  string    `json:"status"`
	Created time.Time `json:"created"`
}

func NewComposeProject(client *client.Client, name string, config *apitypes.ComposeConfig) (*ComposeProject, error) {
	return &ComposeProject{
		Name:       name,
		ConfigPath: filepath.Join("compose", name, "docker-compose.yml"),
		Config:     config,
		client:     client,
	}, nil
}

func (p *ComposeProject) Up(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Create networks first
	if err := p.createNetworks(ctx); err != nil {
		return fmt.Errorf("failed to create networks: %w", err)
	}

	// Create and start services in dependency order
	services := p.getServiceOrder()
	for _, serviceName := range services {
		if err := p.startService(ctx, serviceName); err != nil {
			return fmt.Errorf("failed to start service %s: %w", serviceName, err)
		}
	}

	return nil
}

func (p *ComposeProject) Down(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop and remove containers in reverse dependency order
	services := p.getServiceOrder()
	for i := len(services) - 1; i >= 0; i-- {
		if err := p.stopService(ctx, services[i]); err != nil {
			return fmt.Errorf("failed to stop service %s: %w", services[i], err)
		}
	}

	// Remove networks
	return p.removeNetworks(ctx)
}

func (p *ComposeProject) Scale(ctx context.Context, service string, replicas int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	svcConfig, exists := p.Config.Services[service]
	if !exists {
		return fmt.Errorf("service %s not found", service)
	}

	// Get current containers for the service
	f := filters.NewArgs()
	f.Add("label", fmt.Sprintf("com.docker.compose.project=%s", p.Name))
	f.Add("label", fmt.Sprintf("com.docker.compose.service=%s", service))

	containers, err := p.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return err
	}

	currentCount := len(containers)

	// Scale up
	if replicas > currentCount {
		for i := currentCount; i < replicas; i++ {
			if err := p.createContainer(ctx, service, svcConfig, i); err != nil {
				return err
			}
		}
	}

	// Scale down
	if replicas < currentCount {
		for i := currentCount - 1; i >= replicas; i-- {
			if err := p.removeContainer(ctx, containers[i].ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *ComposeProject) Status(ctx context.Context) (*ProjectStatus, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	f := filters.NewArgs()
	f.Add("label", fmt.Sprintf("com.docker.compose.project=%s", p.Name))

	containers, err := p.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return nil, err
	}

	services := make(map[string]ServiceInfo)
	for serviceName := range p.Config.Services {
		services[serviceName] = ServiceInfo{
			Name:       serviceName,
			Status:     "not_created",
			Replicas:   0,
			Containers: make([]ContainerInfo, 0),
		}
	}

	for _, c := range containers {
		serviceName := c.Labels["com.docker.compose.service"]
		svc := services[serviceName]
		svc.Replicas++
		svc.Status = c.State
		svc.Containers = append(svc.Containers, ContainerInfo{
			ID:      c.ID,
			Name:    strings.TrimPrefix(c.Names[0], "/"),
			Status:  c.Status,
			Created: time.Unix(c.Created, 0),
		})
		services[serviceName] = svc
	}

	return &ProjectStatus{
		Name:     p.Name,
		Services: services,
		Status:   p.determineProjectStatus(services),
	}, nil
}

func (p *ComposeProject) Logs(ctx context.Context, service string, follow bool) (io.ReadCloser, error) {
	f := filters.NewArgs()
	f.Add("label", fmt.Sprintf("com.docker.compose.project=%s", p.Name))
	if service != "" {
		f.Add("label", fmt.Sprintf("com.docker.compose.service=%s", service))
	}

	containers, err := p.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return nil, err
	}

	readers := make([]io.Reader, 0, len(containers))
	for _, c := range containers {
		logs, err := p.client.ContainerLogs(ctx, c.ID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     follow,
			Timestamps: true,
		})
		if err != nil {
			continue
		}
		readers = append(readers, logs)
	}

	return io.NopCloser(io.MultiReader(readers...)), nil
}

// Helper functions
func loadComposeFile(path string) (*apitypes.ComposeConfig, error) {
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

func (p *ComposeProject) getServiceOrder() []string {
	// Create dependency graph
	graph := make(map[string][]string)
	for name, service := range p.Config.Services {
		graph[name] = service.DependsOn
	}

	// Perform topological sort
	visited := make(map[string]bool)
	temp := make(map[string]bool)
	var order []string

	var visit func(node string) bool
	visit = func(node string) bool {
		if temp[node] {
			return false // Cycle detected
		}
		if visited[node] {
			return true
		}
		temp[node] = true

		for _, dep := range graph[node] {
			if !visit(dep) {
				return false
			}
		}

		temp[node] = false
		visited[node] = true
		order = append(order, node)
		return true
	}

	for name := range p.Config.Services {
		if !visited[name] {
			if !visit(name) {
				// Handle circular dependencies
				log.Printf("Warning: circular dependency detected in service %s", name)
			}
		}
	}

	return order
}

func (p *ComposeProject) createNetworks(ctx context.Context) error {
	for name, config := range p.Config.Networks {
		if config.External {
			continue
		}

		_, err := p.client.NetworkCreate(ctx, fmt.Sprintf("%s_%s", p.Name, name), types.NetworkCreate{
			Driver: "bridge",
			Labels: map[string]string{
				"com.docker.compose.project": p.Name,
				"com.docker.compose.network": name,
			},
		})
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create network %s: %w", name, err)
		}
	}
	return nil
}

func (p *ComposeProject) removeNetworks(ctx context.Context) error {
	f := filters.NewArgs()
	f.Add("label", fmt.Sprintf("com.docker.compose.project=%s", p.Name))

	networks, err := p.client.NetworkList(ctx, types.NetworkListOptions{Filters: f})
	if err != nil {
		return err
	}

	for _, network := range networks {
		if err := p.client.NetworkRemove(ctx, network.ID); err != nil {
			log.Printf("Warning: failed to remove network %s: %v", network.Name, err)
		}
	}
	return nil
}

func (p *ComposeProject) startService(ctx context.Context, service string) error {
	svcConfig := p.Config.Services[service]
	replicas := 1
	if svcConfig.Deploy != nil && svcConfig.Deploy.Replicas > 0 {
		replicas = svcConfig.Deploy.Replicas
	}

	for i := 0; i < replicas; i++ {
		if err := p.createContainer(ctx, service, svcConfig, i); err != nil {
			return err
		}
	}
	return nil
}

func (p *ComposeProject) stopService(ctx context.Context, service string) error {
	f := filters.NewArgs()
	f.Add("label", fmt.Sprintf("com.docker.compose.project=%s", p.Name))
	f.Add("label", fmt.Sprintf("com.docker.compose.service=%s", service))

	containers, err := p.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return err
	}

	timeout := 30
	for _, c := range containers {
		if err := p.client.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
			log.Printf("Warning: failed to stop container %s: %v", c.ID, err)
			continue
		}
		if err := p.client.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true}); err != nil {
			log.Printf("Warning: failed to remove container %s: %v", c.ID, err)
		}
	}
	return nil
}

func (p *ComposeProject) createContainer(ctx context.Context, service string, config apitypes.ServiceSpec, index int) error {
	// Parse port mappings
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for _, portStr := range config.Ports {
		portMapping, err := nat.ParsePortSpec(portStr)
		if err != nil {
			return fmt.Errorf("invalid port mapping %s: %w", portStr, err)
		}
		for _, pm := range portMapping {
			portBindings[pm.Port] = append(portBindings[pm.Port], pm.Binding)
			exposedPorts[pm.Port] = struct{}{}
		}
	}

	// Create container config
	containerConfig := &container.Config{
		Image:        config.Image,
		Cmd:          config.Command,
		Env:          mapToEnvSlice(config.Environment),
		ExposedPorts: exposedPorts,
		Labels: map[string]string{
			"com.docker.compose.project":  p.Name,
			"com.docker.compose.service":  service,
			"com.docker.compose.instance": strconv.Itoa(index),
		},
	}

	// Create host config
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        config.Volumes,
	}

	// Create networking config
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: make(map[string]*network.EndpointSettings),
	}
	for netName := range p.Config.Networks {
		networkConfig.EndpointsConfig[fmt.Sprintf("%s_%s", p.Name, netName)] = &network.EndpointSettings{}
	}

	// Create container
	resp, err := p.client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, fmt.Sprintf("%s_%s_%d", p.Name, service, index))
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := p.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

func (p *ComposeProject) removeContainer(ctx context.Context, containerID string) error {
	// Stop container first
	timeout := 30
	if err := p.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}

	// Remove container
	if err := p.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerID, err)
	}

	return nil
}

func (p *ComposeProject) determineProjectStatus(services map[string]ServiceInfo) string {
	total := len(services)
	if total == 0 {
		return "not_created"
	}

	running := 0
	stopped := 0
	for _, svc := range services {
		switch svc.Status {
		case "running":
			running++
		case "exited", "dead":
			stopped++
		}
	}

	if running == total {
		return "running"
	}
	if stopped == total {
		return "stopped"
	}
	return "partial"
}

// Helper function to convert map[string]string to []string for environment variables
func mapToEnvSlice(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}
