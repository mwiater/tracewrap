// cmd/tracewrap/list_commands.go

/*
Package cmd contains the root command and all subcommand implementations.
This file defines the 'list commands' subcommand, which provides
a way to list all available commands and subcommands in a hierarchical format.
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// commandsCmd represents the 'list commands' subcommand.
// It lists all available commands and subcommands of the tracewrap tool
// in a hierarchical, indented format with two-column alignment.
var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "List all available commands and subcommands in two columns",
	Long: `The 'commands' subcommand lists all available commands and subcommands
in a hierarchical, indented format, presented in two columns for better readability.
The first column displays the command path, and the second column shows the
short description of each command.

This provides a comprehensive overview of all the actions that can be performed
by the tracewrap tool, with a clear and aligned output.`,
	Run: func(cmd *cobra.Command, args []string) {
		listAllCommands(rootCmd) // Call the listing function without initial indent
	},
}

// init adds the commands subcommand to the list command.
func init() {
	listCmd.AddCommand(commandsCmd)
}

// listAllCommands recursively traverses the command tree starting from the given cmd,
// and prints each command's name and short description in a two-column format.
// The first column is for the command path (indented for hierarchy), and the
// second column is for the short description, aligned for readability.
func listAllCommands(rootCmd *cobra.Command) {
	commandData := collectCommandData(rootCmd, "", "")

	maxPathLength := 0
	for _, data := range commandData {
		if len(data.path) > maxPathLength {
			maxPathLength = len(data.path)
		}
	}

	fmt.Println("Commands and Subcommands:")
	for _, data := range commandData {
		fmt.Printf("  %s%s%s\n", data.path, strings.Repeat(" ", maxPathLength-len(data.path)+2), data.description) // +2 for a little extra space
	}
}

type commandInfo struct {
	path        string
	description string
}

func collectCommandData(cmd *cobra.Command, currentPath string, indent string) []commandInfo {
	var allData []commandInfo

	fullPath := currentPath + cmd.Name()
	if currentPath != "" {
		fullPath = currentPath + " " + cmd.Name()
	}

	data := commandInfo{
		path:        indent + fullPath,
		description: cmd.Short,
	}
	allData = append(allData, data)

	for _, subCmd := range cmd.Commands() {
		allData = append(allData, collectCommandData(subCmd, fullPath, indent+"  ")...) // Indent subcommands further with two spaces
	}

	return allData
}
