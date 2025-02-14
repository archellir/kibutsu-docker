import type { Container, Image, ComposeProject, SystemInfo, DiskUsage } from '../types/docker';

const API_BASE = '/api';
const getWsUrl = () => {
  if (typeof window !== 'undefined') {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    return `${protocol}//${host}${API_BASE}`;
  }
  return null;
};

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

interface ContainerStats {
  cpuPercentage: number;
  memoryUsage: number;
  networkRx: number;
  networkTx: number;
}

export class DockerClient {
  private token?: string;
  private ws: WebSocket | null = null;
  private baseUrl: string;
  private wsUrl: string | null;

  constructor() {
    this.baseUrl = API_BASE;
    this.wsUrl = null; // Initialize without WebSocket
    
    // Only setup WebSocket on client side
    if (typeof window !== 'undefined') {
      // Defer WebSocket setup to avoid SSR issues
      setTimeout(() => {
        this.wsUrl = getWsUrl();
        this.setupWebSocket();
      }, 0);
    }
  }

  // Authentication methods
  async login(username: string, password: string): Promise<void> {
    const response = await this.fetch('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password })
    });
    this.token = await response.text();
  }

  // WebSocket handling
  private setupWebSocket() {
    if (!this.wsUrl || typeof window === 'undefined') {
      return;
    }

    try {
      this.ws = new WebSocket(`${this.wsUrl}/docker`);
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
    }
  }

  // Container operations
  async getContainers(): Promise<Container[]> {
    return this.fetch('/containers').then(r => r.json());
  }

  async startContainer(id: string): Promise<void> {
    await this.fetch(`/containers/${id}/start`, { method: 'POST' });
  }

  async stopContainer(id: string): Promise<void> {
    await this.fetch(`/containers/${id}/stop`, { method: 'POST' });
  }

  async restartContainer(id: string): Promise<void> {
    await this.fetch(`/containers/${id}/restart`, { method: 'POST' });
  }

  // Image operations
  async getImages(): Promise<Image[]> {
    return this.fetch('/images').then(r => r.json());
  }

  async pullImage(name: string): Promise<ReadableStream> {
    const response = await this.fetch(`/images/pull?name=${encodeURIComponent(name)}`, {
      method: 'POST'
    });
    return response.body!;
  }

  // Compose operations
  async getComposeProjects(): Promise<ComposeProject[]> {
    return this.fetch('/compose').then(r => r.json());
  }

  async composeUp(project: string): Promise<ReadableStream> {
    const response = await this.fetch(`/compose/${project}/up`, {
      method: 'POST'
    });
    return response.body!;
  }

  async composeDown(project: string): Promise<void> {
    await this.fetch(`/compose/${project}/down`, {
      method: 'POST'
    });
  }

  // System operations
  async getSystemInfo(): Promise<SystemInfo> {
    return this.fetch('/system/info').then(r => r.json());
  }

  async getDiskUsage(): Promise<DiskUsage> {
    return this.fetch('/system/disk').then(r => r.json());
  }

  // Base fetch method with error handling
  private async fetch(path: string, options: RequestInit = {}): Promise<Response> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      throw new ApiError(response.status, await response.text());
    }

    return response;
  }

  // Stream handling
  async* streamLogs(id: string): AsyncGenerator<string> {
    const response = await this.fetch(`/containers/${id}/logs`);
    const reader = response.body?.getReader();
    if (!reader) throw new Error('Failed to get log stream');

    const decoder = new TextDecoder();
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      yield decoder.decode(value);
    }
  }

  // WebSocket event subscription
  onDockerEvent(callback: (event: any) => void): () => void {
    const handler = (event: MessageEvent) => {
      callback(JSON.parse(event.data));
    };
    
    this.ws?.addEventListener('message', handler);
    return () => this.ws?.removeEventListener('message', handler);
  }

  // Cleanup
  destroy(): void {
    this.ws?.close();
  }

  async getContainerStats(id: string): Promise<ContainerStats> {
    const response = await this.fetch(`/containers/${id}/stats?stream=false`);
    const stats = await response.json();
    
    // Calculate CPU percentage
    const cpuDelta = stats.cpu_stats.cpu_usage.total_usage - stats.precpu_stats.cpu_usage.total_usage;
    const systemDelta = stats.cpu_stats.system_cpu_usage - stats.precpu_stats.system_cpu_usage;
    const cpuPercentage = (cpuDelta / systemDelta) * 100;

    // Calculate memory usage percentage
    const memoryUsage = (stats.memory_stats.usage / stats.memory_stats.limit) * 100;

    // Get network I/O
    const networkStats = Object.values(stats.networks)[0] as { rx_bytes: number; tx_bytes: number };

    return {
      cpuPercentage,
      memoryUsage,
      networkRx: networkStats.rx_bytes,
      networkTx: networkStats.tx_bytes
    };
  }

  getWebSocketUrl(path: string = '/docker'): string | null {
    return this.wsUrl ? `${this.wsUrl}${path}` : null;
  }

  async removeImage(id: string): Promise<void> {
    await this.fetch(`/images/${id}`, { method: 'DELETE' });
  }
}

// Export singleton instance
export const dockerClient = new DockerClient(); 