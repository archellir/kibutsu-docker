import { errorStore } from '../stores/docker';
import type { DockerError } from '../types/docker';

interface ConnectionOptions {
  maxRetries?: number;
  retryDelay?: number;
  queueSize?: number;
}

interface QueuedMessage {
  data: any;
  resolve: (value: any) => void;
  reject: (reason: any) => void;
}

export class WebSocketConnection {
  private ws: WebSocket | null = null;
  private messageQueue: QueuedMessage[] = [];
  private retryCount = 0;
  private retryTimer: NodeJS.Timeout | null = null;
  private isReconnecting = false;

  constructor(
    private url: string,
    private options: ConnectionOptions = {
      maxRetries: 5,
      retryDelay: 5000,
      queueSize: 100
    }
  ) {}

  async connect(): Promise<void> {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url);
        this.setupEventHandlers(resolve, reject);
      } catch (error) {
        this.handleError('CONNECTION_ERROR', error);
        reject(error);
      }
    });
  }

  private setupEventHandlers(resolve: () => void, reject: (error: any) => void): void {
    if (!this.ws) return;

    this.ws.onopen = () => {
      this.retryCount = 0;
      this.isReconnecting = false;
      this.processQueue();
      resolve();
    };

    this.ws.onclose = () => {
      if (!this.isReconnecting) {
        this.handleReconnect();
      }
    };

    this.ws.onerror = (error) => {
      this.handleError('WEBSOCKET_ERROR', error);
      reject(error);
    };
  }

  private handleReconnect(): void {
    if (this.retryCount >= (this.options.maxRetries || 5)) {
      this.handleError('MAX_RETRIES_EXCEEDED', new Error('Maximum retry attempts exceeded'));
      return;
    }

    this.isReconnecting = true;
    this.retryCount++;

    this.retryTimer = setTimeout(() => {
      this.connect().catch(error => {
        this.handleError('RECONNECTION_ERROR', error);
      });
    }, this.options.retryDelay);
  }

  async send(data: any): Promise<void> {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
      return;
    }

    if (this.messageQueue.length >= (this.options.queueSize || 100)) {
      throw new Error('Message queue full');
    }

    return new Promise((resolve, reject) => {
      this.messageQueue.push({ data, resolve, reject });
      
      if (!this.isReconnecting) {
        this.connect().catch(error => {
          this.handleError('SEND_ERROR', error);
          reject(error);
        });
      }
    });
  }

  private processQueue(): void {
    while (this.messageQueue.length > 0 && this.ws?.readyState === WebSocket.OPEN) {
      const message = this.messageQueue.shift();
      if (message) {
        try {
          this.ws.send(JSON.stringify(message.data));
          message.resolve(undefined);
        } catch (error) {
          message.reject(error);
          this.handleError('QUEUE_PROCESSING_ERROR', error);
        }
      }
    }
  }

  private handleError(code: string, error: unknown): void {
    const errorData: DockerError = {
      message: error instanceof Error ? error.message : 'Unknown WebSocket error',
      code,
      timestamp: new Date()
    };
    errorStore.add(errorData);
  }

  subscribe(handler: (data: any) => void): () => void {
    const messageHandler = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data);
        handler(data);
      } catch (error) {
        this.handleError('MESSAGE_PARSE_ERROR', error);
      }
    };

    this.ws?.addEventListener('message', messageHandler);
    return () => this.ws?.removeEventListener('message', messageHandler);
  }

  close(): void {
    if (this.retryTimer) {
      clearTimeout(this.retryTimer);
    }
    this.ws?.close();
    this.ws = null;
    this.messageQueue = [];
    this.isReconnecting = false;
  }
}

// Connection pool
export class ConnectionPool {
  private connections = new Map<string, WebSocketConnection>();

  getConnection(url: string, options?: ConnectionOptions): WebSocketConnection {
    if (!this.connections.has(url)) {
      this.connections.set(url, new WebSocketConnection(url, options));
    }
    return this.connections.get(url)!;
  }

  closeAll(): void {
    this.connections.forEach(connection => connection.close());
    this.connections.clear();
  }
}

// Export singleton instance
export const connectionPool = new ConnectionPool(); 