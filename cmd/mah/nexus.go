package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var nexusCmd = &cobra.Command{
	Use:   "nexus",
	Short: "Manage nexuses (logical server groupings)",
	Long: `Nexus commands allow you to manage logical groupings of servers and services.
A nexus represents an elastic connection point to your infrastructure.`,
}

var nexusListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured nexuses",
	RunE: func(cmd *cobra.Command, args []string) error {
		nexuses, err := nexusManager.List()
		if err != nil {
			return fmt.Errorf("failed to list nexuses: %w", err)
		}

		if len(nexuses) == 0 {
			fmt.Println("No nexuses configured.")
			return nil
		}

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		defer w.Flush()

		// Header
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", 
			color.CyanString("NAME"), 
			color.CyanString("ENVIRONMENT"),
			color.CyanString("SERVERS"),
			color.CyanString("DESCRIPTION"),
			color.CyanString("STATUS"))

		// Current nexus
		current := configManager.GetCurrentNexus()

		// Rows
		for _, nexus := range nexuses {
			name := nexus.Name
			if nexus.Name == current {
				name = color.GreenString("* %s", nexus.Name)
			}

			status := color.YellowString("UNKNOWN")
			serverCount := fmt.Sprintf("%d", len(nexus.Servers))

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				name,
				nexus.Environment,
				serverCount,
				nexus.Description,
				status)
		}

		return nil
	},
}

var nexusSwitchCmd = &cobra.Command{
	Use:   "switch <nexus-name>",
	Short: "Switch to a different nexus",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nexusName := args[0]

		if err := nexusManager.Switch(nexusName); err != nil {
			return fmt.Errorf("failed to switch nexus: %w", err)
		}

		fmt.Printf("%s Switched to nexus: %s\n", 
			color.GreenString("✓"), 
			color.CyanString(nexusName))

		return nil
	},
}

var nexusCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the currently active nexus",
	RunE: func(cmd *cobra.Command, args []string) error {
		current, err := nexusManager.GetCurrent()
		if err != nil {
			return fmt.Errorf("failed to get current nexus: %w", err)
		}

		fmt.Printf("Current nexus: %s\n", color.CyanString(current.Name))
		fmt.Printf("Environment: %s\n", current.Environment)
		fmt.Printf("Description: %s\n", current.Description)
		fmt.Printf("Servers: %d\n", len(current.Servers))

		return nil
	},
}

var nexusStatusCmd = &cobra.Command{
	Use:   "status [nexus-name]",
	Short: "Show nexus status and health",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var nexusName string
		if len(args) > 0 {
			nexusName = args[0]
		} else {
			current, err := nexusManager.GetCurrent()
			if err != nil {
				return fmt.Errorf("no current nexus set and no nexus specified")
			}
			nexusName = current.Name
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		status, err := nexusManager.Status(ctx, nexusName)
		if err != nil {
			return fmt.Errorf("failed to get nexus status: %w", err)
		}

		// Print nexus header
		fmt.Printf("╭─ Nexus: %s ", color.CyanString(nexusName))
		for i := 0; i < 50-len(nexusName); i++ {
			fmt.Print("─")
		}
		fmt.Println("╮")

		// Overall status
		healthIcon := color.GreenString("✓")
		healthText := color.GreenString("HEALTHY")
		if !status.Healthy {
			healthIcon = color.RedString("✗")
			healthText = color.RedString("UNHEALTHY")
		}

		fmt.Printf("│ Status: %s %s │ Servers: %s%d/%d%s online          │\n",
			healthIcon, healthText,
			color.CyanString(""),
			status.ServersOnline, 
			status.ServersTotal,
			color.CyanString(""))

		// Server details
		if len(status.ServerStatuses) > 0 {
			fmt.Println("│                                                    │")
			fmt.Printf("│ %-50s │\n", color.CyanString("Server Details:"))

			for serverID, serverStatus := range status.ServerStatuses {
				statusIcon := color.RedString("✗")
				statusText := "OFFLINE"
				if serverStatus.Online {
					statusIcon = color.GreenString("✓")
					statusText = "ONLINE"
				}

				fmt.Printf("│ %s %-20s %s", statusIcon, serverID, statusText)

				if serverStatus.Resources != nil {
					fmt.Printf(" │ CPU: %s%.1f%%%s │ RAM: %s%.1f%%%s │",
						color.YellowString(""),
						serverStatus.Resources.CPU.Usage,
						color.YellowString(""),
						color.YellowString(""),
						serverStatus.Resources.Memory.Usage,
						color.YellowString(""))
				}

				fmt.Println()

				if serverStatus.Error != "" {
					fmt.Printf("│   Error: %s\n", color.RedString(serverStatus.Error))
				}
			}
		}

		fmt.Println("╰────────────────────────────────────────────────────╯")

		return nil
	},
}

func init() {
	nexusCmd.AddCommand(nexusListCmd)
	nexusCmd.AddCommand(nexusSwitchCmd)
	nexusCmd.AddCommand(nexusCurrentCmd)  
	nexusCmd.AddCommand(nexusStatusCmd)
}