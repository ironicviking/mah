package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/jonas-jonas/mah/internal/config"
	"github.com/jonas-jonas/mah/pkg"
)

// ServerFactory creates server instances based on configuration
type ServerFactory struct{}

// NewFactory creates a new server factory
func NewFactory() *ServerFactory {
	return &ServerFactory{}
}

// CreateServer creates a server instance from configuration
func (f *ServerFactory) CreateServer(id string, config *config.Server) (pkg.Server, error) {
	if config == nil {
		return nil, fmt.Errorf("server configuration is nil")
	}

	if config.Host == "" {
		return nil, fmt.Errorf("server host is required")
	}

	if config.SSHUser == "" {
		return nil, fmt.Errorf("SSH user is required")
	}

	if config.SSHKey == "" {
		return nil, fmt.Errorf("SSH key is required")
	}

	// Create base SSH server
	server := NewSSHServer(id, config)
	
	return server, nil
}

// CreateServerWithDistroDetection creates a server and detects its distribution
func (f *ServerFactory) CreateServerWithDistroDetection(ctx context.Context, id string, config *config.Server) (pkg.Server, error) {
	server, err := f.CreateServer(id, config)
	if err != nil {
		return nil, err
	}

	// Connect to detect distribution if not specified
	if config.Distro == "" {
		err = server.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect for distro detection: %w", err)
		}
		defer server.Disconnect()

		distro, err := server.GetDistro(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to detect distribution: %w", err)
		}

		config.Distro = distro
	}

	return server, nil
}

// ValidateServerConfig validates server configuration
func (f *ServerFactory) ValidateServerConfig(config *config.Server) error {
	if config == nil {
		return fmt.Errorf("server configuration is nil")
	}

	if config.Host == "" {
		return fmt.Errorf("host is required")
	}

	if config.SSHUser == "" {
		return fmt.Errorf("ssh_user is required")
	}

	if config.SSHKey == "" {
		return fmt.Errorf("ssh_key is required")
	}

	// Validate distribution if specified
	if config.Distro != "" {
		supportedDistros := []string{
			"ubuntu", "debian", "centos", "rhel", "rocky", "fedora", "alpine", "unknown",
		}
		
		distroSupported := false
		for _, supported := range supportedDistros {
			if strings.ToLower(config.Distro) == supported {
				distroSupported = true
				break
			}
		}

		if !distroSupported {
			return fmt.Errorf("unsupported distribution: %s (supported: %s)", 
				config.Distro, strings.Join(supportedDistros, ", "))
		}
	}

	// Validate SSH port
	if config.SSHPort < 0 || config.SSHPort > 65535 {
		return fmt.Errorf("invalid SSH port: %d (must be 1-65535 or 0 for default)", config.SSHPort)
	}

	return nil
}

// GetSupportedDistributions returns a list of supported Linux distributions
func (f *ServerFactory) GetSupportedDistributions() []string {
	return []string{
		"ubuntu",
		"debian", 
		"centos",
		"rhel",
		"rocky",
		"fedora",
		"alpine",
	}
}

// GetDistroInfo returns information about a specific distribution
func (f *ServerFactory) GetDistroInfo(distro string) (*DistroInfo, error) {
	distro = strings.ToLower(distro)
	
	switch distro {
	case "ubuntu":
		return &DistroInfo{
			Name:           "Ubuntu",
			Family:         "debian",
			PackageManager: "apt",
			FirewallTool:   "ufw",
			ServiceManager: "systemctl",
			InitSystem:     "systemd",
		}, nil
	case "debian":
		return &DistroInfo{
			Name:           "Debian",
			Family:         "debian",
			PackageManager: "apt",
			FirewallTool:   "ufw",
			ServiceManager: "systemctl",
			InitSystem:     "systemd",
		}, nil
	case "centos", "rhel":
		return &DistroInfo{
			Name:           "CentOS/RHEL",
			Family:         "rhel",
			PackageManager: "yum",
			FirewallTool:   "firewalld",
			ServiceManager: "systemctl",
			InitSystem:     "systemd",
		}, nil
	case "rocky":
		return &DistroInfo{
			Name:           "Rocky Linux",
			Family:         "rhel",
			PackageManager: "dnf",
			FirewallTool:   "firewalld",
			ServiceManager: "systemctl",
			InitSystem:     "systemd",
		}, nil
	case "fedora":
		return &DistroInfo{
			Name:           "Fedora",
			Family:         "rhel",
			PackageManager: "dnf",
			FirewallTool:   "firewalld",
			ServiceManager: "systemctl",
			InitSystem:     "systemd",
		}, nil
	case "alpine":
		return &DistroInfo{
			Name:           "Alpine Linux",
			Family:         "alpine",
			PackageManager: "apk",
			FirewallTool:   "iptables",
			ServiceManager: "rc-service",
			InitSystem:     "openrc",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported distribution: %s", distro)
	}
}

// DistroInfo contains information about a Linux distribution
type DistroInfo struct {
	Name           string
	Family         string
	PackageManager string
	FirewallTool   string
	ServiceManager string
	InitSystem     string
}