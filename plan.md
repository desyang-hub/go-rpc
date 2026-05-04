# 企业级跨语言 RPC 框架开发任务

请作为资深后端架构师和 Go 语言专家，完成一个企业级、高稳定、可扩展的 RPC 框架项目。项目需遵循严格的工程规范，确保生产可用，并具备跨语言调用能力。

## 一、核心目标
开发一个以 Go 语言实现的服务端与客户端 RPC 库，同时为其他主流语言（至少包含 Python、TypeScript）生成对应的客户端调用代码和接口定义，实现真正的跨语言互通。整个项目必须满足 GitHub 开源社区的企业级标准。

## 二、技术要求

### 1. RPC 协议与传输
- 使用 **Protocol Buffers (proto3)** 作为接口定义语言 (IDL)，保证强类型和高效序列化。
- 传输层基于 **HTTP/2** 或 **gRPC 标准协议**，优先直接采用 gRPC 生态（成熟可靠）。
- 支持四种调用模式：一元 RPC、服务端流、客户端流、双向流。
- 内置连接池、健康检查、Keepalive 机制。
- 支持 TLS/SSL 加密和基于 Token/OAuth2 的认证拦截器。

### 2. 服务治理（Go 服务端/客户端库）
- 服务注册与发现：抽象接口，至少提供 **Consul** 和 **etcd** 两种实现，支持通过配置文件切换。
- 负载均衡：支持轮询、加权轮询、最少连接、一致性哈希等策略。
- 熔断与降级：集成 **Google SRE 熔断器** 或 **Hystrix-go** 类似组件，支持降级逻辑。
- 超时控制、重试机制（指数退避）、速率限制。
- 可观测性：自动集成 **OpenTelemetry** 追踪和 **Prometheus** 指标暴露，同时输出结构化日志（使用 `zerolog` 或 `zap`）。
- 插件化拦截器链：允许用户自定义中间件（如日志、鉴权、指标收集）。

### 3. 跨语言代码生成工具
- 提供 CLI 工具 `rpc-gen`，输入 `.proto` 文件路径，自动生成：
  - Go 语言的 gRPC 桩代码和服务骨架。
  - Python 的 gRPC 客户端桩代码以及封装好的 `Client` 类，使其调用方式接近原生 Python 库。
  - TypeScript 的 gRPC-Web 客户端桩代码，以及适配现代前端框架（如 React/Vue）的 Hook/Composable 包装。
- 生成的代码需包含完善的类型提示、错误处理和连接管理，用户可直接安装使用，无需手动适配。
- 工具需支持插件机制，允许未来扩展其他语言（如 Java、C#）。

### 4. 多语言兼容性脚本
- 提供 Shell 脚本和 Makefile，用于：
  - 基于 `proto` 文件批量生成所有目标语言的代码。
  - 运行各语言示例的集成测试（Go 服务端 + Python/TypeScript 客户端）。
- 每个语言提供独立的示例项目（example 目录），开箱即用，包含 Docker 环境。

## 三、企业级工程规范

### 1. 项目结构
必须符合 Go 社区标准布局（`cmd/`, `internal/`, `pkg/`, `api/`, `configs/`, `test/` 等），同时包含多语言工具目录：
```
/
├── api/                  # proto 定义
├── cmd/
│   ├── rpc-server/       # Go 服务端入口
│   ├── rpc-client/       # Go CLI 调试客户端
│   └── rpc-gen/          # 代码生成 CLI
├── internal/             # 内部实现（服务治理、连接池等）
├── pkg/                  # 可公开导出的库（客户端 SDK 基类）
├── plugins/              # 拦截器、注册中心、负载均衡等插件
├── scripts/              # 跨语言生成脚本、CI 脚本
├── generators/           # 各语言生成器模板与逻辑
├── examples/
│   ├── go-server/
│   ├── python-client/
│   └── ts-client/
├── docs/                 # 详细文档（架构、快速开始、API 参考）
├── .github/
│   ├── workflows/        # CI/CD 流水线
│   └── ISSUE_TEMPLATE/
├── Makefile
├── Dockerfile
├── go.mod
├── README.md
└── LICENSE
```

### 2. Git 与版本控制
- 遵循 **Conventional Commits** 规范（feat:, fix:, docs: 等）。
- 配置 pre-commit 钩子：代码格式检查（gofmt, goimports）、lint（golangci-lint）、单元测试。
- 使用 `CHANGELOG.md` 自动生成工具（如 `git-cliff`）跟进版本变更。

### 3. CI/CD（GitHub Actions）
提供至少以下工作流：
- **CI**：推送或 PR 时触发，执行 lint、全量单元测试、覆盖率检查（>80%）、代码安全扫描（gosec）、所有语言示例的集成测试。
- **Release**：推送 semver 标签时，交叉编译多平台二进制文件（Go 服务端、rpc-gen），构建并推送 Docker 镜像，发布到 GitHub Releases 并附带各语言 SDK 包（Python 发布到 PyPI，TypeScript 发布到 npm）。
- **文档构建**：自动将 `/docs` 部署到 GitHub Pages 或 ReadTheDocs。

### 4. 文档
- README.md：项目简介、特性列表、快速开始（1 分钟启动示例）、架构图链接。
- 完整的 docs 目录：含架构设计、配置参考、跨语言使用指南、API 参考、FAQ。
- 代码注释覆盖率 > 70%，为公开 API 编写 godoc 示例。

## 四、质量要求
- 单元测试覆盖核心模块，使用 `testify` 和 `gomock` 进行隔离测试。
- 所有 RPC 方法必须有集成测试，模拟网络闪断、延迟注入等场景。
- 各语言客户端必须通过相同的功能测试用例（例如基于同一个 proto 的测试套件）。
- 性能测试：提供基准测试脚本，单机四核下 QPS 不低于 10 万（简单 Echo 方法）。
- 安全扫描无严重漏洞，依赖项在合理范围内无已知严重 CVE。

## 五、可交付成果
请按以下顺序逐步输出，每个部分需完整、可直接使用：
1. 项目总体架构说明与模块设计文档。
2. 核心 proto 文件和生成的 Go 桩代码（示例）。
3. Go 服务端与客户端库的核心实现代码，包含服务治理全部功能。
4. 代码生成工具 `rpc-gen` 的实现，及为 Python、TypeScript 生成的最终客户端代码示例。
5. 跨语言集成测试脚本与示例项目的完整文件。
6. GitHub Actions 工作流配置、Makefile、Docker 文件、README 及其他工程化配置文件。

请开始实现。