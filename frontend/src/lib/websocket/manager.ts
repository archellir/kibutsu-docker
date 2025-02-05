import type { DockerError } from '../types/docker';
import { errorStore } from '../stores/docker';

interface WebSocketOptions {
  url: string;
  reconnectDelay?: number;
  maxRetries?: number;
}

type MessageHandler = (data: any) => void;
type MessageQueue = { type: string; data: any }[];

export class WebSocketManager {
  private connections: Map<string, WebSocket> = new Map();
  private messageHandlers: Map<string, Set<MessageHandler>> = new Map();
  private messageQueues: Map<string, MessageQueue> = new Map();
  private reconnectTimers: Map<string, NodeJS.Timeout> = new Map();
  private retryCount: Map<string, number> = new Map();

  constructor(private baseUrl: string = 'ws://localhost:3000') {}

  connect(path: string, options: WebSocketOptions = { url: '' }): WebSocket {
    const url = options.url || `${this.baseUrl}${path}`;
    const existingConnection = this.connections.get(url);
    
    if (existingConnection?.readyState === WebSocket.OPEN) {
      return existingConnection;
    }

    const ws = new WebSocket(url);
    this.connections.set(url, ws);
    this.messageQueues.set(url, []);
    this.retryCount.set(url, 0);

    ws.onmessage = (event) => this.handleMessage(url, event);
    ws.onclose = () => this.handleClose(url, options);
    ws.onerror = () => this.handleError(url);
    ws.onopen = () => this.handleOpen(url);

    return ws;
  }

  subscribe(path: string, handler: MessageHandler): () => void {
    const url = `${this.baseUrl}${path}`;
    if (!this.messageHandlers.has(url)) {
      this.messageHandlers.set(url, new Set());
    }
    this.messageHandlers.get(url)!.add(handler);

    return () => {
      this.messageHandlers.get(url)?.delete(handler);
    };
  }

  send(path: string, data: any): void {
    const url = `${this.baseUrl}${path}`;
    const ws = this.connections.get(url);

    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(data));
    } else {
      this.messageQueues.get(url)?.push({ type: 'send', data });
    }
  }

  private handleMessage(url: string, event: MessageEvent): void {
    try {
      const data = JSON.parse(event.data);
      this.messageHandlers.get(url)?.forEach(handler => handler(data));
    } catch (error) {
      this.reportError('MESSAGE_PARSE_ERROR', error);
    }
  }

  private handleClose(url: string, options: WebSocketOptions): void {
    const retryCount = this.retryCount.get(url) || 0;
    if (retryCount < (options.maxRetries || Infinity)) {
      const timer = setTimeout(() => {
        this.retryCount.set(url, retryCount + 1);
        this.connect(url, options);
      }, options.reconnectDelay || 5000);
      this.reconnectTimers.set(url, timer);
    }
  }

  private handleError(url: string): void {
    this.reportError('WEBSOCKET_ERROR', new Error(`WebSocket error for ${url}`));
  }

  private handleOpen(url: string): void {
    this.retryCount.set(url, 0);
    const queue = this.messageQueues.get(url) || [];
    while (queue.length > 0) {
      const message = queue.shift();
      if (message) this.send(url, message.data);
    }
  }

  private reportError(code: string, error: unknown): void {
    errorStore.add({
      message: error instanceof Error ? error.message : 'Unknown WebSocket error',
      code,
      timestamp: new Date()
    });
  }

  // Specialized connections
  connectToContainerStats(containerId: string): WebSocket {
    return this.connect(`/containers/${containerId}/stats`, {
      url: `${this.baseUrl}/containers/${containerId}/stats`
    });
  }

  connectToContainerLogs(containerId: string): WebSocket {
    return this.connect(`/containers/${containerId}/logs`, {
      url: `${this.baseUrl}/containers/${containerId}/logs`
    });
  }

  connectToTerminal(): WebSocket {
    return this.connect('/terminal', {
      url: `${this.baseUrl}/terminal`,
      reconnectDelay: 3000
    });
  }

  connectToComposeLogs(projectName: string): WebSocket {
    return this.connect(`/compose/${projectName}/logs`, {
      url: `${this.baseUrl}/compose/${projectName}/logs`
    });
  }

  connectToPullProgress(image: string): WebSocket {
    return this.connect(`/images/pull/${image}`, {
      url: `${this.baseUrl}/images/pull/${image}`
    });
  }

  disconnect(path: string): void {
    const url = `${this.baseUrl}${path}`;
    const ws = this.connections.get(url);
    if (ws) {
      ws.close();
      this.connections.delete(url);
      this.messageQueues.delete(url);
      this.messageHandlers.delete(url);
      clearTimeout(this.reconnectTimers.get(url));
      this.reconnectTimers.delete(url);
      this.retryCount.delete(url);
    }
  }

  disconnectAll(): void {
    for (const url of this.connections.keys()) {
      this.disconnect(url);
    }
  }
}

// Export singleton instance
export const wsManager = new WebSocketManager(); 