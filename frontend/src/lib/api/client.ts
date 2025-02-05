import type { Container, Image, ComposeProject, SystemInfo } from '../types/docker';

const API_BASE = '/api';
const WS_BASE = 'ws://localhost:3000';

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

export class DockerClient {
  private token?: string;
  private ws?: WebSocket;

  constructor(baseUrl = API_BASE, wsUrl = WS_BASE) {
    this.setupWebSocket();
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
  private setupWebSocket(): void {
    this.ws = new WebSocket(`${WS_BASE}/docker`);
    
    this.ws.onclose = () => {
      setTimeout(() => this.setupWebSocket(), 5000);
    };
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

  // System operations
  async getSystemInfo(): Promise<SystemInfo> {
    return this.fetch('/system').then(r => r.json());
  }

  // Base fetch method with error handling
  private async fetch(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<Response> {
    const url = `${API_BASE}${endpoint}`;
    const headers = new Headers(options.headers);
    
    if (this.token) {
      headers.set('Authorization', `Bearer ${this.token}`);
    }

    const response = await fetch(url, {
      ...options,
      headers
    });

    if (!response.ok) {
      throw new ApiError(
        response.status,
        await response.text()
      );
    }

    return response;
  }

  // Stream handling
  async* streamLogs(containerId: string): AsyncGenerator<string> {
    const response = await this.fetch(`/containers/${containerId}/logs`);
    const reader = response.body!.getReader();
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
}

// Export singleton instance
export const dockerClient = new DockerClient(); 