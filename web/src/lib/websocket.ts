import { WebSocketMessage, DatabaseSchema } from '@/types';

export type WebSocketEventListener = (message: WebSocketMessage) => void;

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error';

interface CorrelatedWebSocketMessage extends WebSocketMessage {
  request_id?: string;
}

interface PendingRequest {
  resolve: (value: any) => void;
  reject: (error: any) => void;
  timeout: NodeJS.Timeout;
}

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000; // Start with 1 second
  private listeners: Set<WebSocketEventListener> = new Set();
  private isManualClose = false;
  private connectionStatus: ConnectionStatus = 'disconnected';
  private pendingRequests: Map<string, PendingRequest> = new Map();

  constructor() {
    // Use NEXT_PUBLIC_API_URL to connect to the backend WebSocket server
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    const wsProtocol = apiUrl.startsWith('https') ? 'wss:' : 'ws:';
    this.url = `${wsProtocol}${apiUrl.substring(apiUrl.indexOf('//'))}/ws`;
  }

  connect(): Promise<void> {
    this.connectionStatus = 'connecting';

    return new Promise((resolve, reject) => {
      try {
        console.log(`[WebSocket] Attempting to connect to ${this.url}`);
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log('[WebSocket] ✓ Connected successfully');
          this.connectionStatus = 'connected';
          this.reconnectAttempts = 0;
          this.reconnectDelay = 1000;
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: CorrelatedWebSocketMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket disconnected:', event.code, event.reason);
          this.connectionStatus = 'disconnected';

          // Reject all pending requests
          this.pendingRequests.forEach((req) => {
            clearTimeout(req.timeout);
            req.reject(new Error('WebSocket disconnected'));
          });
          this.pendingRequests.clear();

          if (!this.isManualClose && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
          }
        };

        this.ws.onerror = (error) => {
          console.error('[WebSocket] ✗ Connection error. URL:', this.url, 'Error:', error);
          this.connectionStatus = 'error';
          reject(new Error(`WebSocket connection failed to ${this.url}`));
        };
      } catch (error) {
        this.connectionStatus = 'error';
        reject(error);
      }
    });
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }

    this.reconnectAttempts++;
    console.log(`Scheduling reconnect attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${this.reconnectDelay}ms`);

    this.reconnectTimer = setTimeout(() => {
      this.connect().catch((error) => {
        console.error('Reconnect failed:', error);
        // Exponential backoff
        this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);
      });
    }, this.reconnectDelay);
  }

  disconnect() {
    this.isManualClose = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  send(message: Partial<WebSocketMessage>) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected. Message not sent:', message);
    }
  }

  // Request-response pattern with correlation
  private sendRequest<T>(message: Partial<WebSocketMessage>, timeout: number = 5000): Promise<T> {
    return new Promise((resolve, reject) => {
      if (!this.isConnected()) {
        reject(new Error('WebSocket is not connected'));
        return;
      }

      const requestId = `${message.type}_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
      const correlatedMessage = { ...message, request_id: requestId };

      // Set up timeout
      const timeoutHandle = setTimeout(() => {
        this.pendingRequests.delete(requestId);
        reject(new Error('WebSocket request timeout'));
      }, timeout);

      this.pendingRequests.set(requestId, { resolve, reject, timeout: timeoutHandle });

      this.send(correlatedMessage);
    });
  }

  // Schema-specific methods
  async requestSchema(dataSourceId: string): Promise<DatabaseSchema> {
    return this.sendRequest<DatabaseSchema>(
      {
        type: 'get_schema',
        payload: { data_source_id: dataSourceId },
      },
      2000 // 2 second timeout - fail fast, use REST fallback
    );
  }

  subscribeToSchema(dataSourceId: string) {
    this.send({
      type: 'subscribe_schema',
      payload: { data_source_id: dataSourceId },
    });
  }

  // Message handler with correlation support
  private handleMessage(message: CorrelatedWebSocketMessage) {
    // Handle correlated responses (request-response pattern)
    if (message.request_id) {
      const pending = this.pendingRequests.get(message.request_id);
      if (pending) {
        clearTimeout(pending.timeout);
        this.pendingRequests.delete(message.request_id);

        if (message.type === 'error') {
          pending.reject(new Error(message.payload?.error || 'Unknown error'));
        } else {
          pending.resolve(message.payload);
        }
        return; // Don't notify listeners for correlated responses
      }
    }

    // Notify all listeners for broadcast messages
    this.notifyListeners(message);
  }

  // Event listener management
  addListener(listener: WebSocketEventListener) {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  }

  private notifyListeners(message: WebSocketMessage) {
    this.listeners.forEach((listener) => {
      try {
        listener(message);
      } catch (error) {
        console.error('Error in WebSocket listener:', error);
      }
    });
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  getConnectionStatus(): ConnectionStatus {
    return this.connectionStatus;
  }
}

// Lazy-loaded singleton to avoid SSR issues
let wsServiceInstance: WebSocketService | null = null;

export const wsService = new Proxy({} as WebSocketService, {
  get(_target, prop) {
    if (typeof window === 'undefined') {
      // During SSR, return a no-op implementation
      return () => {
        console.warn('WebSocketService: Cannot access during SSR');
      };
    }

    if (!wsServiceInstance) {
      wsServiceInstance = new WebSocketService();
    }

    const value = (wsServiceInstance as any)[prop];
    if (typeof value === 'function') {
      return value.bind(wsServiceInstance);
    }
    return value;
  },
});

