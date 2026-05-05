# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git curl

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go-rpc-server ./examples/go-server/main.go

# Final stage
FROM alpine:3.19

RUN apk add --no-cache curl tini

# Create non-root user
RUN addgroup -g 1000 -S appgroup && adduser -u 1000 -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /go-rpc-server /usr/local/bin/go-rpc-server

# Run as non-root user
USER appuser
ENTRYPOINT ["tini", "--"]
CMD ["go-rpc-server"]
