package types

// SystemInfo represents information about the Docker system
type SystemInfo struct {
	// ID is the unique identifier of the daemon
	ID string `json:"id"`

	// Name is the hostname of the system
	Name string `json:"name"`

	// ServerVersion is the version of the Docker daemon
	ServerVersion string `json:"server_version"`

	// OperatingSystem is the host operating system
	OperatingSystem string `json:"operating_system"`

	// Architecture is the host architecture
	Architecture string `json:"architecture"`

	// NCPU is the number of CPUs
	NCPU int `json:"n_cpu"`

	// MemTotal is the total amount of memory in bytes
	MemTotal int64 `json:"mem_total"`

	// DockerRootDir is the root directory of the Docker daemon
	DockerRootDir string `json:"docker_root_dir"`

	// Debug indicates whether debug mode is enabled
	Debug bool `json:"debug"`

	// Images is the number of images
	Images int `json:"images"`

	// Containers is the number of containers
	Containers int `json:"containers"`

	// ContainersRunning is the number of running containers
	ContainersRunning int `json:"containers_running"`

	// ContainersPaused is the number of paused containers
	ContainersPaused int `json:"containers_paused"`

	// ContainersStopped is the number of stopped containers
	ContainersStopped int `json:"containers_stopped"`
}

// DiskUsage represents disk usage information
type DiskUsage struct {
	// LayersSize is the total size of all image layers
	LayersSize int64 `json:"layers_size"`

	// Images is detailed information about images
	Images []ImageDiskUsage `json:"images"`

	// Containers is detailed information about containers
	Containers []ContainerDiskUsage `json:"containers"`

	// Volumes is detailed information about volumes
	Volumes []VolumeDiskUsage `json:"volumes"`
}

// ImageDiskUsage represents disk usage of an image
type ImageDiskUsage struct {
	ID          string            `json:"id"`
	ParentID    string            `json:"parent_id,omitempty"`
	RepoTags    []string          `json:"repo_tags,omitempty"`
	Size        int64             `json:"size"`
	SharedSize  int64             `json:"shared_size"`
	VirtualSize int64             `json:"virtual_size"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// ContainerDiskUsage represents disk usage of a container
type ContainerDiskUsage struct {
	ID         string   `json:"id"`
	Names      []string `json:"names"`
	Image      string   `json:"image"`
	SizeRw     int64    `json:"size_rw"`
	SizeRootFs int64    `json:"size_root_fs"`
}

// VolumeDiskUsage represents disk usage of a volume
type VolumeDiskUsage struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Mountpoint string `json:"mountpoint"`
	Size       int64  `json:"size"`
}
