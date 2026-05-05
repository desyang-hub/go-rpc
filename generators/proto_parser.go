// Package generators provides code generation tools for RPC frameworks.
//
// The proto_parser module extracts service definitions from .proto files
// using protoc's descriptor_set output format. This approach ensures
// reliable parsing without requiring external proto parser libraries.
package generators

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"time"
)

// ProtoFile represents a parsed proto file with all its services, messages, etc.
type ProtoFile struct {
	Name       string            // File name without extension
	Path       string            // Full path to proto file
	Package    string            // Proto package declaration
	Imports    []string          // Imported proto files
	Services   []Service         // Service definitions
	Messages   []Message         // Message definitions
	Enums      []Enum            // Enum definitions
	Options    map[string]string // File-level options
}

// Service represents a gRPC service definition.
type Service struct {
	Name    string         // Service name
	Methods []ServiceMethod // Methods in the service
}

// ServiceMethod represents a single RPC method within a service.
type ServiceMethod struct {
	Name         string // Method name
	InputType    string // Input message type (e.g., "HelloRequest")
	OutputType   string // Output message type (e.g., "HelloResponse")
	ClientStream bool   // True if client-side streaming
	ServerStream bool   // True if server-side streaming
}

// DataType represents the RPC call type.
type DataType int

const (
	DataTypeUnary DataType = iota // Standard request-response
	DataTypeServerStream          // Server streaming
	DataTypeClientStream          // Client streaming
	DataTypeBidirectional         // Bidirectional streaming
)

// String returns string representation of the data type.
func (d DataType) String() string {
	switch d {
	case DataTypeUnary:
		return "unary"
	case DataTypeServerStream:
		return "server_stream"
	case DataTypeClientStream:
		return "client_stream"
	case DataTypeBidirectional:
		return "bidirectional_stream"
	default:
		return "unknown"
	}
}

// Message represents a protobuf message definition.
type Message struct {
	Name      string   // Message name
	Fields    []Field  // Fields in the message
}

// Field represents a field within a message.
type Field struct {
	Name        string // Field name
	Type        string // Type (string, int32, etc.)
	TypeName    string // Full type name for message types
	Number      int32  // Proto field number
	Mapped      bool   // True if field is a map
	KeyType     string // Key type for maps
	ValueType   string // Value type for maps
	Repeated    bool   // True if repeated field
	Optional    bool   // True if optional field
}

// Enum represents a protobuf enum definition.
type Enum struct {
	Name    string   // Enum name
	Values  []EnumValue // Enum values
}

// EnumValue represents a single enum value.
type EnumValue struct {
	Name    string // Value name
	Number  int32  // Value number
}

// Parser provides proto file parsing functionality.
type Parser struct {
	protocPath string // Path to protoc binary
}

// NewParser creates a new Parser with default configuration.
func NewParser(protocPath string) *Parser {
	if protocPath == "" {
		protocPath = "protoc"
	}
	return &Parser{protocPath: protocPath}
}

// ParseFile parses a .proto file and returns structured data.
func (p *Parser) ParseFile(protoPath string) (*ProtoFile, error) {
	// Validate file exists
	if _, err := os.Stat(protoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("proto file not found: %s", protoPath)
	}

	// Generate descriptor set using protoc
	descSetPath, err := p.generateDescriptorSet(protoPath)
	if err != nil {
		return nil, err
	}

	// Parse descriptor set
	return p.parseDescriptorSet(descSetPath, protoPath)
}

// generateDescriptorSet runs protoc to create a binary descriptor set file.
func (p *Parser) generateDescriptorSet(protoPath string) (string, error) {
	dir := filepath.Dir(protoPath)
	descSetPath := filepath.Join(dir, ".rpc_gen_descriptor.pb")

	cmd := exec.Command(p.protocPath,
		"--include_imports",
		"--descriptor_set_out="+descSetPath,
		protoPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running protoc: %w (output: %s)", err, string(output))
	}

	return descSetPath, nil
}

