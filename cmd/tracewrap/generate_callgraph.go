// cmd/tracewrap/generate_callgraph.go

package cmd

import (
	"fmt"
	"os"

	"github.com/mwiater/tracewrap/pkg/instrument"
	"github.com/spf13/cobra"
)

var logFile string

// callgraphCmd is the subcommand under generate for generating a call graph.
var callgraphCmd = &cobra.Command{
	Use:   "callgraph",
	Short: "Generate a call graph from a tracewrap log file.",
	Long:  `Parses the specified tracewrap.log file and generates a callgraph.dot file in the same directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		if logFile == "" {
			fmt.Println("Please specify the path to the tracewrap log file using the --log flag.")
			os.Exit(1)
		}
		if err := instrument.ParseLogAndGenerateCallGraph(logFile); err != nil {
			fmt.Printf("Error generating call graph: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Call graph generated successfully.")
	},
}

func init() {
	generateCmd.AddCommand(callgraphCmd)
	callgraphCmd.Flags().StringVar(&logFile, "log", "", "Path to the tracewrap.log file")
}
