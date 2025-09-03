package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/jonas-jonas/mah/internal/config"
	"github.com/jonas-jonas/mah/internal/server"
	"github.com/jonas-jonas/mah/pkg"
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
		if len(args) == 0 {
			return showAllServerStatus()
		}
		return showServerStatus(args[0])
	},
}

var serverInitCmd = &cobra.Command{
	Use:   "init <server-name>",
	Short: "Initialize a server (install Docker, configure firewall, etc.)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return initializeServer(args[0])
	},
}

func init() {
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverStatusCmd)
	serverCmd.AddCommand(serverInitCmd)
}

// initializeServer initializes a server with Docker, firewall, and security hardening
func initializeServer(serverName string) error {
	config := configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	serverConfig := config.Servers[serverName]
	if serverConfig == nil {
		return fmt.Errorf("server '%s' not found in configuration", serverName)
	}

	fmt.Printf("🚀 Initializing server '%s' (%s)...\n", serverName, serverConfig.Host)

	// Create server instance
	factory := server.NewFactory()
	ctx := context.Background()
	
	srv, err := factory.CreateServerWithDistroDetection(ctx, serverName, serverConfig)
	if err != nil {
		return fmt.Errorf("failed to create server instance: %w", err)
	}
	defer srv.Disconnect()

	fmt.Printf("📡 Detected distribution: %s\n", serverConfig.Distro)

	// Perform health check
	fmt.Print("🔍 Performing health check... ")
	err = srv.HealthCheck(ctx)
	if err != nil {
		color.Red("FAILED")
		return fmt.Errorf("health check failed: %w", err)
	}
	color.Green("OK")

	// Update system packages
	fmt.Print("📦 Updating system packages... ")
	err = updateSystemPackages(ctx, srv, serverConfig.Distro)
	if err != nil {
		color.Red("FAILED")
		return fmt.Errorf("failed to update system packages: %w", err)
	}
	color.Green("OK")

	// Install Docker
	fmt.Print("🐳 Installing Docker... ")
	err = installDocker(ctx, srv, serverConfig.Distro)
	if err != nil {
		color.Red("FAILED")
		return fmt.Errorf("failed to install Docker: %w", err)
	}
	color.Green("OK")

	// Configure firewall
	if config.Firewall != nil {
		fmt.Print("🔥 Configuring firewall... ")
		err = configureFirewall(ctx, srv, serverName, config, serverConfig.Distro)
		if err != nil {
			color.Red("FAILED")
			return fmt.Errorf("failed to configure firewall: %w", err)
		}
		color.Green("OK")
	}

	// Harden SSH
	fmt.Print("🔐 Hardening SSH configuration... ")
	err = hardenSSH(ctx, srv, serverConfig.Distro)
	if err != nil {
		color.Yellow("WARNING")
		fmt.Printf("   SSH hardening failed (continuing): %v\n", err)
	} else {
		color.Green("OK")
	}

	// Configure automatic updates
	fmt.Print("🔄 Configuring automatic updates... ")
	err = configureAutomaticUpdates(ctx, srv, serverConfig.Distro)
	if err != nil {
		color.Yellow("WARNING")
		fmt.Printf("   Automatic updates failed (continuing): %v\n", err)
	} else {
		color.Green("OK")
	}

	color.Green("\n✅ Server initialization completed successfully!")
	fmt.Printf("Server '%s' is ready for service deployment.\n", serverName)

	return nil
}

// showServerStatus shows status for a specific server
func showServerStatus(serverName string) error {
	config := configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	serverConfig := config.Servers[serverName]
	if serverConfig == nil {
		return fmt.Errorf("server '%s' not found in configuration", serverName)
	}

	// Create server instance
	factory := server.NewFactory()
	ctx := context.Background()
	
	srv, err := factory.CreateServer(serverName, serverConfig)
	if err != nil {
		return fmt.Errorf("failed to create server instance: %w", err)
	}
	defer srv.Disconnect()

	fmt.Printf("📊 Server Status: %s (%s)\n", serverName, serverConfig.Host)
	fmt.Println("─────────────────────────────────────")

	// Test connectivity
	fmt.Print("🔗 Connectivity: ")
	err = srv.Connect(ctx)
	if err != nil {
		color.Red("DISCONNECTED")
		fmt.Printf("   Error: %v\n", err)
		return nil
	}
	color.Green("CONNECTED")

	// Health check
	fmt.Print("💓 Health: ")
	err = srv.HealthCheck(ctx)
	if err != nil {
		color.Red("UNHEALTHY")
		fmt.Printf("   Error: %v\n", err)
	} else {
		color.Green("HEALTHY")
	}

	// Get system resources
	fmt.Print("📈 Resources: ")
	resources, err := srv.GetResources(ctx)
	if err != nil {
		color.Red("UNAVAILABLE")
		fmt.Printf("   Error: %v\n", err)
	} else {
		color.Green("AVAILABLE")
		fmt.Printf("   CPU: %d cores (%.1f%% used)\n", resources.CPU.Cores, resources.CPU.Usage)
		fmt.Printf("   Memory: %.1f GB (%.1f%% used)\n", 
			float64(resources.Memory.Total)/(1024*1024*1024), resources.Memory.Usage)
		fmt.Printf("   Disk: %.1f GB (%.1f%% used)\n", 
			float64(resources.Disk.Total)/(1024*1024*1024), resources.Disk.Usage)
		fmt.Printf("   Load: %.2f, %.2f, %.2f\n", 
			resources.Load.Load1, resources.Load.Load5, resources.Load.Load15)
	}

	// Check Docker status
	fmt.Print("🐳 Docker: ")
	dockerStatus, err := getDockerStatus(ctx, srv, serverConfig.Distro)
	if err != nil {
		color.Red("UNKNOWN")
		fmt.Printf("   Error: %v\n", err)
	} else {
		if dockerStatus == "active" {
			color.Green("ACTIVE")
		} else {
			color.Yellow(strings.ToUpper(dockerStatus))
		}
	}

	return nil
}

