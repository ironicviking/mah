package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/jonas-jonas/mah/pkg"
)

// RockyOperations provides Rocky Linux-specific server operations
type RockyOperations struct {
	server pkg.Server
}

// NewRockyOperations creates Rocky Linux operations for a server
func NewRockyOperations(server pkg.Server) *RockyOperations {
	return &RockyOperations{server: server}
}

// InstallDocker installs Docker CE on Rocky Linux
func (r *RockyOperations) InstallDocker(ctx context.Context) error {
	// Check if Docker is already installed
	result, err := r.server.Execute(ctx, "which docker", false)
	if err == nil && result.ExitCode == 0 {
		// Docker is installed, check version
		result, err = r.server.Execute(ctx, "docker --version", false)
		if err == nil && result.ExitCode == 0 {
			return nil // Docker is already installed and working
		}
	}

	// Install required packages
	packages := "yum-utils device-mapper-persistent-data lvm2"
	result, err = r.server.Execute(ctx, fmt.Sprintf("dnf install -y %s", packages), true)
	if err != nil {
		return fmt.Errorf("failed to install prerequisites: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("prerequisite installation failed: %s", result.Stderr)
	}

	// Add Docker CE repository
	result, err = r.server.Execute(ctx, "dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo", true)
	if err != nil {
		return fmt.Errorf("failed to add Docker repository: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker repository setup failed: %s", result.Stderr)
	}

	// Install Docker CE
	result, err = r.server.Execute(ctx, "dnf install -y docker-ce docker-ce-cli containerd.io", true)
	if err != nil {
		return fmt.Errorf("failed to install Docker: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker installation failed: %s", result.Stderr)
	}

	// Start and enable Docker service
	result, err = r.server.Execute(ctx, "systemctl start docker", true)
	if err != nil {
		return fmt.Errorf("failed to start Docker service: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker service start failed: %s", result.Stderr)
	}

	result, err = r.server.Execute(ctx, "systemctl enable docker", true)
	if err != nil {
		return fmt.Errorf("failed to enable Docker service: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker service enable failed: %s", result.Stderr)
	}

	// Add user to docker group (if not root)
	result, err = r.server.Execute(ctx, "whoami", false)
	if err == nil && result.ExitCode == 0 {
		user := strings.TrimSpace(result.Stdout)
		if user != "root" {
			result, err = r.server.Execute(ctx, fmt.Sprintf("usermod -aG docker %s", user), true)
			if err == nil && result.ExitCode == 0 {
				// Note: User will need to log out and back in for group changes to take effect
			}
		}
	}

	// Verify Docker installation
	result, err = r.server.Execute(ctx, "docker --version", true)
	if err != nil {
		return fmt.Errorf("failed to verify Docker installation: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Docker verification failed: %s", result.Stderr)
	}

	return nil
}

// ConfigureFirewall sets up firewalld on Rocky Linux
func (r *RockyOperations) ConfigureFirewall(ctx context.Context, rules []pkg.FirewallRule) error {
	// Ensure firewalld is installed
	result, err := r.server.Execute(ctx, "which firewall-cmd", false)
	if err != nil || result.ExitCode != 0 {
		result, err = r.server.Execute(ctx, "dnf install -y firewalld", true)
		if err != nil {
			return fmt.Errorf("failed to install firewalld: %w", err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("firewalld installation failed: %s", result.Stderr)
		}
	}

	// Start and enable firewalld
	result, err = r.server.Execute(ctx, "systemctl start firewalld", true)
	if err != nil {
		return fmt.Errorf("failed to start firewalld: %w", err)
	}

	result, err = r.server.Execute(ctx, "systemctl enable firewalld", true)
	if err != nil {
		return fmt.Errorf("failed to enable firewalld: %w", err)
	}

	// Set default zone to public
	result, err = r.server.Execute(ctx, "firewall-cmd --set-default-zone=public", true)
	if err != nil {
		return fmt.Errorf("failed to set default zone: %w", err)
	}

	// Clear existing rules by reloading defaults
	result, err = r.server.Execute(ctx, "firewall-cmd --complete-reload", true)
	if err != nil {
		return fmt.Errorf("failed to reload firewall: %w", err)
	}

	// Add firewall rules
	for _, rule := range rules {
		err = r.addFirewallRule(ctx, rule)
		if err != nil {
			return fmt.Errorf("failed to add firewall rule for port %d: %w", rule.Port, err)
		}
	}

	// Make rules permanent and reload
	result, err = r.server.Execute(ctx, "firewall-cmd --runtime-to-permanent", true)
	if err != nil {
		return fmt.Errorf("failed to make firewall rules permanent: %w", err)
	}

	result, err = r.server.Execute(ctx, "firewall-cmd --reload", true)
	if err != nil {
		return fmt.Errorf("failed to reload firewall: %w", err)
	}

	return nil
}

// addFirewallRule adds a single firewall rule using firewalld
func (r *RockyOperations) addFirewallRule(ctx context.Context, rule pkg.FirewallRule) error {
	var cmd string

	protocol := rule.Protocol
	if protocol == "" {
		protocol = "tcp"
	}

	// Handle tcp/udp protocol
	if protocol == "tcp/udp" {
		// Add both TCP and UDP rules
		tcpRule := rule
		tcpRule.Protocol = "tcp"
		err := r.addFirewallRule(ctx, tcpRule)
		if err != nil {
			return err
		}

		udpRule := rule
		udpRule.Protocol = "udp"
		return r.addFirewallRule(ctx, udpRule)
	}

	if rule.Source == "any" || rule.Source == "" {
		// Open port for all sources
		cmd = fmt.Sprintf("firewall-cmd --add-port=%d/%s", rule.Port, protocol)
	} else {
		// Add rich rule for specific source
		if rule.Action == "deny" {
			cmd = fmt.Sprintf("firewall-cmd --add-rich-rule='rule family=\"ipv4\" source address=\"%s\" port protocol=\"%s\" port=\"%d\" reject'", 
				rule.Source, protocol, rule.Port)
		} else {
			cmd = fmt.Sprintf("firewall-cmd --add-rich-rule='rule family=\"ipv4\" source address=\"%s\" port protocol=\"%s\" port=\"%d\" accept'", 
				rule.Source, protocol, rule.Port)
		}
	}

	result, err := r.server.Execute(ctx, cmd, true)
	if err != nil {
		return fmt.Errorf("failed to execute firewalld command: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("firewalld command failed: %s", result.Stderr)
	}

	return nil
}

// UpdateSystem updates Rocky Linux system packages
func (r *RockyOperations) UpdateSystem(ctx context.Context) error {
	// Update packages
	result, err := r.server.Execute(ctx, "dnf upgrade -y", true)
	if err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("package update failed: %s", result.Stderr)
	}

	return nil
}

// ConfigureAutomaticUpdates sets up automatic updates with dnf-automatic
func (r *RockyOperations) ConfigureAutomaticUpdates(ctx context.Context) error {
	// Install dnf-automatic
	result, err := r.server.Execute(ctx, "dnf install -y dnf-automatic", true)
	if err != nil {
		return fmt.Errorf("failed to install dnf-automatic: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("dnf-automatic installation failed: %s", result.Stderr)
	}

	// Configure automatic updates (security updates only)
	config := `[commands]
upgrade_type = security
random_sleep = 0

[emitters]
emit_via = stdio

[email]
email_from = root@localhost
email_to = root

[base]
debuglevel = 1`

	// Write configuration using sudo tee to handle permissions properly
	result, err = r.server.Execute(ctx, fmt.Sprintf("cat << 'EOF' | sudo tee /etc/dnf/automatic.conf > /dev/null\n%s\nEOF", config), false)
	if err != nil {
		return fmt.Errorf("failed to configure dnf-automatic: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("dnf-automatic configuration failed: %s", result.Stderr)
	}

	// Enable and start the timer
	result, err = r.server.Execute(ctx, "systemctl enable dnf-automatic.timer", true)
	if err != nil {
		return fmt.Errorf("failed to enable dnf-automatic timer: %w", err)
	}

	result, err = r.server.Execute(ctx, "systemctl start dnf-automatic.timer", true)
	if err != nil {
		return fmt.Errorf("failed to start dnf-automatic timer: %w", err)
	}

	return nil
}

// HardenSSH applies SSH security hardening
func (r *RockyOperations) HardenSSH(ctx context.Context) error {
	// Create backup of SSH config
	result, err := r.server.Execute(ctx, "cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup", true)
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
		result, err = r.server.Execute(ctx, cmd, true)
		if err != nil {
			return fmt.Errorf("failed to apply SSH hardening: %w", err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("SSH hardening command failed: %s", result.Stderr)
		}
	}

	// Test SSH configuration
	result, err = r.server.Execute(ctx, "sshd -t", true)
	if err != nil {
		return fmt.Errorf("failed to test SSH configuration: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("SSH configuration test failed: %s", result.Stderr)
	}

	// Restart SSH service
	result, err = r.server.Execute(ctx, "systemctl restart sshd", true)
	if err != nil {
		return fmt.Errorf("failed to restart SSH service: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("SSH service restart failed: %s", result.Stderr)
	}

	return nil
}

// InstallPackage installs a package using dnf
func (r *RockyOperations) InstallPackage(ctx context.Context, packageName string) error {
	result, err := r.server.Execute(ctx, fmt.Sprintf("dnf install -y %s", packageName), true)
	if err != nil {
		return fmt.Errorf("failed to install package %s: %w", packageName, err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("package installation failed: %s", result.Stderr)
	}
	return nil
}

// GetDockerStatus returns the status of Docker service
func (r *RockyOperations) GetDockerStatus(ctx context.Context) (string, error) {
	result, err := r.server.Execute(ctx, "systemctl is-active docker", false)
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