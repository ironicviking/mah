package docker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jonas-jonas/mah/internal/config"
	"github.com/jonas-jonas/mah/pkg"
)

// Provider implements the ContainerProvider interface for Docker
type Provider struct {
	servers map[string]pkg.Server
	config  *config.Config
}

// NewProvider creates a new Docker provider
func NewProvider(servers map[string]pkg.Server, config *config.Config) *Provider {
	return &Provider{
		servers: servers,
		config:  config,
	}
}

// Deploy deploys a service using Docker Compose
func (p *Provider) Deploy(serviceConfig *pkg.ServiceConfig) error {
	ctx := context.Background()

	// Generate docker-compose file
	composeContent, err := p.generateComposeFile(serviceConfig)
	if err != nil {
		return fmt.Errorf("failed to generate docker-compose file: %w", err)
	}

	// Deploy to each specified server
	for _, serverName := range serviceConfig.Servers {
		server, exists := p.servers[serverName]
		if !exists {
			return fmt.Errorf("server '%s' not found", serverName)
		}

		fmt.Printf("🚀 Deploying service '%s' to server '%s'...\n", serviceConfig.Name, serverName)

		err := p.deployToServer(ctx, server, serviceConfig, composeContent)
		if err != nil {
			return fmt.Errorf("failed to deploy to server '%s': %w", serverName, err)
		}

		fmt.Printf("✅ Service '%s' deployed successfully to '%s'\n", serviceConfig.Name, serverName)
	}

	return nil
}

