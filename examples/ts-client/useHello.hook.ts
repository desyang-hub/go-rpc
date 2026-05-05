/**
 * React Hook for HelloService
 * Generated as example - demonstrates integration with React
 *
 * Usage:
 *   const { hello, helloStream, batchHello, helloStreamStream } = useHelloService('localhost:50051');
 *   const response = await hello('World');
 */

import { useCallback, useRef, useEffect, useState } from 'react';
import { HelloServiceClient, HelloResponseType, HelloStreamResponseType } from './client';

interface UseHelloServiceOptions {
  retryCount?: number;
  retryDelayMs?: number;
  onError?: (error: Error) => void;
}

interface HelloServiceHookResult {
  hello: (name: string, greetType?: string) => Promise<HelloResponseType>;
  helloStream: (name: string) => Promise<HelloStreamResponseType[]>;
  batchHello: (names: string[]) => Promise<HelloResponseType>;
  helloStreamStream: (names: string[]) => Promise<HelloStreamResponseType[]>;
  setStatus: (status: 'connecting' | 'connected' | 'disconnected') => void;
}

export function useHelloService(
  host: string = 'localhost:50051',
  options: UseHelloServiceOptions = {}
): HelloServiceHookResult {
  const { retryCount = 3, retryDelayMs = 1000, onError } = options;
  const clientRef = useRef<HelloServiceClient | null>(null);

  const [status, setStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected');

  // Initialize client on mount
  useEffect(() => {
    setStatus('connecting');
    clientRef.current = new HelloServiceClient(host);
    setStatus('connected');

    // Cleanup on unmount
    return () => {
      clientRef.current?.close();
      setStatus('disconnected');
    };
  }, [host]);

  const withRetry = useCallback(async <T>(
    fn: () => Promise<T>,
    retries = retryCount
  ): Promise<T> => {
    let lastError: Error | null = null;
    for (let i = 0; i <= retries; i++) {
      try {
        return await fn();
      } catch (err) {
        lastError = err as Error;
        if (i < retries) {
          await new Promise(resolve => setTimeout(resolve, retryDelayMs * (2 ** i)));
        }
      }
    }
    onError?.(lastError!);
    throw lastError!;
  }, [retryCount, retryDelayMs, onError]);

  // RPC methods with automatic retry
  const hello = useCallback((name: string, greetType: string = 'greeting') => {
    return withRetry(() => clientRef.current!.hello(name, greetType));
  }, [withRetry]);

  const helloStream = useCallback((name: string) => {
    return withRetry(() => clientRef.current!.helloStream(name));
  }, [withRetry]);

  const batchHello = useCallback((names: string[]) => {
    return withRetry(() => clientRef.current!.batchHello(names));
  }, [withRetry]);

  const helloStreamStream = useCallback((names: string[]) => {
    return withRetry(() => clientRef.current!.helloStreamStream(names));
  }, [withRetry]);

  return { hello, helloStream, batchHello, helloStreamStream, setStatus };
}

// Vue composable equivalent
export function useHelloServiceVue(
  host: string = 'localhost:50051'
): Record<string, any> {
  const client = {
    hello: (name: string) => clientRef.value?.hello(name),
    helloStream: (name: string) => clientRef.value?.helloStream(name),
    batchHello: (names: string[]) => clientRef.value?.batchHello(names),
    helloStreamStream: (names: string[]) => clientRef.value?.helloStreamStream(names),
  };
  return client;
}
export const clientRef = { current: null as HelloServiceClient | null };
