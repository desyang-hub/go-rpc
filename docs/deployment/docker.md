# Docker Deployment

This guide covers deploying go-rpc services with Docker.

## Docker Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Docker Host                               │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                     Docker Network                           │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │ │
│  │  │  Go Server   │  │  Go Server   │  │  Go Server   │       │ │
│  │  │   Instance 1  │  │   Instance 2  │  │   Instance 3  │       │ │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │ │
│  │         │                 │                  │               │ │
│  │         └─────────────────┼──────────────────┘               │ │
│  │                           │                                  │ │
│  │                  ┌────────┴────────┐                         │ │
│  │                  │  Load Balancer  │                         │ │
│  │                  │   (Traefik/     │                         │ │
│  │                  │    nginx)       │                         │ │
│  │                  └─────────────────┘                         │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                    │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                   External Services                          │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │ │
│  │  │    Consul    │  │    etcd      │  │   Otel       │       │ │
│  │  │  (Registry)  │  │  (Registry)  │  │  (Tracing)   │       │ │
│  │  └──────────────┘  └──────────────┘  └──────────────┘       │ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Docker Compose Setup

### docker-compose.yml

```yaml
version: '3.8'

services:
  consul:
    image: consul:1.16
    command: agent -dev -bind=0.0.0.0 -client=0.0.0.0
    ports:
      - "8500:8500"  # UI
      - "8600:8600/udp"  # DNS
    volumes:
      - consul-data:/consul/data
    networks:
      - rpc-network

  go-server:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "50051:50051"  # gRPC port
      - "9090:9090"    # Metrics port
    environment:
      - SERVICE_NAME=my-service
      - DISCOVERY_BACKEND=consul
      - CONSUL_ADDRESS=consul:8500
    depends_on:
      - consul
    volumes:
      - ./config:/app/config
    networks:
      - rpc-network

volumes:
  consul-data:

networks:
  rpc-network:
    driver: bridge
```

### docker-compose.override.yml

```yaml
version: '3.8'

services:
  go-server:
    volumes:
      - .:/app
    command: go run ./cmd/rpc-server
```

## Multi-Stage Dockerfile

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/rpc-server ./cmd/rpc-server

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/rpc-server /app/rpc-server
COPY configs /app/configs

EXPOSE 50051
EXPOSE 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:9090/health || exit 1

CMD ["/app/rpc-server", "--config=/app/configs"])
```

## Build and Run

```bash
# Build and start
docker compose up --build

# Run in detached mode
docker compose up -d

# Scale replicas
docker compose up -d --scale go-server=3

# Stop all services
docker compose down
```

## Docker Best Practices

1. **Multi-stage Builds**: Reduce final image size
2. **Non-root User**: Run as non-root for security
3. **Health Checks**: Add HEALTHCHECK to Dockerfile
4. **Resource Limits**: Set memory/CPU limits in docker-compose

## Next Steps

- [Kubernetes Deployment](kubernetes.md) — Production-grade deploy
