# Authentication

This guide covers configuring authentication for go-rpc services.

## Overview

go-rpc supports multiple authentication mechanisms:

| Mechanism | Use Case |
|-----------|----------|
| Token-based | API key / bearer token authentication |
| OAuth2 | Third-party service integration |
| mTLS | Service-to-service mutual TLS |

## Token-Based Authentication

### Configuration

```yaml
# config.yaml
auth:
  type: token
  token_provider: "static"
  static:
    tokens:
      - "your-service-token"
  header: "Authorization"
  prefix: "Bearer"
```

### Go Configuration

```go
import "github.com/desyang-hub/go-rpc/pkg/middleware"

// Token validation function
validator := func(token string) bool {
    return token == "your-service-token"
}

authMiddleware := middleware.Auth(validator)

srv := server.NewServer().
    AddMiddleware(authMiddleware).
    Build()
```

### Client Example

```go
cl := client.NewClient().
    Address("server:50051").
    Token("your-service-token").
    Build()

// Or use metadata manually
conn := client.Dial("server:50051", client.WithUnaryInterceptor(
    metadata.AppendToOutgoingContext("Authorization", "Bearer your-token"),
))
```

## mTLS Authentication

### Configuration

```yaml
# config.yaml
auth:
  type: mtls
  tls:
    ca_cert: "/path/to/ca.crt"
    cert: "/path/to/server.crt"
    key: "/path/to/server.key"
    exclude_client_certs: false  # Require client certs

  mtls:
    require_client_cert: true
    client_ca_cert: "/path/to/client-ca.crt"
```

### Go Configuration

```go
import "crypto/tls"
import "crypto/x509"
import "os"

// Load TLS certificates
caCert, _ := os.ReadFile("/path/to/ca.crt")
cert, _ := tls.LoadX509KeyPair("/path/to/server.crt", "/path/to/server.key")

caPool := x509.NewCertPool()
caPool.AppendCertsFromPEM(caCert)

tlsConfig := &tls.Config{
    ClientCAs: caPool,
    ClientAuth: tls.RequireAndVerifyClientCert,
    Certificates: []tls.Certificate{cert},
}

srv := server.NewServer().
    TLS(tlsConfig).
    Build()
```

## Service-to-Service Authentication

### Configuration Example

```yaml
# config.yaml
auth:
  type: jwt
  jwt:
    issuer: "my-auth-server"
    audience: "my-service"
    jwks_url: "https://auth.my-domain.com/.well-known/jwks.json"
```

### Go Configuration

```go
import "github.com/desyang-hub/go-rpc/pkg/middleware"

jwtMiddleware := middleware.JWT(
    middleware.WithIssuer("my-auth-server"),
    middleware.WithAudience("my-service"),
    middleware.WithJWKSURL("https://auth.my-domain.com/.well-known/jwks.json"),
)

srv := server.NewServer().
    AddMiddleware(jwtMiddleware).
    Build()
```

## Important Notes

1. **Token Rotation**: Implement token rotation strategies for production
2. **Certificate Expiry**: Set up monitoring for cert expiry dates
3. **Audit Logging**: Log all authentication attempts in production

## Next Steps

- [Rate Limiting](rate-limiting.md) — Protect against abuse
- [Kubernetes Deployment](../deployment/kubernetes.md) — Deploy with secrets management
