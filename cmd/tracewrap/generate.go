// cmd/generate.go

package cmd

import (
	"github.com/spf13/cobra"
)

// generateCmd is the parent command for generating artifacts in tracewrap.
// It serves as a container for various generation tasks.
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate various artifacts for tracewrap.",
	Long: `The generate command serves as a parent for subcommands that create
artifacts such as call graphs, configuration templates, and more for tracewrap.`,
	// No Run functionality; this command exists solely to group subcommands.
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
