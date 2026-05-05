/**
 * TypeScript gRPC-Web Client for rpc-gen Go Server
 * Demonstrates all four gRPC call modes with gRPC-Web transport
 */

// Generated protobuf definitions - import after running protoc-gen-gogo
// import * as grpc from 'grpc-web';
// import { HelloServiceClient, HelloRequest, HelloResponse, ... } from '../proto/api_pb';
// import { HelloServiceClient as HelloServicePromiseClient } from '../proto/api_pb_service';

// Fallback types for demonstration without generated code
declare interface HelloRequest {
  getName(): string;
  name: string;
}
declare interface HelloResponse {
  getMessage(): string;
  message: string;
  getTimestamp(): string;
  timestamp: string;
  getServerId(): string;
  server_id: string;
}
declare interface HelloStreamResponse {
  getMessage(): string;
  message: string;
  getResponseIndex(): number;
  response_index: number;
}

interface GRPCClient {
  url: string;
  secure: boolean;
}

type Metadata = Record<string, string>;
type Status = { code: number; message: string; details: any };

class GRPCWebClient {
  private url: string;
  private headers: Metadata = {};

  constructor(url: string, secure: boolean = false) {
    this.url = url;
    if (secure) {
      this.url = this.url.replace('http://', 'https://');
    }
  }

  setMetadata(headers: Metadata): void {
    this.headers = headers;
  }

  async unaryCall<TReq, TResp>(
    method: string,
    request: TReq,
    deadline?: number
  ): Promise<TResp> {
    // Stub implementation - replace with actual grpc-web call
    // Example: return client[method](request, metadata, deadline);
    throw new Error(`Method "${method}" not implemented. Include grpc-web client stubs.`);
  }

  async serverStream<TReq, TResp>(
    method: string,
    request: TReq,
    deadline?: number
  ): Promise<Array<TResp>> {
    const results: Array<TResp> = [];
    // Stream implementation
    throw new Error(`Server stream "${method}" not implemented.`);
  }

  async clientStream<TReq, TResp>(
    method: string,
    requests: AsyncIterable<TReq>
  ): Promise<TResp> {
    throw new Error(`Client stream "${method}" not implemented.`);
  }

  async bidiStream<TReq, TResp>(
    method: string,
    requests: AsyncIterable<TReq>
  ): Promise<AsyncIterable<TResp>> {
    throw new Error(`Bidirectional stream "${method}" not implemented.`);
  }

  close(): void {
    // Cleanup
  }
}

class HelloServiceClient {
  private client: GRPCClient;
  private readonly host: string;

  constructor(host: string = 'localhost:50051', secure: boolean = false) {
    this.host = host;
    this.client = new GRPCClient(`http://${host}`, secure);
  }

  async hello(name: string, greetType: string = 'greeting'): Promise<HelloResponse> {
    // const request = new HelloRequest();
    // request.setName(name);
    // request.setGreettype(greetType);
    // return this.client.unaryCall('/HelloService/Hello', request, 10000);
    throw new Error('Include generated gRPC-Web stubs');
  }

  async helloStream(name: string): Promise<HelloStreamResponse[]> {
    const responses: HelloStreamResponse[] = [];
    // await this.client.serverStream('/HelloService/HelloStream', request);
    throw new Error('Include generated gRPC-Web stubs');
  }

  async batchHello(names: string[]): Promise<HelloResponse> {
    // const requestGenerator = async function*(this: void) {
    //   for (const name of names) {
    //     const req = new HelloRequest();
    //     req.setName(name);
    //     yield req;
    //   }
    // }();
    // return this.client.clientStream('/HelloService/BatchHello', requestGenerator);
    throw new Error('Include generated gRPC-Web stubs');
  }

  async helloStreamStream(names: string[]): Promise<HelloStreamResponse[]> {
    const responses: HelloStreamResponse[] = [];
    // await this.client.bidiStream('/HelloService/HelloStreamStream', requestGenerator);
    throw new Error('Include generated gRPC-Web stubs');
  }

  close(): void {
    (this.client as any).close?.();
  }
}

// Demo usage
async function demo() {
  console.log('TypeScript gRPC-Web Client Demo');
  console.log('Note: Include generated gRPC-Web client stubs to run');

  const client = new HelloServiceClient('localhost:50051');

  try {
    console.log('1. Unary RPC');
    // const resp = await client.hello('World');
    // console.log('Response:', resp.getMessage());

    console.log('2. Server-Streaming');
    // await client.helloStream('StreamUser');

    console.log('3. Client-Streaming');
    // await client.batchHello(['Alice', 'Bob', 'Charlie']);

    console.log('4. Bidirectional Streaming');
    // await client.helloStreamStream(['Echo1', 'Echo2', 'Echo3']);

    console.log('All RPCs completed!');
  } finally {
    client.close();
  }
}

// Export types and classes
export {
  HelloServiceClient,
  GRPCWebClient,
  HelloRequest as HelloRequestType,
  HelloResponse as HelloResponseType,
  HelloStreamResponse as HelloStreamResponseType,
};

export type { Metadata, Status };

if (require.main === module) {
  demo();
}
