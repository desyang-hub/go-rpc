// TypeScript gRPC Client - HTTP/REST wrapper for cross-language RPC calls
//
// This client is designed to work with gRPC-Gateway which automatically
// generates REST/JSON endpoints from proto definitions.
//
// Usage:
//
//   import { RpcClient } from '@go-rpc/typescript-client';
//
//   const client = new RpcClient('http://localhost:8081');
//   const response = await client.hello({ name: 'World' });
//   console.log(response.message);
//
// For React/Vue projects, use the composable hooks:
//   import { useHello } from './useHello.hook';
//
// React Example:
//
//   function MyComponent() {
//     const { data, loading, error, call } = useHello(client);
//
//     const handleClick = () => {
//       call({ name: 'User' });
//     };
//
//     return <button onClick={handleClick}>Say Hello</button>;
//   }

import https from 'https';
import url from 'url';

// ==================== Request/Response Types ====================

export interface HelloRequest {
  name: string;
  greet_type?: string;
}

export interface HelloResponse {
  message: string;
  timestamp: string;
  server_id: string;
}

export interface HelloStreamRequest {
  name: string;
  index: number;
}

export interface HelloStreamResponse {
  message: string;
  response_index: number;
}

export interface EchoRequest {
  payload: string;
  data?: string; // base64 encoded
  echo_headers?: boolean;
}

export interface EchoResponse {
  payload: string;
  data?: string;
  headers?: Record<string, string>;
}

export interface PingRequest {
  payload?: string;
}

export interface PingResponse {
  payload: string;
  timestamp: string;
  latency_us: number;
}

// ==================== HTTP Client Configuration ====================

export interface ClientConfig {
  baseUrl?: string;
  timeout?: number;
  headers?: Record<string, string>;
  tls?: {
    cert?: string;
    key?: string;
    ca?: string;
  };
}

const DEFAULT_BASE_URL = 'http://localhost:8081';
const DEFAULT_TIMEOUT = 30000;

// ==================== Core HTTP Client ====================

export class RpcClient {
  private baseUrl: string;
  private timeout: number;
  private headers: Record<string, string>;

  constructor(config?: ClientConfig) {
    this.baseUrl = (config?.baseUrl || DEFAULT_BASE_URL).replace(/\/+$/, '');
    this.timeout = config?.timeout || DEFAULT_TIMEOUT;
    this.headers = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      ...config?.headers,
    };
  }

  // Generic HTTP request method
  private async request<T>(method: string, path: string, body?: any): Promise<T> {
    const fullUrl = `${this.baseUrl}${path}`;
    const reqOptions: https.RequestOptions = {
      hostname: new url.URL(fullUrl).hostname,
      port: new url.URL(fullUrl).port || (fullUrl.startsWith('https://') ? 443 : 80),
      path: new url.URL(fullUrl).pathname + new url.URL(fullUrl).search,
      method: method,
      headers: this.headers,
      timeout: this.timeout,
    };

    return new Promise((resolve, reject) => {
      const client = fullUrl.startsWith('https://') ? https : require('http');
      const req = client.request(reqOptions, (res) => {
        let data = '';

        res.on('data', (chunk: string) => {
          data += chunk;
        });

        res.on('end', () => {
          if (res.statusCode >= 200 && res.statusCode < 300) {
            try {
              resolve(JSON.parse(data) as T);
            } catch (e) {
              reject(new Error(`Failed to parse response: ${e}`));
            }
          } else {
            reject(new Error(`HTTP ${res.statusCode}: ${data}`));
          }
        });
      });

      req.on('error', (e) => {
        reject(e);
      });

      req.on('timeout', () => {
        req.destroy();
        reject(new Error('Request timeout'));
      });

      if (body !== undefined) {
        req.write(JSON.stringify(body));
      }

      req.end();
    });
  }

  // Bytes are encoded/decoded as base64
  private encodeBytes(data: Buffer): string {
    return data.toString('base64');
  }

  // ==================== HelloService Methods ====================

  async hello(req: HelloRequest): Promise<HelloResponse> {
    return this.request<HelloResponse>('POST', '/api/v1/hello', {
      name: req.name,
      greet_type: req.greet_type,
    });
  }

  async ping(req?: PingRequest): Promise<PingResponse> {
    return this.request<PingResponse>('GET', '/api/v1/ping', undefined);
  }

  // ==================== EchoService Methods ====================

  async echo(req: EchoRequest): Promise<EchoResponse> {
    const payload = {
      payload: req.payload,
      data: req.data ? this.encodeBytes(Buffer.from(req.data, 'base64')) : undefined,
      echo_headers: req.echo_headers,
    };
    return this.request<EchoResponse>('POST', '/api/v1/echo', payload);
  }

  // ==================== Utility Methods ====================

  async healthCheck(): Promise<{status: string}> {
    try {
      const response = await this.request<any>('GET', '/api/v1/health');
      return { status: response.status || 'healthy' };
    } catch (error) {
      return { status: 'unhealthy' };
    }
  }

  isConnected(): boolean {
    try {
      new url.URL(this.baseUrl);
      return true;
    } catch {
      return false;
    }
  }
}

// ==================== Error Handling ====================

export class RpcError extends Error {
  readonly code: number;
  readonly message: string;
  readonly details?: any;

  constructor(code: number, message: string, details?: any) {
    super(message);
    this.name = 'RpcError';
    this.code = code;
    this.message = message;
    this.details = details;
  }

  static fromHttpError(status: number, body?: any): RpcError {
    const codes: {[key: number]: number} = {
      400: 3,  // INVALID_ARGUMENT
      401: 16, // UNAUTHENTICATED
      403: 7,  // PERMISSION_DENIED
      404: 5,  // NOT_FOUND
      409: 6,  // ALREADY_EXISTS
      429: 8,  // RESOURCE_EXHAUSTED
      500: 13, // INTERNAL
      502: 14, // UNAVAILABLE
      503: 14, // UNAVAILABLE
      504: 4,  // DEADLINE_EXCEEDED
    };

    return new RpcError(
      codes[status] || 13,
      body?.message || `HTTP ${status}`,
      body
    );
  }
}

// ==================== Client Factory ====================

export function createClient(config?: ClientConfig): RpcClient {
  return new RpcClient(config);
}
