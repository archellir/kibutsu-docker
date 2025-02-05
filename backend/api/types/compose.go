package types

import "time"

// ComposeProject represents a Docker Compose project
type ComposeProject struct {
	Name       string            `json:"name"`
	Status     string            `json:"status"`
	Services   []ComposeService  `json:"services"`
	Created    time.Time         `json:"created"`
	LastUpdate time.Time         `json:"lastUpdate"`
	Labels     map[string]string `json:"labels"`
}

// ComposeService represents a service within a Docker Compose project
type ComposeService struct {
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	Status       string            `json:"status"`
	Replicas     int               `json:"replicas"`
	Ports        []PortMapping     `json:"ports"`
	Networks     []NetworkInfo     `json:"networks"`
	Volumes      []MountInfo       `json:"volumes"`
	Environment  map[string]string `json:"environment"`
	Dependencies []string          `json:"dependencies"`
}

// ComposeConfig represents the configuration for a Docker Compose project
type ComposeConfig struct {
	Version  string                 `json:"version"`
	Services map[string]ServiceSpec `json:"services"`
	Networks map[string]NetworkSpec `json:"networks,omitempty"`
	Volumes  map[string]VolumeSpec  `json:"volumes,omitempty"`
}

// ServiceSpec defines the configuration for a service
type ServiceSpec struct {
	Image       string            `json:"image"`
	Command     []string          `json:"command,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Ports       []string          `json:"ports,omitempty"`
	Volumes     []string          `json:"volumes,omitempty"`
	DependsOn   []string          `json:"depends_on,omitempty"`
	Deploy      *DeploySpec       `json:"deploy,omitempty"`
}

// DeploySpec defines deployment configuration for a service
type DeploySpec struct {
	Replicas int `json:"replicas,omitempty"`
}

// NetworkSpec defines network configuration
type NetworkSpec struct {
	External bool `json:"external,omitempty"`
}

// VolumeSpec defines volume configuration
type VolumeSpec struct {
	External bool `json:"external,omitempty"`
}

// ComposeError represents an error that occurred during compose operations
type ComposeError struct {
	Service   string `json:"service,omitempty"`
	Operation string `json:"operation"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Time      string `json:"time"`
}

// ComposeOperation represents the status of a compose operation
type ComposeOperation struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Project   string    `json:"project"`
	Service   string    `json:"service,omitempty"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime,omitempty"`
	Error     string    `json:"error,omitempty"`
}