// Scale scales a service to the specified number of replicas
func (p *Provider) Scale(serviceName string, replicas int) error {
	ctx := context.Background()

	// Find service configuration
	service := p.config.Services[serviceName]
	if service == nil {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	// Scale on each server
	for _, serverName := range service.Servers {
		server, exists := p.servers[serverName]
		if !exists {
			continue
		}

		fmt.Printf("📈 Scaling service '%s' to %d replicas on server '%s'...\n", serviceName, replicas, serverName)

		// Use docker-compose scale command
		cmd := fmt.Sprintf("sh -c 'cd /opt/mah/services/%s && docker-compose up -d --scale %s=%d'", 
			serviceName, serviceName, replicas)
		
		result, err := server.Execute(ctx, cmd, true)
		if err != nil {
			return fmt.Errorf("failed to scale service on server '%s': %w", serverName, err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("scale command failed on server '%s': %s", serverName, result.Stderr)
		}
	}

	return nil
}

// Status returns the status of a service
func (p *Provider) Status(serviceName string) (*pkg.ServiceStatus, error) {
	ctx := context.Background()

	// Find service configuration
	service := p.config.Services[serviceName]
	if service == nil {
		return nil, fmt.Errorf("service '%s' not found", serviceName)
	}

	// Get status from first server (could be enhanced to aggregate from all servers)
	if len(service.Servers) == 0 {
		return nil, fmt.Errorf("service '%s' has no servers configured", serviceName)
	}

	serverName := service.Servers[0]
	server, exists := p.servers[serverName]
	if !exists {
		return nil, fmt.Errorf("server '%s' not found", serverName)
	}

	// Get container status using docker-compose
	cmd := fmt.Sprintf("sh -c 'cd /opt/mah/services/%s && docker-compose ps --format json'", serviceName)
	result, err := server.Execute(ctx, cmd, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get service status: %w", err)
	}

	if result.ExitCode != 0 {
		return &pkg.ServiceStatus{
			Name:   serviceName,
			Status: "not_deployed",
			Health: "unknown",
		}, nil
	}

	// Parse the docker-compose output (simplified)
	status := "running"
	health := "healthy"
	if strings.Contains(result.Stdout, "exited") {
		status = "stopped"
		health = "unhealthy"
	}

	return &pkg.ServiceStatus{
		Name:      serviceName,
		Status:    status,
		Replicas:  service.Replicas,
		Health:    health,
		Ports:     extractPortNumbers(service.Ports),
		Variables: service.Environment,
	}, nil
}

// Logs streams logs from a service
func (p *Provider) Logs(serviceName string, follow bool) (<-chan string, error) {
	ctx := context.Background()
	logChan := make(chan string, 100)

	// Find service configuration
	service := p.config.Services[serviceName]
	if service == nil {
		close(logChan)
		return logChan, fmt.Errorf("service '%s' not found", serviceName)
	}

	// Stream logs from first server
	if len(service.Servers) == 0 {
		close(logChan)
		return logChan, fmt.Errorf("service '%s' has no servers configured", serviceName)
	}

	serverName := service.Servers[0]
	server, exists := p.servers[serverName]
	if !exists {
		close(logChan)
		return logChan, fmt.Errorf("server '%s' not found", serverName)
	}

	// Start log streaming in a goroutine
	go func() {
		defer close(logChan)

		followFlag := ""
		if follow {
			followFlag = "-f"
		}

		cmd := fmt.Sprintf("sh -c 'cd /opt/mah/services/%s && docker-compose logs %s --tail=50'", serviceName, followFlag)
		result, err := server.Execute(ctx, cmd, false)
		if err != nil {
			logChan <- fmt.Sprintf("Error getting logs: %v", err)
			return
		}

		if result.ExitCode != 0 {
			logChan <- fmt.Sprintf("Log command failed: %s", result.Stderr)
			return
		}

		// Send logs line by line
		lines := strings.Split(result.Stdout, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				select {
				case logChan <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return logChan, nil
}

// Remove removes a service
func (p *Provider) Remove(serviceName string) error {
	ctx := context.Background()

	// Find service configuration
	service := p.config.Services[serviceName]
	if service == nil {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	// Remove from each server
	for _, serverName := range service.Servers {
		server, exists := p.servers[serverName]
		if !exists {
			continue
		}

		fmt.Printf("🗑️  Removing service '%s' from server '%s'...\n", serviceName, serverName)

		// Stop and remove containers
		cmd := fmt.Sprintf("sh -c 'cd /opt/mah/services/%s && docker-compose down -v'", serviceName)
		result, err := server.Execute(ctx, cmd, true)
		if err != nil {
			return fmt.Errorf("failed to remove service from server '%s': %w", serverName, err)
		}
		if result.ExitCode != 0 {
			fmt.Printf("⚠️  Warning: removal command had issues on server '%s': %s\n", serverName, result.Stderr)
		}

		// Remove service directory
		cmd = fmt.Sprintf("rm -rf /opt/mah/services/%s", serviceName)
		result, err = server.Execute(ctx, cmd, true)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to remove service directory on server '%s': %v\n", serverName, err)
		}
	}

	return nil
}

// deployToServer deploys a service to a specific server
func (p *Provider) deployToServer(ctx context.Context, server pkg.Server, serviceConfig *pkg.ServiceConfig, composeContent string) error {
	// Ensure server is connected
	err := server.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	// Create service directory
	serviceDir := fmt.Sprintf("/opt/mah/services/%s", serviceConfig.Name)
	result, err := server.Execute(ctx, fmt.Sprintf("mkdir -p %s", serviceDir), true)
	if err != nil {
		return fmt.Errorf("failed to create service directory: %w", err)
	}

	// Write docker-compose.yml file using sudo tee to handle permissions properly
	composeFile := fmt.Sprintf("%s/docker-compose.yml", serviceDir)
	cmd := fmt.Sprintf("cat << 'EOF' | sudo tee %s > /dev/null\n%s\nEOF", composeFile, composeContent)
	result, err = server.Execute(ctx, cmd, false)
	if err != nil {
		return fmt.Errorf("failed to write docker-compose file: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to write docker-compose file: %s", result.Stderr)
	}

	// Create .env file if service has environment variables
	if len(serviceConfig.Environment) > 0 {
		envContent := ""
		for key, value := range serviceConfig.Environment {
			envContent += fmt.Sprintf("%s=%s\n", key, value)
		}

		envFile := fmt.Sprintf("%s/.env", serviceDir)
		cmd = fmt.Sprintf("cat << 'EOF' | sudo tee %s > /dev/null\n%s\nEOF", envFile, envContent)
		result, err = server.Execute(ctx, cmd, false)
		if err != nil {
			return fmt.Errorf("failed to write .env file: %w", err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("failed to write .env file: %s", result.Stderr)
		}
	}

	// Pull images
	cmd = fmt.Sprintf("sh -c 'cd %s && docker-compose pull'", serviceDir)
	result, err = server.Execute(ctx, cmd, true)
	if err != nil {
		return fmt.Errorf("failed to pull Docker images: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to pull Docker images: %s", result.Stderr)
	}

	// Deploy service
	cmd = fmt.Sprintf("sh -c 'cd %s && docker-compose up -d'", serviceDir)
	result, err = server.Execute(ctx, cmd, true)
	if err != nil {
		return fmt.Errorf("failed to deploy service: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to deploy service: %s", result.Stderr)
	}

	return nil
}

// generateComposeFile generates a docker-compose.yml file for the service
func (p *Provider) generateComposeFile(serviceConfig *pkg.ServiceConfig) (string, error) {
	compose := ComposeFile{
		Version:  "3.8",
		Services: make(map[string]ComposeService),
	}

	// Create main service
	service := ComposeService{
		Image:       serviceConfig.Image,
		Environment: serviceConfig.Environment,
		Volumes:     serviceConfig.Volumes,
		Networks:    serviceConfig.Networks,
		DependsOn:   serviceConfig.Depends,
		Labels:      serviceConfig.Labels,
	}

	// Add port mappings
	if serviceConfig.Public {
		for _, port := range serviceConfig.Ports {
			service.Ports = append(service.Ports, fmt.Sprintf("%d:%d", port, port))
		}
	}

	// Add restart policy
	service.Restart = "unless-stopped"

	compose.Services[serviceConfig.Name] = service

	// Generate networks if specified
	if len(serviceConfig.Networks) > 0 {
		compose.Networks = make(map[string]ComposeNetwork)
		for _, network := range serviceConfig.Networks {
			compose.Networks[network] = ComposeNetwork{
				External: true,
			}
		}
	}

	return compose.ToYAML(), nil
}

// extractPortNumbers converts Docker port mappings (e.g., "8080:8080", "53:53/tcp") to port numbers
func extractPortNumbers(portMappings []string) []int {
	var ports []int
	for _, mapping := range portMappings {
		// Remove protocol suffix if present (e.g., "/tcp", "/udp")
		mapping = strings.Split(mapping, "/")[0]
		
		// Split by colon to get host:container port mapping
		parts := strings.Split(mapping, ":")
		if len(parts) >= 1 {
			// Use the first part (host port) for the exposed port
			if port, err := strconv.Atoi(parts[0]); err == nil {
				ports = append(ports, port)
			}
		}
	}
	return ports
}