// Package genapi provides stub types for the HelloService.
// This is a placeholder - run `make gen-proto` to generate real code from proto definitions.
package genapi

import (
	"context"

	"google.golang.org/grpc"
)

// ==================== Request/Response Types ====================

// HelloRequest defines a request message structure.
type HelloRequest struct {
	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	GreetType string `protobuf:"bytes,2,opt,name=greet_type,json=greetType,proto3" json:"greet_type,omitempty"`
}

func (x *HelloRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *HelloRequest) GetGreetType() string {
	if x != nil {
		return x.GreetType
	}
	return ""
}

// HelloResponse defines a response message structure.
type HelloResponse struct {
	Message   string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Timestamp string `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	ServerId  string `protobuf:"bytes,3,opt,name=server_id,json=serverId,proto3" json:"server_id,omitempty"`
}

func (x *HelloResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *HelloResponse) GetTimestamp() string {
	if x != nil {
		return x.Timestamp
	}
	return ""
}

func (x *HelloResponse) GetServerId() string {
	if x != nil {
		return x.ServerId
	}
	return ""
}

// HelloStreamRequest is a streaming request message.
type HelloStreamRequest struct {
	Name   string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Index  int32  `protobuf:"varint,2,opt,name=index,proto3" json:"index,omitempty"`
}

func (x *HelloStreamRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *HelloStreamRequest) GetIndex() int32 {
	if x != nil {
		return x.Index
	}
	return 0
}

// HelloStreamResponse is a streaming response message.
type HelloStreamResponse struct {
	Message       string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	ResponseIndex int32  `protobuf:"varint,2,opt,name=response_index,json=responseIndex,proto3" json:"response_index,omitempty"`
}

func (x *HelloStreamResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *HelloStreamResponse) GetResponseIndex() int32 {
	if x != nil {
		return x.ResponseIndex
	}
	return 0
}

// ==================== Service Interfaces ====================

// HelloServiceServer is the server-side interface for HelloService.
type HelloServiceServer interface {
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	HelloStream(*HelloRequest, HelloService_HelloStreamServer) error
	BatchHello(HelloService_BatchHelloServer) error
	HelloStreamStream(HelloService_HelloStreamStreamServer) error
}

// HelloService_HelloStreamServer is the server stream for HelloStream.
type HelloService_HelloStreamServer interface {
	Send(*HelloStreamResponse) error
	Context() context.Context
}

// HelloService_BatchHelloServer is the client stream for BatchHello.
type HelloService_BatchHelloServer interface {
	Recv() (*HelloRequest, error)
	SendAndClose(*HelloResponse) error
	Context() context.Context
}

// HelloService_HelloStreamStreamServer is the bidirectional stream for HelloStreamStream.
type HelloService_HelloStreamStreamServer interface {
	Recv() (*HelloStreamRequest, error)
	Send(*HelloStreamResponse) error
	Context() context.Context
}

// RegisterHelloServiceServer registers a HelloServiceServer implementation.
func RegisterHelloServiceServer(s *grpc.Server, srv HelloServiceServer) {
	// This is a stub - use protoc-gen-go for real registration
	_ = s
	_ = srv
}

// ServiceDesc is a simplified service description.
type ServiceDesc struct {
	ServiceName string
	MethodName  string
	Handler     interface{}
}
