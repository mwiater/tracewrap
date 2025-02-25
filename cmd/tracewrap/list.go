// cmd/tracewrap/list.go

/*
Package cmd contains the root command and all subcommand implementations.
This file specifically defines the 'list' command, which acts as a parent
command to group listing related subcommands.
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd represents the 'list' command.
// It serves as a command group for listing related subcommands,
// such as 'list commands' to show all available commands.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Group commands for listing resources",
	Long: `The 'list' command is used to group subcommands that provide
different ways to list resources or information related to tracewrap.
It does not perform any actions on its own but serves as a namespace
for more specific listing commands.`,
	// No Run functionality; this command exists solely to group subcommands.
}

// init adds the list command to the root command.
func init() {
	rootCmd.AddCommand(listCmd)
}
