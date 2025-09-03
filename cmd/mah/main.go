package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/jonas-jonas/mah/internal/config"
	"github.com/jonas-jonas/mah/internal/nexus"
)

var (
	configFile    string
	currentNexus  string
	allNexuses    bool
	verbose       bool
	configManager *config.Manager
	nexusManager  *nexus.Manager
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mah",
	Short: "Modern Application Hub - Infrastructure orchestration made simple",
	Long: `MAH (Modern Application Hub) is a next-generation infrastructure orchestration tool
that provides a unified interface for managing distributed containerized applications
across multiple servers with automatic SSL, firewall management, and extensible plugin architecture.

Examples:
  mah nexus list                    # List all nexuses
  mah nexus switch thor-prod        # Switch to thor-prod nexus
  mah service deploy blog           # Deploy blog service to current nexus
  mah server status                 # Show server status for current nexus`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration manager
		configManager = config.NewManager()
		
		// Load runtime config first
		if err := configManager.LoadRuntimeConfig(); err != nil {
			return fmt.Errorf("failed to load runtime config: %w", err)
		}
		
		// Load main config if it exists
		if _, err := os.Stat(configFile); err == nil {
			if err := configManager.LoadConfig(configFile); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
		}
		
		// Initialize nexus manager
		nexusManager = nexus.NewManager(configManager)
		
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "mah.yaml", "config file")
	rootCmd.PersistentFlags().StringVarP(&currentNexus, "nexus", "n", "", "target specific nexus")
	rootCmd.PersistentFlags().BoolVarP(&allNexuses, "all", "a", false, "operate on all nexuses")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(nexusCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(configCmd)
}