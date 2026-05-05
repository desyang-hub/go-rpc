// Package main provides the rpc-gen CLI tool for generating RPC code.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/desyang-hub/go-rpc/generators"
	"github.com/desyang-hub/go-rpc/generators/go_generator"
	"github.com/desyang-hub/go-rpc/generators/python_generator"
	"github.com/desyang-hub/go-rpc/generators/typescript_generator"
)

var version = "dev"

// generateCmd builds the "gen" subcommand with all supported options.
func generateCmd() *cobra.Command {
	var (
		protoPath      string
		lang           string
		outputDir      string
		packageName    string
		dryRun         bool
		generateServer bool
		generateReact  bool
		generateGateway bool // New: Generate gRPC-gateway HTTP handlers
	)

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate RPC code from proto file",
		Long:  "Generate client/server code from proto file for the specified language with optional gateway support.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if protoPath == "" {
				return fmt.Errorf("proto file path is required (--proto)")
			}
			if lang == "" {
				return fmt.Errorf("language is required (--lang)")
			}
			if outputDir == "" {
				return fmt.Errorf("output directory is required (--output)")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse proto file
			parser := generators.NewParser("")
			parsed, err := parser.ParseFile(protoPath)
			if err != nil {
				return fmt.Errorf("failed to parse proto file: %w", err)
			}

			// Create plugin registry and register plugins
			registry := generators.NewPluginRegistry()
			registry.Register(go_generator.NewGoPlugin())
			registry.Register(python_generator.NewPythonPlugin())
			registry.Register(typescript_generator.NewTypeScriptPlugin())

			// Get the plugin
			plugin, err := registry.Get(lang)
			if err != nil {
				return fmt.Errorf("failed to get plugin: %w", err)
			}

			// Build generation config with gateway option
			config := generators.GenConfig{
				ProtoFile:   protoPath,
				OutputDir:   outputDir,
				PackageName: packageName,
				Language:    lang,
				DryRun:      dryRun,
				Options: map[string]string{
					"generate_server":      fmt.Sprint(generateServer),
					"generate_react":       fmt.Sprint(generateReact),
					"generate_gateway":     fmt.Sprint(generateGateway),
					"gateway_port":         "8081",
					"gateway_host":         "localhost",
				},
			}

			// Generate code
			files, err := plugin.Generate(config, parsed)
			if err != nil {
				return fmt.Errorf("code generation failed: %w", err)
			}

			// Print summary
			fmt.Printf("Generated %d file(s):\n", len(files))
			for _, f := range files {
				fmt.Printf("  %s (%s)\n", f.Path, f.Description)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&protoPath, "proto", "", "Path to proto file (required)")
	cmd.Flags().StringVar(&lang, "lang", "", "Target language: go, python, typescript (required)")
	cmd.Flags().StringVar(&outputDir, "output", "", "Output directory (required)")
	cmd.Flags().StringVar(&packageName, "package", "", "Package name for output")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print output instead of writing files")
	cmd.Flags().BoolVar(&generateServer, "generate-server", false, "Generate server stubs")
	cmd.Flags().BoolVar(&generateReact, "react", false, "Generate React/Vue Hooks (TypeScript only)")
	cmd.Flags().BoolVar(&generateGateway, "gateway", false, "Generate gRPC-gateway HTTP handlers (Go)")

	cmd.MarkFlagRequired("proto")
	cmd.MarkFlagRequired("lang")
	cmd.MarkFlagRequired("output")

	return cmd
}

// versionCmd prints version information
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("rpc-gen version %s\n", version)
		},
	}
}

// main is the entry point for the rpc-gen CLI tool
func main() {
	rootCmd := &cobra.Command{
		Use:   "rpc-gen",
		Short: "RPC code generation tool",
		Long:  "A powerful CLI tool for generating RPC code in multiple languages from proto definitions.",
	}

	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(generateCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
