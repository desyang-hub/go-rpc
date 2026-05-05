# RPC Examples

This directory contains example applications demonstrating various client-server interactions.

## Table of Contents

- [Go Server](#go-server)
- [Python Client](#python-client)
- [TypeScript Client](#typescript-client)

## Go Server

Production-ready gRPC server with full feature set.

### Features
- Four gRPC call modes (unary, server-stream, client-stream, bidirectional-stream)
- Prometheus metrics at `/metrics`
- Health check endpoint at `/health`
- Graceful shutdown
- Structured logging

### Running

**Docker Compose:**
```bash
docker-compose up -d
```

**Directly:**
```bash
cd go-server
go run main.go
```

**Endpoints:**
- gRPC: `localhost:50051`
- Metrics: `http://localhost:9090/metrics`
- Health: `http://localhost:8081/health`
- Health check: `curl localhost:8081/health`

## Python Client

gRPC client supporting all four call modes with automatic retry logic.

### Features
- Retry with exponential backoff
- All four RPC modes (unary, server-stream, client-stream, bidirectional-stream)
- Type hints for better IDE support

### Running

**Docker Compose:**
```bash
cd python-client
docker-compose up -d
```

**Directly:**
```bash
cd python-client
pip install -r requirements.txt
python3 client.py
```

## TypeScript Client

Type-safe gRPC-Web client with React Hook integration.

### Features
- Type-safe client methods
- React Hook (`useHelloService`) for easy integration
- Vue composable support
- Error handling

### Running

**Docker Compose:**
```bash
cd ts-client
docker-compose up -d
```

**Directly:**
```bash
cd ts-client
npm install
npm start
```

## Generating Protobuf Stubs

```bash
# Go
protoc --go_out=. --go-grpc_out=. api/*.proto

# Python
protoc --python_out=. --grpc_python_out=. api/*.proto

# TypeScript
protoc --js_out=import_mode=copy,index_package:./proto \
       --grpc-web_out=import_mode=copy,index_package:./proto \
       --proto_path=../api api/*.proto
```
