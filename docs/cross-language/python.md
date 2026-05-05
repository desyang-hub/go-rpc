# Python Client

This guide covers integrating Python clients with your go-rpc services.

## Installation

```bash
pip install go-rpc-client
```

Or add to requirements.txt:

```
go-rpc-client>=0.2.0
grpcio>=1.58.0
grpcio-tools>=1.58.0
```

## Quick Start

```python
from go_rpc_client import Client, Config

# Create a client with configuration
config = Config(
    address="localhost:50051",
    service_name="my-service",
    timeout=30,
)

client = Client(config)

# Connect to server
await client.connect()

# Make a request
result = await client.call("HelloService.Hello", HelloRequest(name="World"))
print(result.message)

# Disconnect
await client.close()
```

## Configuration

```python
config = Config(
    address="localhost:50051",     # Server address
    port=50051,
    timeout=30,                    # Request timeout (seconds)
    reconnect=True,                # Auto-reconnect on failure
    retry_count=3,                 # Number of retry attempts
    retry_delay=1,                 # Retry delay (seconds)
    tls_enabled=False,             # Enable TLS
    tls_options={                  # TLS configuration
        "cert_file": "/path/to/cert.pem",
    },
    discovery={                    # Service discovery
        "backend": "consul",
        "address": "localhost:8500",
    },
    load_balancer="round_robin",   # Load balancing strategy
    circuit_breaker={              # Circuit breaker settings
        "enabled": True,
        "failure_threshold": 5,
        "recovery_timeout": 30,
    },
)
```

## Generated Code Usage

```python
# Import generated client code
from generated.client import HelloServiceClient
from generated.grpc_generated import HelloRequest, HelloResponse

client = HelloServiceClient("localhost:50051")

async def main():
    response = await client.hello(HelloRequest(name="World"))
    print(response.message)

import asyncio
asyncio.run(main())
```

## Error Handling

```python
from go_rpc_client import RpcError

try:
    result = await client.call("HelloService.Hello", HelloRequest(name="World"))
except RpcError as e:
    print(f"RPC error: {e.code()} - {e.details()}")
except TimeoutError:
    print("Request timed out")
```

## Authentication

```python
# Token-based authentication
config = Config(
    address="localhost:50051",
    auth_type="token",
    auth_token="your-service-token",
)
```

## Observability

```python
import opentelemetry
from go_rpc_client import Config

tracer = opentelemetry.trace.get_tracer("my-client")

config = Config(
    address="localhost:50051",
    tracing_enabled=True,
    tracer_channel=tracer,
)
```

## Next Steps

- [TypeScript Client](typescript.md) — TypeScript integration
- [Code Generation](code-generation.md) — Generating client code
