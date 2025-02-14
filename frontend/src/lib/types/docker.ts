export interface Container {
  Id: string;
  Names: string[];
  Image: string;
  ImageID: string;
  Command: string;
  Created: number;
  State: string;
  Status: string;
  Ports: Port[];
}

export interface Port {
  IP: string;
  PrivatePort: number;
  PublicPort: number;
  Type: string;
}

export interface Image {
  Id: string;
  ParentId: string;
  RepoTags: string[];
  Created: number;
  Size: number;
  VirtualSize: number;
}

export interface ComposeProject {
  name: string;
  path: string;
  status: 'running' | 'stopped' | 'partial';
  services: string[];
}

export interface SystemInfo {
  containers: number;
  images: number;
  memoryUsage: number;
  cpuUsage: number;
  version: string;
  NCPU: number;
  MemTotal: number;
  operatingSystem: string;
  architecture: string;
}

export interface DockerError {
  message: string;
  code: string;
  timestamp: Date;
}

export interface DiskUsage {
  layersSize: number;
  images: ImageDiskUsage[];
  containers: ContainerDiskUsage[];
  volumes: VolumeDiskUsage[];
}

export interface ImageDiskUsage {
  id: string;
  size: number;
  sharedSize: number;
  uniqueSize: number;
}

export interface ContainerDiskUsage {
  id: string;
  size: number;
}

export interface VolumeDiskUsage {
  name: string;
  size: number;
} 