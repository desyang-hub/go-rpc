---
layout: home
title: Go-RPC Documentation
hero:
  name: Go-RPC
  text: Enterprise RPC Framework
  tagline: Cross-language RPC framework built for production environments.
  image:
    src: /logo.svg
    alt: Go-RPC logo
  actions:
    - theme: brand
      text: Get Started
      link: /getting-started/quick-start/
    - theme: alt
      text: Architecture
      link: /architecture/overview/
    - theme: alt
      text: Find on GitHub
      link: https://github.com/desyang-hub/go-rpc
---

## Key Features

<table>
  <thead>
    <tr>
      <th>Feature</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>gRPC Standard</strong></td>
      <td>Based on Protocol Buffers and HTTP/2, supporting all four call modes</td>
    </tr>
    <tr>
      <td><strong>Service Discovery</strong></td>
      <td>Consul and etcd backends, switchable via configuration</td>
    </tr>
    <tr>
      <td><strong>Load Balancing</strong></td>
      <td>Round-robin, weighted round-robin, least connections, consistent hashing</td>
    </tr>
    <tr>
      <td><strong>Circuit Breaker</strong></td>
      <td>Google SRE algorithm with custom degradation logic</td>
    </tr>
    <tr>
      <td><strong>Observability</strong></td>
      <td>OpenTelemetry tracing + Prometheus metrics</td>
    </tr>
    <tr>
      <td><strong>Cross-Language</strong></td>
      <td>Auto-generate Python/TypeScript client code via `rpc-gen`</td>
    </tr>
    <tr>
      <td><strong>Pluggable Architecture</strong></td>
      <td>Interceptors, discovery backends, and load balancers are all swappable</td>
    </tr>
  </tbody>
</table>
