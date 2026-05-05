package generators

import (
	"testing"
)

// ─── ToCamelCase Tests ──────────────────────────────────────────────────────

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "helloWorld"},
		{"Hello World", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"hello", "hello"},
		{"HELLO_WORLD", "helloWorld"},
		{"HelloWorld", "helloWorld"},
		{"", ""},
		{"A", "a"},
		{"A_B_C", "aBC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ─── ToSnakeCase Tests ───────────────────────────────────────────────────────

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"helloWorld", "hello_world"},
		{"HelloWorld", "hello_world"},
		{"hello_world", "hello_world"},
		{"hello", "hello"},
		{"Hello World", "hello_world"},
		{"XMLParser", "xml_parser"},
		{"ID", "id"},
		{"HTTPServer", "http_server"},
		{"", ""},
		{"A", "a"},
		{"ABC", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ─── ToPascalCase Tests ──────────────────────────────────────────────────────

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "HelloWorld"},
		{"helloWorld", "HelloWorld"},
		{"hello world", "HelloWorld"},
		{"hello", "Hello"},
		{"HELLO_WORLD", "HelloWorld"},
		{"HelloWorld", "HelloWorld"},
		{"", ""},
		{"a", "A"},
		{"hello-world", "HelloWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
