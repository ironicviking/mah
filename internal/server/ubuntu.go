package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/jonas-jonas/mah/pkg"
)

// UbuntuOperations provides Ubuntu-specific server operations
type UbuntuOperations struct {
	server pkg.Server
}

// NewUbuntuOperations creates Ubuntu operations for a server
func NewUbuntuOperations(server pkg.Server) *UbuntuOperations {
	return &UbuntuOperations{server: server}
}

// InstallDocker installs Docker CE on Ubuntu
func (u *UbuntuOperations) InstallDocker(ctx context.Context) error {
	// Check if Docker is already installed
	result, err := u.server.Execute(ctx, "which docker", false)
	if err == nil && result.ExitCode == 0 {
		// Docker is installed, check version
		result, err = u.server.Execute(ctx, "docker --version", false)
		if err == nil && result.ExitCode == 0 {
			return nil // Docker is already installed and working
		}
	}

	// Update package index
	result, err = u.server.Execute(ctx, "apt-get update", true)
	if err != nil {
		return fmt.Errorf("failed to update package index: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("apt-get update failed: %s", result.Stderr)
	}

	// Install required packages
	packages := "apt-transport-https ca-certificates curl gnupg lsb-release"
	result, err = u.server.Execute(ctx, fmt.Sprintf("apt-get install -y %s", packages), true)
	if err != nil {
		return fmt.Errorf("failed to install prerequisites: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("prerequisite installation failed: %s", result.Stderr)
	}

	// Add Docker's official GPG key (broken into steps to handle sudo properly)
	// First, download the GPG key
	result, err = u.server.Execute(ctx, "curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /tmp/docker.gpg", false)
	if err != nil {
		return fmt.Errorf("failed to download Docker GPG key: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker GPG key download failed: %s", result.Stderr)
	}
	
	// Then, install it with proper sudo permissions
	result, err = u.server.Execute(ctx, "gpg --dearmor < /tmp/docker.gpg | sudo tee /usr/share/keyrings/docker-archive-keyring.gpg > /dev/null", false)
	if err != nil {
		return fmt.Errorf("failed to install Docker GPG key: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker GPG key installation failed: %s", result.Stderr)
	}
	
	// Clean up temporary file
	u.server.Execute(ctx, "rm -f /tmp/docker.gpg", false)

	// Add Docker repository
	cmd := `echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`
	result, err = u.server.Execute(ctx, cmd, false)
	if err != nil {
		return fmt.Errorf("failed to add Docker repository: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker repository setup failed: %s", result.Stderr)
	}

	// Update package index again
	result, err = u.server.Execute(ctx, "apt-get update", true)
	if err != nil {
		return fmt.Errorf("failed to update package index after adding Docker repo: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("apt-get update failed: %s", result.Stderr)
	}

	// Install Docker CE
	result, err = u.server.Execute(ctx, "apt-get install -y docker-ce docker-ce-cli containerd.io", true)
	if err != nil {
		return fmt.Errorf("failed to install Docker: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker installation failed: %s", result.Stderr)
	}

	// Start and enable Docker service
	result, err = u.server.Execute(ctx, "systemctl start docker", true)
	if err != nil {
		return fmt.Errorf("failed to start Docker service: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker service start failed: %s", result.Stderr)
	}

	result, err = u.server.Execute(ctx, "systemctl enable docker", true)
	if err != nil {
		return fmt.Errorf("failed to enable Docker service: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker service enable failed: %s", result.Stderr)
	}

	// Add user to docker group (if not root)
	result, err = u.server.Execute(ctx, "whoami", false)
	if err == nil && result.ExitCode == 0 {
		user := strings.TrimSpace(result.Stdout)
		if user != "root" {
			result, err = u.server.Execute(ctx, fmt.Sprintf("usermod -aG docker %s", user), true)
			if err == nil && result.ExitCode == 0 {
				// Note: User will need to log out and back in for group changes to take effect
			}
		}
	}

	// Verify Docker installation
	result, err = u.server.Execute(ctx, "docker --version", true)
	if err != nil {
		return fmt.Errorf("failed to verify Docker installation: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker verification failed: %s", result.Stderr)
	}

	return nil
}

// ConfigureFirewall sets up UFW firewall on Ubuntu
func (u *UbuntuOperations) ConfigureFirewall(ctx context.Context, rules []pkg.FirewallRule) error {
	// Install UFW if not present
	result, err := u.server.Execute(ctx, "which ufw", false)
	if err != nil || result.ExitCode != 0 {
		result, err = u.server.Execute(ctx, "apt-get update && apt-get install -y ufw", true)
		if err != nil {
			return fmt.Errorf("failed to install UFW: %w", err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("UFW installation failed: %s", result.Stderr)
		}
	}

	// Reset UFW to default state
	result, err = u.server.Execute(ctx, "ufw --force reset", true)
	if err != nil {
		return fmt.Errorf("failed to reset UFW: %w", err)
	}

	// Set default policies
	result, err = u.server.Execute(ctx, "ufw default deny incoming", true)
	if err != nil {
		return fmt.Errorf("failed to set default deny incoming: %w", err)
	}

	result, err = u.server.Execute(ctx, "ufw default allow outgoing", true)
	if err != nil {
		return fmt.Errorf("failed to set default allow outgoing: %w", err)
	}

	// Add firewall rules
	for _, rule := range rules {
		err = u.addFirewallRule(ctx, rule)
		if err != nil {
			return fmt.Errorf("failed to add firewall rule for port %d: %w", rule.Port, err)
		}
	}

	// Enable UFW
	result, err = u.server.Execute(ctx, "ufw --force enable", true)
	if err != nil {
		return fmt.Errorf("failed to enable UFW: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("UFW enable failed: %s", result.Stderr)
	}

	return nil
}

// addFirewallRule adds a single firewall rule
func (u *UbuntuOperations) addFirewallRule(ctx context.Context, rule pkg.FirewallRule) error {
	var cmd string

	// Build UFW command based on rule
	action := "allow"
	if rule.Action == "deny" {
		action = "deny"
	}

	protocol := rule.Protocol
	if protocol == "" {
		protocol = "tcp"
	}

	if rule.Source == "any" || rule.Source == "" {
		cmd = fmt.Sprintf("ufw %s %d/%s", action, rule.Port, protocol)
	} else {
		cmd = fmt.Sprintf("ufw %s from %s to any port %d proto %s", action, rule.Source, rule.Port, protocol)
	}

	result, err := u.server.Execute(ctx, cmd, true)
	if err != nil {
		return fmt.Errorf("failed to execute UFW command: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("UFW command failed: %s", result.Stderr)
	}

	return nil
}

// UpdateSystem updates Ubuntu system packages
func (u *UbuntuOperations) UpdateSystem(ctx context.Context) error {
	// Update package lists
	result, err := u.server.Execute(ctx, "apt-get update", true)
	if err != nil {
		return fmt.Errorf("failed to update package lists: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("package list update failed: %s", result.Stderr)
	}

	// Upgrade packages
	result, err = u.server.Execute(ctx, "DEBIAN_FRONTEND=noninteractive apt-get upgrade -y", true)
	if err != nil {
		return fmt.Errorf("failed to upgrade packages: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("package upgrade failed: %s", result.Stderr)
	}

	return nil
}

// ConfigureAutomaticUpdates sets up unattended upgrades
func (u *UbuntuOperations) ConfigureAutomaticUpdates(ctx context.Context) error {
	// Install unattended-upgrades
	result, err := u.server.Execute(ctx, "apt-get install -y unattended-upgrades", true)
	if err != nil {
		return fmt.Errorf("failed to install unattended-upgrades: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("unattended-upgrades installation failed: %s", result.Stderr)
	}

	// Enable automatic updates
	result, err = u.server.Execute(ctx, "dpkg-reconfigure -plow unattended-upgrades", true)
	if err != nil {
		return fmt.Errorf("failed to configure unattended-upgrades: %w", err)
	}

	// Enable and start the service
	result, err = u.server.Execute(ctx, "systemctl enable unattended-upgrades", true)
	if err != nil {
		return fmt.Errorf("failed to enable unattended-upgrades service: %w", err)
	}

	result, err = u.server.Execute(ctx, "systemctl start unattended-upgrades", true)
	if err != nil {
		return fmt.Errorf("failed to start unattended-upgrades service: %w", err)
	}

	return nil
}

// HardenSSH applies SSH security hardening
func (u *UbuntuOperations) HardenSSH(ctx context.Context) error {
	// Create backup of SSH config
	result, err := u.server.Execute(ctx, "cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup", true)
	if err != nil {
		return fmt.Errorf("failed to backup SSH config: %w", err)
	}

	hardening := []string{
		"sed -i 's/#PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config",
		"sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config",
		"sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config",
		"sed -i 's/#Protocol 2/Protocol 2/' /etc/ssh/sshd_config",
		"sed -i 's/#X11Forwarding yes/X11Forwarding no/' /etc/ssh/sshd_config",
	}

	for _, cmd := range hardening {
		result, err = u.server.Execute(ctx, cmd, true)
		if err != nil {
			return fmt.Errorf("failed to apply SSH hardening: %w", err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("SSH hardening command failed: %s", result.Stderr)
		}
	}

	// Test SSH configuration
	result, err = u.server.Execute(ctx, "sshd -t", true)
	if err != nil {
		return fmt.Errorf("failed to test SSH configuration: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("SSH configuration test failed: %s", result.Stderr)
	}

	// Restart SSH service
	result, err = u.server.Execute(ctx, "systemctl restart sshd", true)
	if err != nil {
		return fmt.Errorf("failed to restart SSH service: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("SSH service restart failed: %s", result.Stderr)
	}

	return nil
}

// InstallPackage installs a package using apt
func (u *UbuntuOperations) InstallPackage(ctx context.Context, packageName string) error {
	result, err := u.server.Execute(ctx, fmt.Sprintf("apt-get install -y %s", packageName), true)
	if err != nil {
		return fmt.Errorf("failed to install package %s: %w", packageName, err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("package installation failed: %s", result.Stderr)
	}
	return nil
}

// GetDockerStatus returns the status of Docker service
func (u *UbuntuOperations) GetDockerStatus(ctx context.Context) (string, error) {
	result, err := u.server.Execute(ctx, "systemctl is-active docker", false)
	if err != nil {
		return "unknown", err
	}
	
	status := strings.TrimSpace(result.Stdout)
	if result.ExitCode != 0 {
		if status == "" {
			status = "inactive"
		}
	}
	
	return status, nil
}