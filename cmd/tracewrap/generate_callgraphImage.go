// cmd/tracewrap/generate_callgraphImage.go

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var dotFile string

// callgraphImageCmd is the subcommand under generate that generates a PNG image from a callgraph.dot file.
var callgraphImageCmd = &cobra.Command{
	Use:   "callgraphImage",
	Short: "Generate a PNG image from a callgraph.dot file.",
	Long: `This command takes a callgraph.dot file and generates a PNG image (callgraph.png)
in the same directory using Graphviz's dot tool.
It first checks whether Graphviz is installed and then runs the command:
  dot -Tpng -o <directory>/callgraph.png <dotfile>`,
	Run: func(cmd *cobra.Command, args []string) {
		if dotFile == "" {
			fmt.Println("Please specify the path to the callgraph.dot file using the --dotfile flag.")
			os.Exit(1)
		}

		// Check if Graphviz's dot command is installed.
		if _, err := exec.LookPath("dot"); err != nil {
			fmt.Println("Graphviz is not installed. Please install Graphviz to use this command.")
			os.Exit(1)
		}

		// Determine the output file path (same directory as the dot file).
		dir := filepath.Dir(dotFile)
		outputFile := filepath.Join(dir, "callgraph.png")

		// Run the dot command to generate the PNG image.
		cmdExec := exec.Command("dot", "-Tpng", "-o", outputFile, dotFile)
		if err := cmdExec.Run(); err != nil {
			fmt.Printf("Error generating PNG image: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("PNG image generated successfully at: %s\n", outputFile)
	},
}

func init() {
	generateCmd.AddCommand(callgraphImageCmd)
	callgraphImageCmd.Flags().StringVar(&dotFile, "dotfile", "", "Path to the callgraph.dot file")
}
