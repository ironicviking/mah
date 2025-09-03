package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Manager handles configuration loading, validation, and management
type Manager struct {
	config        *Config
	runtime       *RuntimeConfig
	viper         *viper.Viper
	secrets       map[string]string
	secretManager *SecretManager
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		viper:   viper.New(),
		secrets: make(map[string]string),
	}
}

// LoadConfig loads configuration from file with environment variable substitution
func (m *Manager) LoadConfig(configPath string) error {
	// Set up viper
	m.viper.SetConfigFile(configPath)
	m.viper.SetConfigType("yaml")
	
	// Read environment variables
	m.loadEnvironmentVariables()
	
	// Read config file
	if err := m.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Unmarshal into struct
	var config Config
	if err := m.viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Substitute environment variables
	if err := m.substituteEnvVars(&config); err != nil {
		return fmt.Errorf("failed to substitute environment variables: %w", err)
	}
	
	// Validate configuration
	if err := m.validateConfig(&config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Validate secrets setup
	if m.secretManager != nil {
		if err := m.secretManager.ValidateSecretsSetup(&config); err != nil {
			// This is a warning, not a fatal error
			fmt.Printf("Warning: %v\n", err)
		}
	}
	
	m.config = &config
	return nil
}

// LoadRuntimeConfig loads runtime configuration and state
func (m *Manager) LoadRuntimeConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	mahDir := filepath.Join(homeDir, ".mah")
	
	// Ensure directories exist
	dirs := []string{
		mahDir,
		filepath.Join(mahDir, "state"),
		filepath.Join(mahDir, "cache"),
		filepath.Join(mahDir, "plugins"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// Initialize secret manager
	secretManager, err := NewSecretManager(mahDir)
	if err != nil {
		return fmt.Errorf("failed to create secret manager: %w", err)
	}
	m.secretManager = secretManager
	
	// Load secrets if available
	if secrets, err := secretManager.LoadSecrets(); err == nil {
		m.secrets = secrets
	}
	
	// Load or create runtime config
	runtimeConfigPath := filepath.Join(mahDir, "config.yaml")
	runtime := &RuntimeConfig{
		StateDir:   filepath.Join(mahDir, "state"),
		CacheDir:   filepath.Join(mahDir, "cache"),
		PluginDir:  filepath.Join(mahDir, "plugins"),
		LogLevel:   "info",
		ConfigFile: "mah.yaml",
	}
	
	// Try to load existing runtime config
	if data, err := os.ReadFile(runtimeConfigPath); err == nil {
		if err := yaml.Unmarshal(data, runtime); err != nil {
			return fmt.Errorf("failed to unmarshal runtime config: %w", err)
		}
	}
	
	m.runtime = runtime
	return m.saveRuntimeConfig()
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// GetRuntimeConfig returns the runtime configuration
func (m *Manager) GetRuntimeConfig() *RuntimeConfig {
	return m.runtime
}

// GetCurrentNexus returns the currently active nexus
func (m *Manager) GetCurrentNexus() string {
	return m.runtime.CurrentNexus
}

// SetCurrentNexus sets the currently active nexus
func (m *Manager) SetCurrentNexus(nexusName string) error {
	if m.config.Nexuses[nexusName] == nil {
		return fmt.Errorf("nexus '%s' not found in configuration", nexusName)
	}
	
	m.runtime.CurrentNexus = nexusName
	return m.saveRuntimeConfig()
}

// GetNexusServers returns servers for a given nexus
func (m *Manager) GetNexusServers(nexusName string) ([]*Server, error) {
	nexus := m.config.Nexuses[nexusName]
	if nexus == nil {
		return nil, fmt.Errorf("nexus '%s' not found", nexusName)
	}
	
	var servers []*Server
	for _, serverName := range nexus.Servers {
		server := m.config.Servers[serverName]
		if server == nil {
			return nil, fmt.Errorf("server '%s' referenced by nexus '%s' not found", serverName, nexusName)
		}
		servers = append(servers, server)
	}
	
	return servers, nil
}

// GetServicesByServer returns services configured for a specific server
func (m *Manager) GetServicesByServer(serverName string) []*Service {
	var services []*Service
	for _, service := range m.config.Services {
		for _, srvName := range service.Servers {
			if srvName == serverName {
				services = append(services, service)
				break
			}
		}
	}
	return services
}

// ValidateConfig performs comprehensive configuration validation
func (m *Manager) ValidateConfig() error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	return m.validateConfig(m.config)
}

// loadEnvironmentVariables loads all environment variables into viper
func (m *Manager) loadEnvironmentVariables() {
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			m.viper.Set("env."+parts[0], parts[1])
			m.secrets[parts[0]] = parts[1]
		}
	}
}

