package generators

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
	"unicode"
)

type OutputType string

const (
	OutputSingleFile OutputType = "single"
	OutputPerService OutputType = "per-service"
)

func (o OutputType) String() string {
	return string(o)
}

// BuildEngine manages template execution and output generation.
type BuildEngine struct {
	templates   map[string]*template.Template
	once        sync.Once
	initErr     error
	outputDir   string
	packageName string
	protoFile   *ProtoFile
	now         time.Time
	dryRun      bool
	outputType  OutputType
}

// EngineConfig configures the engine behavior.
type EngineConfig struct {
	OutputDir   string
	PackageName string
	DryRun      bool
	OutputType  OutputType
	Now         time.Time // Optional: override timestamp for templates
}

func (c EngineConfig) withDefaults() EngineConfig {
	if c.OutputType == "" {
		c.OutputType = OutputPerService
	}
	if c.Now.IsZero() {
		c.now = time.Now()
	}
	return c
}

// NewEngine creates a new BuildEngine with the given configuration.
func NewEngine(cfg EngineConfig) *BuildEngine {
	cfg.withDefaults()
	return &BuildEngine{
		outputDir:   cfg.OutputDir,
		packageName: cfg.PackageName,
		dryRun:      cfg.DryRun,
		outputType:  cfg.OutputType,
		templates:   make(map[string]*template.Template),
		now:         time.Now(),
	}
}

// SetProtoFile sets the parsed proto file data for use in templates.
func (e *BuildEngine) SetProtoFile(pf *ProtoFile) {
	e.protoFile = pf
}

// LoadTemplate adds a template by name and source.
func (e *BuildEngine) LoadTemplate(name, source string) error {
	funcMap := template.FuncMap{
		"upper":     func(s string) string { return strings.ToUpper(s) },
		"lower":     func(s string) string { return strings.ToLower(s) },
		"camel":     ToCamelCase,
		"snake":     ToSnakeCase,
		"pascal":    ToPascalCase,
		"timestamp": func() string { return e.now.Format(time.RFC3339) },
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(source)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	e.templates[name] = tmpl
	return nil
}

// Render executes a template with the provided data and returns the result as a string.
func (e *BuildEngine) Render(name string, data interface{}) (string, error) {
	tmpl, ok := e.templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not loaded", name)
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", name, err)
	}

	return buf.String(), nil
}

// RenderToFile executes a template and writes the result to a file.
// If dryRun is enabled, it prints to stdout instead.
func (e *BuildEngine) RenderToFile(name, outputFile string, data interface{}) error {
	content, err := e.Render(name, data)
	if err != nil {
		return err
	}

	if e.dryRun {
		fmt.Printf("=== DRY RUN: %s ===\n%s\n---\n", outputFile, content)
		return nil
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Write file
	return os.WriteFile(outputFile, []byte(content), 0644)
}

// GenerateOutput generates all output files for a service based on loaded templates.
func (e *BuildEngine) GenerateOutput(svc Service, svcData map[string]interface{}, templates map[string]string) error {
	for _, name := range getTemplateNames(svcData, svc) {
		source := templates[name]
		if source == "" {
			continue
		}

		err := e.RenderToFile(name, filepath.Join(e.outputDir, name), svcData)
		if err != nil {
			return fmt.Errorf("failed to generate file for service %s: %w", svc.Name, err)
		}
	}
	return nil
}

// ─── String Processing Helpers ────────────────────────────────────────────

// ToCamelCase converts a string to camelCase.
// "hello_world" -> "helloWorld", "Hello World" -> "helloWorld"
func ToCamelCase(s string) string {
	s = ToPascalCase(s)
	if len(s) > 0 {
		return strings.ToLower(s[:1]) + s[1:]
	}
	return s
}

// ToSnakeCase converts a string to snake_case.
// "helloWorld" -> "hello_world", "HelloWorld" -> "hello_world"
func ToSnakeCase(s string) string {
	result := make([]rune, 0, len(s)+1)
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := rune(s[i-1])
				// Don't add underscore before a capital letter if previous is also capital
				if !(prev >= 'A' && prev <= 'Z') {
					result = append(result, '_')
				}
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// ToPascalCase converts a string to PascalCase.
// "hello_world" -> "HelloWorld", "helloWorld" -> "HelloWorld"
func ToPascalCase(s string) string {
	// Split by non-alphanumeric characters
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	result := make([]string, len(words))
	for i, w := range words {
		if len(w) > 0 {
			result[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		} else {
			result[i] = w
		}
	}
	return strings.Join(result, "")
}

// getTemplateNames returns template names from the data map, or defaults.
func getTemplateNames(data map[string]interface{}, svc Service) []string {
	if names, ok := data["templates"].(map[string]string); ok {
		n := make([]string, 0, len(names))
		for k := range names {
			n = append(n, k)
		}
		return n
	}
	// Default: generate all standard templates for the service
	return []string{
		"service_interface.tmpl",
		"service_client.tmpl",
		"service_server.tmpl",
	}
}
