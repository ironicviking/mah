package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Configuration commands allow you to validate, show, and manage MAH configuration.",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := configManager.GetConfig()
		if config == nil {
			return fmt.Errorf("no configuration loaded")
		}

		data, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Printf("%s\n", string(data))
		return nil
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("configuration file not found: %s", configFile)
		}

		// Try to load configuration
		tempManager := configManager
		if err := tempManager.LoadConfig(configFile); err != nil {
			fmt.Printf("%s Configuration validation failed:\n", color.RedString("✗"))
			fmt.Printf("  %s\n", err.Error())
			return err
		}

		// Validate configuration
		if err := tempManager.ValidateConfig(); err != nil {
			fmt.Printf("%s Configuration validation failed:\n", color.RedString("✗"))
			fmt.Printf("  %s\n", err.Error())
			return err
		}

		fmt.Printf("%s Configuration is valid\n", color.GreenString("✓"))

		// Show summary
		config := tempManager.GetConfig()
		fmt.Printf("  Servers: %d\n", len(config.Servers))
		fmt.Printf("  Nexuses: %d\n", len(config.Nexuses))
		fmt.Printf("  Services: %d\n", len(config.Services))

		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a sample configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if config file already exists
		if _, err := os.Stat(configFile); err == nil {
			return fmt.Errorf("configuration file already exists: %s", configFile)
		}

		// Create sample configuration with environment variables
		sampleConfig := `# MAH Configuration File
version: "1.0"
project: "my-infrastructure"

# Server definitions
servers:
  thor:
    host: "${SERVER_HOST}"           # Set SERVER_HOST environment variable
    ssh_user: "${SSH_USER}"          # Set SSH_USER environment variable  
    ssh_key: "~/.ssh/id_rsa"
    sudo: true
    distro: "ubuntu"
    nexus: "production"

# Nexus definitions - logical groupings
nexuses:
  production:
    description: "Production environment"
    servers: ["thor"]
    environment: "production"

# Service definitions
services:
  blog:
    servers: ["thor"]
    image: "wordpress:latest"
    domains:
      thor: "${BLOG_DOMAIN}"         # Set BLOG_DOMAIN environment variable
    public: true
    environment:
      WORDPRESS_DB_HOST: "mysql"
      WORDPRESS_DB_PASSWORD: "${MYSQL_PASSWORD}"
    
  mysql:
    servers: ["thor"]
    image: "mysql:8"
    internal: true
    volumes:
      - "mysql_data:/var/lib/mysql"
    environment:
      MYSQL_ROOT_PASSWORD: "${MYSQL_PASSWORD}"

# Plugin configurations
plugins:
  dns:
    provider: "name.com"
    config:
      username: "${NAMECOM_USERNAME}"
      token: "${NAMECOM_TOKEN}"
      
  ssl:
    provider: "traefik"
    email: "${ADMIN_EMAIL}"          # Set ADMIN_EMAIL environment variable
    config:
      dns_challenge: true
      dns_provider: "name.com"

# Firewall rules
firewall:
  global:
    - port: 22
      protocol: tcp
      from: "any"
      comment: "SSH access"
    - port: 80
      protocol: tcp
      from: "any"
      comment: "HTTP traffic"
    - port: 443
      protocol: tcp
      from: "any"
      comment: "HTTPS traffic"
`

		if err := os.WriteFile(configFile, []byte(sampleConfig), 0644); err != nil {
			return fmt.Errorf("failed to create configuration file: %w", err)
		}

		fmt.Printf("%s Created sample configuration: %s\n", 
			color.GreenString("✓"), 
			color.CyanString(configFile))
		fmt.Println()
		fmt.Printf("%s %s\n", color.YellowString("⚠️"), "This config uses environment variables for sensitive data.")
		fmt.Println("Set these environment variables before using:")
		fmt.Println("  - SERVER_HOST=your.server.ip")
		fmt.Println("  - SSH_USER=your-ssh-username")  
		fmt.Println("  - BLOG_DOMAIN=blog.example.com")
		fmt.Println("  - ADMIN_EMAIL=admin@example.com")
		fmt.Println("  - MYSQL_PASSWORD=secure-password")
		fmt.Println("  - NAMECOM_USERNAME=your-namecom-username")
		fmt.Println("  - NAMECOM_TOKEN=your-namecom-token")
		fmt.Println()
		fmt.Printf("Or run: %s\n", color.CyanString("mah config secrets init"))

		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd) 
	configCmd.AddCommand(configInitCmd)
	// secretsCmd is added in secrets.go
}