// substituteEnvVars performs environment variable substitution in config
func (m *Manager) substituteEnvVars(config *Config) error {
	// Regex to match ${VAR} patterns
	envRegex := regexp.MustCompile(`\$\{([^}]+)\}`)
	
	// Convert to YAML for processing
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config for env substitution: %w", err)
	}
	
	configStr := string(data)
	
	// Replace all environment variable references
	configStr = envRegex.ReplaceAllStringFunc(configStr, func(match string) string {
		varName := strings.Trim(match, "${}")
		if value, exists := m.secrets[varName]; exists {
			return value
		}
		if value := os.Getenv(varName); value != "" {
			return value
		}
		// Return original if not found
		return match
	})
	
	// Unmarshal back to struct
	return yaml.Unmarshal([]byte(configStr), config)
}

// validateConfig performs comprehensive configuration validation
func (m *Manager) validateConfig(config *Config) error {
	// Validate version
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}
	
	// Validate servers
	if len(config.Servers) == 0 {
		return fmt.Errorf("at least one server must be defined")
	}
	
	for name, server := range config.Servers {
		if server == nil {
			return fmt.Errorf("server '%s': configuration is nil", name)
		}
		if server.Host == "" {
			return fmt.Errorf("server '%s': host is required", name)
		}
		if server.SSHUser == "" {
			return fmt.Errorf("server '%s': ssh_user is required (got '%s')", name, server.SSHUser)
		}
		if server.SSHKey == "" {
			return fmt.Errorf("server '%s': ssh_key is required", name)
		}
		if server.Nexus == "" {
			return fmt.Errorf("server '%s': nexus is required", name)
		}
		
		// Expand SSH key path
		if strings.HasPrefix(server.SSHKey, "~/") {
			homeDir, _ := os.UserHomeDir()
			server.SSHKey = filepath.Join(homeDir, server.SSHKey[2:])
		}
		
		// Check if SSH key exists
		if _, err := os.Stat(server.SSHKey); os.IsNotExist(err) {
			return fmt.Errorf("server '%s': SSH key file not found: %s", name, server.SSHKey)
		}
		
		// Set defaults
		if server.SSHPort == 0 {
			server.SSHPort = 22
		}
	}
	
	// Validate nexuses
	if len(config.Nexuses) == 0 {
		return fmt.Errorf("at least one nexus must be defined")
	}
	
	for name, nexus := range config.Nexuses {
		if len(nexus.Servers) == 0 {
			return fmt.Errorf("nexus '%s': at least one server must be defined", name)
		}
		
		// Validate referenced servers exist
		for _, serverName := range nexus.Servers {
			if config.Servers[serverName] == nil {
				return fmt.Errorf("nexus '%s': references non-existent server '%s'", name, serverName)
			}
		}
	}
	
	// Validate services
	for name, service := range config.Services {
		if service.Image == "" && !service.Internal {
			return fmt.Errorf("service '%s': image is required for non-internal services", name)
		}
		
		if len(service.Servers) == 0 {
			return fmt.Errorf("service '%s': at least one server must be specified", name)
		}
		
		// Validate referenced servers exist
		for _, serverName := range service.Servers {
			if config.Servers[serverName] == nil {
				return fmt.Errorf("service '%s': references non-existent server '%s'", name, serverName)
			}
		}
		
		// Set defaults
		if service.Replicas == 0 {
			service.Replicas = 1
		}
	}
	
	// Validate firewall configuration
	if config.Firewall != nil {
		for _, rule := range config.Firewall.Global {
			if err := m.validateFirewallRule(rule, "global"); err != nil {
				return err
			}
		}
		
		for serverName, rules := range config.Firewall.ServerSpecific {
			if config.Servers[serverName] == nil {
				return fmt.Errorf("firewall: references non-existent server '%s'", serverName)
			}
			
			for _, rule := range rules {
				if err := m.validateFirewallRule(rule, fmt.Sprintf("server '%s'", serverName)); err != nil {
					return err
				}
			}
		}
	}
	
	return nil
}

// validateFirewallRule validates a single firewall rule
func (m *Manager) validateFirewallRule(rule FirewallRule, context string) error {
	if rule.Port <= 0 || rule.Port > 65535 {
		return fmt.Errorf("firewall %s: invalid port %d", context, rule.Port)
	}
	
	validProtocols := map[string]bool{
		"tcp":     true,
		"udp":     true,
		"tcp/udp": true,
	}
	
	if !validProtocols[rule.Protocol] {
		return fmt.Errorf("firewall %s: invalid protocol '%s'", context, rule.Protocol)
	}
	
	if rule.From == "" {
		return fmt.Errorf("firewall %s: 'from' field is required", context)
	}
	
	return nil
}

// saveRuntimeConfig saves the current runtime configuration
func (m *Manager) saveRuntimeConfig() error {
	if m.runtime == nil {
		return fmt.Errorf("no runtime config to save")
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	configPath := filepath.Join(homeDir, ".mah", "config.yaml")
	
	data, err := yaml.Marshal(m.runtime)
	if err != nil {
		return fmt.Errorf("failed to marshal runtime config: %w", err)
	}
	
	return os.WriteFile(configPath, data, 0644)
}