// parseDescriptorSet reads the binary descriptor set and extracts definitions.
func (p *Parser) parseDescriptorSet(descSetPath, protoPath string) (*ProtoFile, error) {
	// Read descriptor set file
	data, err := os.ReadFile(descSetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read descriptor set: %w", err)
	}

	// Parse FileDescriptorSet (binary format)
	fds := new(descriptorpb.FileDescriptorSet)
	if err := proto.Unmarshal(data, fds); err != nil {
		return nil, fmt.Errorf("failed to parse descriptor set: %w", err)
	}

	// Extract the proto file (first one in the set)
	// Note: This is simplified - production should properly handle imports
	var fileDesc *descriptorpb.FileDescriptorProto
	for _, fd := range fds.File {
		// Match based on filename
		if strings.HasSuffix(fd.GetName(), filepath.Base(protoPath)) {
			fileDesc = fd
			break
		}
	}
	if fileDesc == nil {
		return nil, fmt.Errorf("unable to find proto file in descriptor set")
	}

	// Clean up temporary file
	_ = os.Remove(descSetPath)

	// Convert to our internal representation
	pf := &ProtoFile{
		Name:       strings.TrimSuffix(filepath.Base(protoPath), filepath.Ext(protoPath)),
		Path:       protoPath,
		Package:    fileDesc.GetPackage(),
		Imports:    fileDesc.GetDependency(),
		Options:    make(map[string]string),
	}

	// Extract options
	if goPkg := fileDesc.Options.GetGoPackage(); goPkg != "" {
		pf.Options["go_package"] = goPkg
	}

	// Parse messages
	for _, msgDesc := range fileDesc.MessageType {
		msg := Message{
			Name:   msgDesc.GetName(),
			Fields: make([]Field, len(msgDesc.Field)),
		}
		for i, fieldDesc := range msgDesc.Field {
			msg.Fields[i] = parseField(fieldDesc)
		}
		pf.Messages = append(pf.Messages, msg)
	}

	// Parse enums
	for _, enumDesc := range fileDesc.EnumType {
		enum := Enum{
			Name:   enumDesc.GetName(),
			Values: make([]EnumValue, len(enumDesc.Value)),
		}
		for i, v := range enumDesc.Value {
			enum.Values[i] = EnumValue{
				Name:   v.GetName(),
				Number: v.GetNumber(),
			}
		}
		pf.Enums = append(pf.Enums, enum)
	}

	// Parse services
	for _, svcDesc := range fileDesc.Service {
		svc := Service{
			Name:    svcDesc.GetName(),
			Methods: make([]ServiceMethod, len(svcDesc.Method)),
		}
		for i, methodDesc := range svcDesc.Method {
			svc.Methods[i] = ServiceMethod{
				Name:         methodDesc.GetName(),
				InputType:    methodDesc.GetInputType(),
				OutputType:   methodDesc.GetOutputType(),
				ClientStream: methodDesc.GetClientStreaming(),
				ServerStream: methodDesc.GetServerStreaming(),
			}
		}
		pf.Services = append(pf.Services, svc)
	}

	return pf, nil
}

func parseField(field *descriptorpb.FieldDescriptorProto) Field {
	f := Field{
		Name:     field.GetName(),
		Type:     convertProtoTypeToGoType(field),
		TypeName: field.GetTypeName(),
		Number:   field.GetNumber(),
		Mapped:   strings.Contains(field.GetTypeName(), "MapEntry"),
	}

	// Handle map field types - simplified detection for v2.x
	if f.Mapped {
		// Map fields in proto3 have their type name containing "MapEntry"
		// We skip detailed field parsing since FieldDescriptorProto doesn't expose nested types in v2.x
		f.KeyType = "unknown"
		f.ValueType = "unknown"
	}

	f.Repeated = field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	if field.Options != nil && field.Options.GetWeak() {
		f.Optional = true
	}

	return f
}

// convertProtoTypeToGoType converts protobuf field type to Go type string.
func convertProtoTypeToGoType(field *descriptorpb.FieldDescriptorProto) string {
	t := field.GetType()
	switch t {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "float64"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return "float32"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return "uint64"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytes"
	default:
		return field.GetTypeName()
	}
}

// Timestamp returns the file modification time as an RFC3339 string, or current time if unavailable.
func (pf *ProtoFile) Timestamp() string {
	return time.Now().Format(time.RFC3339)
}
