// Package main provides the rpc-gen CLI tool for generating RPC code.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go-rpc/generators"
	"go-rpc/generators/go_generator"
	"go-rpc/generators/python_generator"
	"go-rpc/generators/typescript_generator"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "rpc-gen",
		Short: "RPC code generation tool",
		Long:  "A powerful CLI tool for generating RPC code in multiple languages from proto definitions.",
	}

	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(genCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("rpc-gen version %s\n", version)
		},
	}
}

func genCmd() *cobra.Command {
	var (
		protoPath     string
		lang          string
		outputDir     string
		packageName   string
		dryRun        bool
		generateServer bool
		generateReact bool
	)

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate RPC code from proto file",
		Long:  "Generate client/server code from proto file for the specified language.",
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
			registry.Register(generators.NewPythonPlugin())
			registry.Register(generators.NewTypeScriptPlugin())

			// Get the plugin
			plugin, err := registry.Get(lang)
			if err != nil {
				return fmt.Errorf("failed to get plugin: %w", err)
			}

			// Build generation config
			config := generators.GenConfig{
				ProtoFile:   protoPath,
				OutputDir:   outputDir,
				PackageName: packageName,
				Language:    lang,
				DryRun:      dryRun,
				Options: map[string]string{
					"generate_server": fmt.Sprint(generateServer),
					"generate_react":  fmt.Sprint(generateReact),
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
	cmd.Flags().BoolVar(&generateReact, "react", false, "Generate React Hooks (TypeScript only)")

	return cmd
}
