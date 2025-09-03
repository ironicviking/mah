package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/jonas-jonas/mah/internal/plugins/docker"
	"github.com/jonas-jonas/mah/internal/server"
	"github.com/jonas-jonas/mah/pkg"
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
		return deployService(args[0])
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status [service-name]",
	Short: "Show service status",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return showAllServiceStatus()
		}
		return showServiceStatus(args[0])
	},
}

var serviceLogsCmd = &cobra.Command{
	Use:   "logs <service-name>",
	Short: "Show service logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		follow, _ := cmd.Flags().GetBool("follow")
		return showServiceLogs(args[0], follow)
	},
}

func init() {
	serviceCmd.AddCommand(serviceListCmd)
	serviceCmd.AddCommand(serviceDeployCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
	serviceCmd.AddCommand(serviceLogsCmd)
	
	// Add flags
	serviceLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
}

// deployService deploys a service to servers
func deployService(serviceName string) error {
	config := configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	service := config.Services[serviceName]
	if service == nil {
		return fmt.Errorf("service '%s' not found in configuration", serviceName)
	}

	fmt.Printf("🚀 Deploying service '%s'...\n", serviceName)
	fmt.Printf("   Image: %s\n", service.Image)
	fmt.Printf("   Servers: %v\n", service.Servers)

	// Create server instances for deployment
	servers := make(map[string]pkg.Server)
	factory := server.NewFactory()

	for _, serverName := range service.Servers {
		serverConfig := config.Servers[serverName]
		if serverConfig == nil {
			return fmt.Errorf("server '%s' not found in configuration", serverName)
		}

		srv, err := factory.CreateServer(serverName, serverConfig)
		if err != nil {
			return fmt.Errorf("failed to create server instance for '%s': %w", serverName, err)
		}
		defer srv.Disconnect()

		servers[serverName] = srv
	}

	// Create Docker provider
	dockerProvider := docker.NewProvider(servers, config)

	// Convert config.Service to pkg.ServiceConfig
	serviceConfig := &pkg.ServiceConfig{
		Name:        serviceName,
		Image:       service.Image,
		Servers:     service.Servers,
		Domains:     service.Domains,
		Public:      service.Public,
		Internal:    service.Internal,
		Ports:       service.Ports,
		Environment: service.Environment,
		Volumes:     service.Volumes,
		Networks:    service.Networks,
		Depends:     service.Depends,
		Labels:      service.Labels,
		Replicas:    service.Replicas,
	}

	// Deploy service
	err := dockerProvider.Deploy(serviceConfig)
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	color.Green("✅ Service '%s' deployed successfully!", serviceName)
	return nil
}

// showServiceStatus shows status for a specific service
func showServiceStatus(serviceName string) error {
	config := configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	service := config.Services[serviceName]
	if service == nil {
		return fmt.Errorf("service '%s' not found in configuration", serviceName)
	}

	// Create server instances
	servers := make(map[string]pkg.Server)
	factory := server.NewFactory()

	for _, serverName := range service.Servers {
		serverConfig := config.Servers[serverName]
		if serverConfig == nil {
			continue
		}

		srv, err := factory.CreateServer(serverName, serverConfig)
		if err != nil {
			continue
		}
		defer srv.Disconnect()

		servers[serverName] = srv
	}

	if len(servers) == 0 {
		return fmt.Errorf("no accessible servers found for service '%s'", serviceName)
	}

	// Create Docker provider and get status
	dockerProvider := docker.NewProvider(servers, config)
	status, err := dockerProvider.Status(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	fmt.Printf("📊 Service Status: %s\n", serviceName)
	fmt.Println("─────────────────────────────────────")

	// Display status with colors
	fmt.Print("🏃 Status: ")
	switch status.Status {
	case "running":
		color.Green("RUNNING")
	case "stopped":
		color.Red("STOPPED")
	case "not_deployed":
		color.Yellow("NOT DEPLOYED")
	default:
		color.Yellow(strings.ToUpper(status.Status))
	}

	fmt.Print("💓 Health: ")
	switch status.Health {
	case "healthy":
		color.Green("HEALTHY")
	case "unhealthy":
		color.Red("UNHEALTHY")
	default:
		color.Yellow(strings.ToUpper(status.Health))
	}

	if status.Replicas > 0 {
		fmt.Printf("📈 Replicas: %d\n", status.Replicas)
	}

	if len(status.Ports) > 0 {
		fmt.Printf("🔌 Ports: %v\n", status.Ports)
	}

	if len(service.Domains) > 0 {
		fmt.Printf("🌐 Domains:\n")
		for server, domain := range service.Domains {
			fmt.Printf("   %s: %s\n", server, domain)
		}
	}

	return nil
}

// showAllServiceStatus shows status for all services in current nexus
func showAllServiceStatus() error {
	current, err := nexusManager.GetCurrent()
	if err != nil {
		return fmt.Errorf("no current nexus: %w", err)
	}

	config := configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	fmt.Printf("📊 Service Status for nexus '%s'\n", current.Name)
	fmt.Println("═══════════════════════════════════════")

	// Find services that deploy to servers in current nexus
	nexusServers := make(map[string]bool)
	for _, server := range current.Config.Servers {
		nexusServers[server] = true
	}

	for serviceName, service := range config.Services {
		// Check if service deploys to any server in current nexus
		deployedHere := false
		for _, serverName := range service.Servers {
			if nexusServers[serverName] {
				deployedHere = true
				break
			}
		}

		if deployedHere {
			fmt.Printf("🔸 %s\n", serviceName)
			err := showServiceStatus(serviceName)
			if err != nil {
				fmt.Printf("   ❌ Error: %v\n", err)
			}
			fmt.Println()
		}
	}

	return nil
}

// showServiceLogs shows logs for a service
func showServiceLogs(serviceName string, follow bool) error {
	config := configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	service := config.Services[serviceName]
	if service == nil {
		return fmt.Errorf("service '%s' not found in configuration", serviceName)
	}

	// Create server instances
	servers := make(map[string]pkg.Server)
	factory := server.NewFactory()

	for _, serverName := range service.Servers {
		serverConfig := config.Servers[serverName]
		if serverConfig == nil {
			continue
		}

		srv, err := factory.CreateServer(serverName, serverConfig)
		if err != nil {
			continue
		}
		defer srv.Disconnect()

		servers[serverName] = srv
	}

	if len(servers) == 0 {
		return fmt.Errorf("no accessible servers found for service '%s'", serviceName)
	}

	// Create Docker provider and get logs
	dockerProvider := docker.NewProvider(servers, config)
	logChan, err := dockerProvider.Logs(serviceName, follow)
	if err != nil {
		return fmt.Errorf("failed to get service logs: %w", err)
	}

	fmt.Printf("📋 Logs for service '%s':\n", serviceName)
	fmt.Println("─────────────────────────────────────")

	// Stream logs
	for logLine := range logChan {
		fmt.Println(logLine)
	}

	return nil
}