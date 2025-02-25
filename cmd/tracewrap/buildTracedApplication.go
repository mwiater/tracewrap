// cmd/tracewrap/buildTracedApplication.go

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mwiater/tracewrap/config"
	"github.com/mwiater/tracewrap/pkg/instrument"
	"github.com/spf13/cobra"
)

var (
	projectDir string
	configPath string
	appName    string
)

// buildCmd represents the buildTracedApplication command.
var buildCmd = &cobra.Command{
	Use:   "buildTracedApplication",
	Short: "Build and run an instrumented version of the application",
	Long: `buildTracedApplication builds an instrumented version of the target Go application.
It prepares the workspace, loads configuration, instruments the source, builds the binary,
optionally moves and renames it, and then executes the instrumented binary.`,
	Run: func(cmd *cobra.Command, args []string) {
		if projectDir == "" {
			fmt.Println("Project directory must be specified using --project")
			os.Exit(1)
		}
		absProjectDir, err := filepath.Abs(projectDir)
		if err != nil {
			fmt.Printf("Error determining absolute path: %v\n", err)
			os.Exit(1)
		}
		info, err := os.Stat(absProjectDir)
		if err != nil || !info.IsDir() {
			fmt.Printf("Project directory does not exist or is not a directory: %s\n", absProjectDir)
			os.Exit(1)
		}
		fmt.Println("Tracing build initiated for project:", absProjectDir)

		workspace, err := instrument.PrepareWorkspace(absProjectDir)
		if err != nil {
			fmt.Printf("Error preparing workspace: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Workspace prepared at:", workspace)

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		err = instrument.SetDynamicTracerImport(workspace)
		if err != nil {
			fmt.Printf("Error setting tracer import: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Dynamic tracer import set to:", instrument.DynamicTracerImport)

		err = instrument.InstrumentWorkspace(workspace, *cfg)
		if err != nil {
			fmt.Printf("Error instrumenting workspace: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Instrumentation completed.")

		// Build the instrumented binary.
		binaryPath, err := instrument.BuildInstrumentedBinary(workspace)
		if err != nil {
			fmt.Printf("Error building binary: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Binary built at:", binaryPath)

		// If the --name flag is provided, move the binary to the project's bin/ directory
		// and rename it as <appName>-tracewrap.
		if appName != "" {
			binDir := filepath.Join(absProjectDir, "bin")
			if err := os.MkdirAll(binDir, 0755); err != nil {
				fmt.Printf("Error creating bin directory: %v\n", err)
				os.Exit(1)
			}
			newBinaryName := appName + "-tracewrap"
			if runtime.GOOS == "windows" {
				newBinaryName += ".exe"
			}
			newBinaryPath := filepath.Join(binDir, newBinaryName)
			if err := os.Rename(binaryPath, newBinaryPath); err != nil {
				fmt.Printf("Error moving binary to bin directory: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Binary moved to:", newBinaryPath)
			binaryPath = newBinaryPath
		}

		// Run the instrumented binary, forwarding any extra arguments.
		err = instrument.RunInstrumentedBinary(binaryPath, args)
		if err != nil {
			fmt.Printf("Error running binary: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Instrumented binary execution completed.")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().StringVarP(&projectDir, "project", "p", "", "Path to the target Go project")
	buildCmd.Flags().StringVarP(&configPath, "config", "c", "tracewrap.yaml", "Path to the configuration YAML file")
	buildCmd.Flags().StringVar(&appName, "name", "", "Name of the application (binary will be moved as <name>-tracewrap)")
}
