# Kubernetes Deployment

This guide covers deploying go-rpc services on Kubernetes.

## Overview

go-rpc services are designed for cloud-native deployment with full Kubernetes compatibility.

## Core Manifests

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rpc-server
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: rpc-server
  template:
    metadata:
      labels:
        app: rpc-server
        version: v1
    spec:
      containers:
        - name: rpc-server
          image: ghcr.io/desyang/go-rpc:v1.0.0
          ports:
            - containerPort: 50051  # gRPC
            - containerPort: 9090  # metrics
          env:
            - name: SERVICE_NAME
              value: rpc-server
            - name: DISCOVERY_BACKEND
              value: consul
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
          readinessProbe:
            httpGet:
              path: /health
              port: 9090
          livenessProbe:
            httpGet:
              path: /health
              port: 9090
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: rpc-server
  namespace: default
spec:
  type: ClusterIP
  selector:
    app: rpc-server
  ports:
    - name: grpc
      port: 50051
      targetPort: 50051
      protocol: TCP
    - name: metrics
      port: 9090
      targetPort: 9090
      protocol: TCP
```

### HPA

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: rpc-server
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: rpc-server
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

### PodDisruptionBudget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: rpc-server
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: rpc-server
```

### ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: rpc-server
spec:
  selector:
    matchLabels:
      app: rpc-server
  endpoints:
    - port: metrics
      path: /metrics
      interval: 15s
```

## Deployment Commands

```bash
kubectl apply -f manifests/
kubectl scale deployment rpc-server --replicas=5
kubectl rollout restart deployment rpc-server
kubectl rollout status deployment rpc-server
```

## Best Practices

1. Use PodDisruptionBudgets to ensure availability during maintenance
2. Configure resource limits and requests
3. Enable horizontal pod autoscaling based on CPU/memory metrics
4. Use Liveness/Readiness probes

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVICE_NAME` | rpc-server | Service name for discovery |
| `DISCOVERY_BACKEND` | consul | Registry backend (consul/etcd) |
| `CONSUL_ADDRESS` | 127.0.0.1:8500 | Consul server address |
| `ETCD_ENDPOINTS` | localhost:2379 | etcd endpoints |

## Next Steps

- [Docker Deployment](docker.md) — Local development with Docker
- [Contributing](../contributing.md) — Contribute to the project
