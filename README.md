# go-rpc

> 企业级、高稳定、可扩展的跨语言 RPC 框架

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 特性

- **gRPC 标准协议** — 基于 Protocol Buffers 和 HTTP/2，支持四种调用模式
- **服务注册与发现** — 支持 Consul 和 etcd，配置文件即可切换
- **负载均衡** — 轮询、加权轮询、最少连接、一致性哈希
- **熔断与降级** — 内置 Google SRE 算法熔断器，支持自定义降级逻辑
- **可观测性** — OpenTelemetry 分布式追踪 + Prometheus 指标暴露
- **跨语言互通** — 通过 `rpc-gen` 工具自动生成 Python/TypeScript 客户端代码
- **插件化架构** — 拦截器链、注册中心、负载均衡等核心组件均可插拔

## 架构图

![Architecture](docs/assets/architecture.png)

> 详细架构设计请参考 [架构设计文档](docs/architecture.md)

## 快速开始

### 环境要求

- Go 1.22+
- Docker & Docker Compose（运行示例）
- protoc + protoc-gen-go + protoc-gen-go-grpc（代码生成）

### 1. 克隆项目

```bash
git clone https://github.com/desyang/go-rpc.git
cd go-rpc
```

### 2. 初始化环境

```bash
make setup
```

### 3. 运行 Go 服务端示例

```bash
cd examples/go-server
docker compose up
```

### 4. 使用 Python 客户端调用

```bash
cd examples/python-client
pip install -r requirements.txt
python client.py
```

### 5. 生成其他语言代码

```bash
./bin/rpc-gen --proto=./api --output=./generated --languages=python,typescript
```

## 项目结构

```
.
├── api/                  # proto 定义
├── cmd/
│   ├── rpc-server/       # Go 服务端入口
│   ├── rpc-client/       # Go CLI 调试客户端
│   └── rpc-gen/          # 代码生成 CLI
├── internal/             # 内部实现
├── pkg/                  # 可导出的公共库
├── generators/           # 代码生成模板
├── examples/             # 多语言示例
├── docs/                 # 文档
└── scripts/              # CI/构建脚本
```

## 状态

| 阶段 | 模块 | 状态 |
|------|------|------|
| Phase 1 | 基础骨架、proto 定义 | 🚧 进行中 |
| Phase 2 | Core RPC Engine | ⏳ 待开始 |
| Phase 3 | Service Governance | ⏳ 待开始 |
| Phase 4 | Observability | ⏳ 待开始 |
| Phase 5 | rpc-gen 代码生成 | ⏳ 待开始 |

## 许可证

MIT License — 详见 [LICENSE](LICENSE) 文件。
