package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage servers",
	Long:  "Server commands allow you to initialize, update, and manage individual servers.",
}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List servers in current nexus",
	RunE: func(cmd *cobra.Command, args []string) error {
		current, err := nexusManager.GetCurrent()
		if err != nil {
			return fmt.Errorf("no current nexus: %w", err)
		}

		fmt.Printf("Servers in nexus '%s':\n", current.Name)
		for _, server := range current.Servers {
			fmt.Printf("  - %s (%s)\n", server.ID(), server.Host())
		}

		return nil
	},
}

var serverStatusCmd = &cobra.Command{
	Use:   "status [server-name]",
	Short: "Show server status",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// This will be implemented when we have the full server management
		return fmt.Errorf("server status command not yet implemented")
	},
}

var serverInitCmd = &cobra.Command{
	Use:   "init <server-name>",
	Short: "Initialize a server (install Docker, configure firewall, etc.)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// This will be implemented when we have the full server management
		return fmt.Errorf("server init command not yet implemented")
	},
}

func init() {
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverStatusCmd)
	serverCmd.AddCommand(serverInitCmd)
}