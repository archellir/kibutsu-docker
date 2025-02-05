package types

import (
	"fmt"
	"time"
)

// ContainerResponse represents the main container information
type ContainerResponse struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	Command    string            `json:"command"`
	Status     string            `json:"status"`
	State      string            `json:"state"`
	Created    time.Time         `json:"created"`
	Started    time.Time         `json:"started,omitempty"`
	Finished   time.Time         `json:"finished,omitempty"`
	Ports      []PortMapping    `json:"ports"`
	Networks   []NetworkInfo    `json:"networks"`
	Mounts     []MountInfo      `json:"mounts"`
	Labels     map[string]string `json:"labels"`
	RestartCount int            `json:"restartCount"`
}

// PortMapping represents container port mappings
type PortMapping struct {
	HostIP        string `json:"hostIp"`
	HostPort      uint16 `json:"hostPort"`
	ContainerPort uint16 `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

// NetworkInfo represents container network information
type NetworkInfo struct {
	Name      string   `json:"name"`
	IPAddress string   `json:"ipAddress"`
	Gateway   string   `json:"gateway"`
	Aliases   []string `json:"aliases,omitempty"`
}

// MountInfo represents container mount information
type MountInfo struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
}

// ContainerStats represents container resource usage statistics
type ContainerStats struct {
	CPU struct {
		UsagePercent float64 `json:"usagePercent"`
		SystemUsage  uint64  `json:"systemUsage"`
		UserUsage    uint64  `json:"userUsage"`
	} `json:"cpu"`
	Memory struct {
		Usage    uint64  `json:"usage"`
		Limit    uint64  `json:"limit"`
		Percent  float64 `json:"percent"`
		RSS      uint64  `json:"rss"`
		Cache    uint64  `json:"cache"`
	} `json:"memory"`
	Network struct {
		RxBytes   uint64 `json:"rxBytes"`
		TxBytes   uint64 `json:"txBytes"`
		RxPackets uint64 `json:"rxPackets"`
		TxPackets uint64 `json:"txPackets"`
	} `json:"network"`
	BlockIO struct {
		Read  uint64 `json:"read"`
		Write uint64 `json:"write"`
	} `json:"blockIO"`
	PIDs     int       `json:"pids"`
	ReadTime time.Time `json:"readTime"`
}

// ContainerLogs represents container log output
type ContainerLogs struct {
	Stdout     []LogEntry `json:"stdout"`
	Stderr     []LogEntry `json:"stderr"`
	StartTime  time.Time  `json:"startTime"`
	EndTime    time.Time  `json:"endTime"`
	TotalLines int        `json:"totalLines"`
}

// LogEntry represents a single container log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Stream    string    `json:"stream"` // "stdout" or "stderr"
	Message   string    `json:"message"`
}

// Container operation errors
type ContainerError struct {
	ID      string `json:"id"`
	Op      string `json:"operation"`
	Message string `json:"message"`
}

func (e *ContainerError) Error() string {
	return fmt.Sprintf("container %s: %s failed: %s", e.ID, e.Op, e.Message)
}

// Common container operation errors
var (
	ErrContainerNotFound = &ContainerError{Op: "find", Message: "container not found"}
	ErrContainerAlreadyRunning = &ContainerError{Op: "start", Message: "container already running"}
	ErrContainerNotRunning = &ContainerError{Op: "stop", Message: "container not running"}
	ErrContainerAccessDenied = &ContainerError{Op: "access", Message: "access denied"}
) 