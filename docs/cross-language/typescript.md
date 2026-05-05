# TypeScript Client

This guide covers integrating TypeScript clients with your go-rpc services.

## Installation

```bash
npm install @desyang-hub/go-rpc-client
```

Or add to package.json:

```json
{
  "dependencies": {
    "@desyang-hub/go-rpc-client": "^0.2.0"
  }
}
```

## Quick Start

```typescript
import { Client, Config } from '@desyang-hub/go-rpc-client';

async function main() {
  // Create a client with configuration
  const config: Config = {
    address: 'localhost:50051',
    service_name: 'my-service',
    timeout: 30,
  };

  const client = new Client(config);

  // Connect to server
  await client.connect();

  // Make a unary RPC call
  const result = await client.call('HelloService.Hello', { name: 'World' });
  console.log(result.message);

  // Disconnect
  await client.close();
}

main().catch(console.error);
```

## Configuration

```typescript
const config: Config = {
  address: 'localhost:50051',       // Server address
  port: 50051,
  timeout: 30,                      // Request timeout (seconds)
  reconnect: true,                  // Auto-reconnect on failure
  retry_count: 3,                   // Number of retry attempts
  retry_delay: 1,                   // Retry delay (seconds)
  tls_enabled: false,               // Enable TLS
  auth_type: 'token',               // Authentication type
  auth_token: 'your-service-token', // Authentication token
  load_balancer: 'round_robin',     // Load balancing strategy
  circuit_breaker: {                // Circuit breaker settings
    enabled: true,
    failure_threshold: 5,
    recovery_timeout: 30,
  },
};
```

## Generated Code Usage

```typescript
// Import generated client code
import {HelloServiceClient} from './generated/client';
import {HelloRequest, HelloResponse} from './generated/grpc_generated';

async function main() {
  const client = new HelloServiceClient('localhost:50051');
  const response = await client.hello({ name: 'World' });
  console.log(response.message);
}

main().catch(console.error);
```

## React Hook Integration

```typescript
import { useGrpcCall } from '@desyang-hub/go-rpc-client/react';

function HelloComponent() {
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const grpcCall = useGrpcCall({
    endpoint: 'HelloService.Hello',
    client: myClient,
  });

  const handleHello = async (name: string) => {
    setLoading(true);
    try {
      const response = await grpcCall.call({ name });
      setResult(response);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      {loading && <Spinner />}
      {error && <ErrorMessage error={error} />}
      {result && <p>Message: {result.message}</p>}
      <button onClick={() => handleHello('World')}>Say Hello</button>
    </div>
  );
}
```

## Error Handling

```typescript
import { RpcError } from '@desyang-hub/go-rpc-client';

try {
  const result = await client.call('HelloService.Hello', { name: 'World' });
} catch (err) {
  if (err instanceof RpcError) {
    console.log(err.code, err.details);
  } else if (err instanceof TimeoutError) {
    console.log('Request timed out');
  }
}
```

## Next Steps

- [Code Generation](code-generation.md) — Generating TypeScript client code
- [Kubernetes Deployment](../deployment/kubernetes.md) — Deploying to K8s
