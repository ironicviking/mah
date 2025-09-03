package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services",
	Long:  "Service commands allow you to deploy, scale, and manage containerized services.",
}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List services in current nexus",
	RunE: func(cmd *cobra.Command, args []string) error {
		current, err := nexusManager.GetCurrent()
		if err != nil {
			return fmt.Errorf("no current nexus: %w", err)
		}

		config := configManager.GetConfig()
		if config == nil {
			return fmt.Errorf("no configuration loaded")
		}

		fmt.Printf("Services configured for nexus '%s':\n", current.Name)

		foundServices := false
		for serviceName, service := range config.Services {
			// Check if this service is deployed to any server in current nexus
			for _, serverName := range service.Servers {
				for _, nexusServer := range current.Config.Servers {
					if serverName == nexusServer {
						if !foundServices {
							foundServices = true
						}
						fmt.Printf("  - %s (image: %s, servers: %v)\n", 
							serviceName, service.Image, service.Servers)
						goto nextService
					}
				}
			}
		nextService:
		}

		if !foundServices {
			fmt.Println("  No services configured for this nexus.")
		}

		return nil
	},
}

var serviceDeployCmd = &cobra.Command{
	Use:   "deploy <service-name>",
	Short: "Deploy a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]
		
		config := configManager.GetConfig()
		if config == nil {
			return fmt.Errorf("no configuration loaded")
		}

		service := config.Services[serviceName]
		if service == nil {
			return fmt.Errorf("service '%s' not found in configuration", serviceName)
		}

		fmt.Printf("Deploying service '%s'...\n", serviceName)
		fmt.Printf("  Image: %s\n", service.Image)
		fmt.Printf("  Servers: %v\n", service.Servers)
		
		// This will be implemented when we have the full service management
		return fmt.Errorf("service deployment not yet implemented")
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status [service-name]",
	Short: "Show service status",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// This will be implemented when we have the full service management
		return fmt.Errorf("service status command not yet implemented")
	},
}

var serviceLogsCmd = &cobra.Command{
	Use:   "logs <service-name>",
	Short: "Show service logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// This will be implemented when we have the full service management
		return fmt.Errorf("service logs command not yet implemented")
	},
}

func init() {
	serviceCmd.AddCommand(serviceListCmd)
	serviceCmd.AddCommand(serviceDeployCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
	serviceCmd.AddCommand(serviceLogsCmd)
}