// showAllServerStatus shows status for all servers in current nexus
func showAllServerStatus() error {
	current, err := nexusManager.GetCurrent()
	if err != nil {
		return fmt.Errorf("no current nexus: %w", err)
	}

	config := configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	fmt.Printf("📊 Server Status for nexus '%s'\n", current.Name)
	fmt.Println("═══════════════════════════════════════")

	for _, serverName := range current.Config.Servers {
		if _, exists := config.Servers[serverName]; !exists {
			continue
		}

		err := showServerStatus(serverName)
		if err != nil {
			fmt.Printf("❌ %s: %v\n", serverName, err)
		}
		fmt.Println()
	}

	return nil
}

// Helper functions for different distributions

func updateSystemPackages(ctx context.Context, srv pkg.Server, distro string) error {
	switch strings.ToLower(distro) {
	case "ubuntu":
		ops := server.NewUbuntuOperations(srv)
		return ops.UpdateSystem(ctx)
	case "debian":
		ops := server.NewDebianOperations(srv)
		return ops.UpdateSystem(ctx)
	case "rocky", "rhel", "centos", "fedora":
		ops := server.NewRockyOperations(srv)
		return ops.UpdateSystem(ctx)
	default:
		return fmt.Errorf("unsupported distribution: %s", distro)
	}
}

func installDocker(ctx context.Context, srv pkg.Server, distro string) error {
	switch strings.ToLower(distro) {
	case "ubuntu":
		ops := server.NewUbuntuOperations(srv)
		return ops.InstallDocker(ctx)
	case "debian":
		ops := server.NewDebianOperations(srv)
		return ops.InstallDocker(ctx)
	case "rocky", "rhel", "centos", "fedora":
		ops := server.NewRockyOperations(srv)
		return ops.InstallDocker(ctx)
	default:
		return fmt.Errorf("unsupported distribution: %s", distro)
	}
}

func configureFirewall(ctx context.Context, srv pkg.Server, serverName string, config *config.Config, distro string) error {
	// Collect firewall rules
	var rules []pkg.FirewallRule

	// Add global rules
	for _, rule := range config.Firewall.Global {
		rules = append(rules, pkg.FirewallRule{
			Port:     rule.Port,
			Protocol: rule.Protocol,
			Source:   rule.From,
			Action:   "allow",
			Comment:  rule.Comment,
		})
	}

	// Add server-specific rules
	if serverRules, exists := config.Firewall.ServerSpecific[serverName]; exists {
		for _, rule := range serverRules {
			rules = append(rules, pkg.FirewallRule{
				Port:     rule.Port,
				Protocol: rule.Protocol,
				Source:   rule.From,
				Action:   "allow",
				Comment:  rule.Comment,
			})
		}
	}

	switch strings.ToLower(distro) {
	case "ubuntu":
		ops := server.NewUbuntuOperations(srv)
		return ops.ConfigureFirewall(ctx, rules)
	case "debian":
		ops := server.NewDebianOperations(srv)
		return ops.ConfigureFirewall(ctx, rules)
	case "rocky", "rhel", "centos", "fedora":
		ops := server.NewRockyOperations(srv)
		return ops.ConfigureFirewall(ctx, rules)
	default:
		return fmt.Errorf("unsupported distribution: %s", distro)
	}
}

func hardenSSH(ctx context.Context, srv pkg.Server, distro string) error {
	switch strings.ToLower(distro) {
	case "ubuntu":
		ops := server.NewUbuntuOperations(srv)
		return ops.HardenSSH(ctx)
	case "debian":
		ops := server.NewDebianOperations(srv)
		return ops.HardenSSH(ctx)
	case "rocky", "rhel", "centos", "fedora":
		ops := server.NewRockyOperations(srv)
		return ops.HardenSSH(ctx)
	default:
		return fmt.Errorf("unsupported distribution: %s", distro)
	}
}

func configureAutomaticUpdates(ctx context.Context, srv pkg.Server, distro string) error {
	switch strings.ToLower(distro) {
	case "ubuntu":
		ops := server.NewUbuntuOperations(srv)
		return ops.ConfigureAutomaticUpdates(ctx)
	case "debian":
		ops := server.NewDebianOperations(srv)
		return ops.ConfigureAutomaticUpdates(ctx)
	case "rocky", "rhel", "centos", "fedora":
		ops := server.NewRockyOperations(srv)
		return ops.ConfigureAutomaticUpdates(ctx)
	default:
		return fmt.Errorf("unsupported distribution: %s", distro)
	}
}

func getDockerStatus(ctx context.Context, srv pkg.Server, distro string) (string, error) {
	switch strings.ToLower(distro) {
	case "ubuntu":
		ops := server.NewUbuntuOperations(srv)
		return ops.GetDockerStatus(ctx)
	case "debian":
		ops := server.NewDebianOperations(srv)
		return ops.GetDockerStatus(ctx)
	case "rocky", "rhel", "centos", "fedora":
		ops := server.NewRockyOperations(srv)
		return ops.GetDockerStatus(ctx)
	default:
		return "unknown", fmt.Errorf("unsupported distribution: %s", distro)
	}
}