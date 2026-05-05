// React Hook for HelloService - Simple state management without external dependencies
//
// Usage in React:
//
//   import { useHello } from './useHello.hook';
//   import { RpcClient } from './client';
//
//   const client = new RpcClient('http://localhost:8081');
//
//   function HelloComponent() {
//     const { data, loading, error, execute } = useHello(client);
//
//     const handleHello = () => {
//       execute({ name: 'User', greet_type: 'greeting' });
//     };
//
//     if (loading) return <div>Loading...</div>;
//     if (error) return <div>Error: {error.message}</div>;
//
//     return (
//       <div>
//         <p>{data?.message}</p>
//         <button onClick={handleHello}>Say Hello</button>
//       </div>
//     );
//   }

import { useState, useCallback } from 'react';

import { RpcClient, HelloRequest, HelloResponse, RpcError } from './client';

// Hook return type
export interface UseHelloResult {
  data: HelloResponse | null;
  loading: boolean;
  error: RpcError | null;
  execute: (request: HelloRequest) => Promise<void>;
  reset: () => void;
}

// Hook implementation
export function useHello(
  client: RpcClient,
  defaultValue: HelloResponse | null = null
): UseHelloResult {
  const [data, setData] = useState<HelloResponse | null>(defaultValue);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<RpcError | null>(null);

  const execute = useCallback(async (request: HelloRequest): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      const response = await client.hello(request);
      setData(response);
    } catch (e) {
      if (e instanceof RpcError) {
        setError(e);
      } else {
        setError(new RpcError(13, (e as Error).message));
      }
    } finally {
      setLoading(false);
    }
  }, [client]);

  const reset = useCallback(() => {
    setData(defaultValue);
    setError(null);
    setLoading(false);
  }, [defaultValue]);

  return { data, loading, error, execute, reset };
}

// ========== React Hook for PingService ==========

export interface UsePingResult {
  data: { timestamp: string; latency_us: number } | null;
  loading: boolean;
  error: RpcError | null;
  execute: () => Promise<void>;
}

export function usePing(
  client: RpcClient,
  defaultValue: null = null
): UsePingResult {
  const [data, setData] = useState(defaultValue);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<RpcError | null>(null);

  const execute = useCallback(async (): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      const response = await client.ping();
      setData({
        timestamp: response.timestamp,
        latency_us: response.latency_us,
      });
    } catch (e) {
      if (e instanceof RpcError) {
        setError(e);
      } else {
        setError(new RpcError(13, (e as Error).message));
      }
    } finally {
      setLoading(false);
    }
  }, [client]);

  return { data, loading, error, execute };
}

// ========== Generic React Hook for Any RPC Method ==========

interface UseGenericRpcResult<TRequest, TResponse> {
  data: TResponse | null;
  loading: boolean;
  error: RpcError | null;
  execute: (request: TRequest) => Promise<void>;
  reset: () => void;
}

export function useGenericRpc<TRequest, TResponse>(
  client: RpcClient,
  call: (req: TRequest) => Promise<TResponse>,
): UseGenericRpcResult<TRequest, TResponse> {
  const [data, setData] = useState<TResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<RpcError | null>(null);

  const execute = useCallback(async (request: TRequest): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      const response = await call(request);
      setData(response);
    } catch (e) {
      if (e instanceof RpcError) {
        setError(e);
      } else {
        setError(new RpcError(13, (e as Error).message));
      }
    } finally {
      setLoading(false);
    }
  }, [client, call]);

  const reset = useCallback(() => {
    setData(null);
    setError(null);
    setLoading(false);
  }, []);

  return { data, loading, error, execute, reset };
}
