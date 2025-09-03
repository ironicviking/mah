package config

// Config represents the main MAH configuration
type Config struct {
	Version  string              `yaml:"version" mapstructure:"version"`
	Project  string              `yaml:"project" mapstructure:"project"`
	Servers  map[string]*Server  `yaml:"servers" mapstructure:"servers"`
	Nexuses  map[string]*Nexus   `yaml:"nexuses" mapstructure:"nexuses"`
	Services map[string]*Service `yaml:"services" mapstructure:"services"`
	Plugins  *PluginConfigs      `yaml:"plugins" mapstructure:"plugins"`
	Firewall *FirewallConfig     `yaml:"firewall" mapstructure:"firewall"`
}

// Server represents a server configuration
type Server struct {
	Host    string `yaml:"host" mapstructure:"host"`
	SSHUser string `yaml:"ssh_user" mapstructure:"ssh_user"`
	SSHKey  string `yaml:"ssh_key" mapstructure:"ssh_key"`
	SSHPort int    `yaml:"ssh_port,omitempty" mapstructure:"ssh_port"`
	Sudo    bool   `yaml:"sudo" mapstructure:"sudo"`
	Distro  string `yaml:"distro" mapstructure:"distro"`
	Nexus   string `yaml:"nexus" mapstructure:"nexus"`
}

// Nexus represents a nexus (logical grouping) configuration
type Nexus struct {
	Description string   `yaml:"description" mapstructure:"description"`
	Servers     []string `yaml:"servers" mapstructure:"servers"`
	Environment string   `yaml:"environment" mapstructure:"environment"`
}

// Service represents a service configuration
type Service struct {
	Servers     []string          `yaml:"servers"`
	Image       string            `yaml:"image"`
	Domains     map[string]string `yaml:"domains"`
	Public      bool              `yaml:"public"`
	Internal    bool              `yaml:"internal"`
	Ports       []int             `yaml:"ports"`
	Environment map[string]string `yaml:"environment"`
	Volumes     []string          `yaml:"volumes"`
	Networks    []string          `yaml:"networks"`
	Depends     []string          `yaml:"depends_on"`
	Auth        *AuthConfig       `yaml:"auth"`
	Labels      map[string]string `yaml:"labels"`
	Replicas    int               `yaml:"replicas,omitempty"`
}

// AuthConfig represents authentication configuration for a service
type AuthConfig struct {
	Type  string            `yaml:"type"` // basic, oauth, none
	Users map[string]string `yaml:"users"` // username: password_hash
}

// PluginConfigs represents plugin configurations
type PluginConfigs struct {
	DNS        *PluginConfig `yaml:"dns"`
	SSL        *PluginConfig `yaml:"ssl"`
	Monitoring *PluginConfig `yaml:"monitoring"`
	Backup     *PluginConfig `yaml:"backup"`
}

// PluginConfig represents a single plugin configuration
type PluginConfig struct {
	Provider string                 `yaml:"provider"`
	Config   map[string]interface{} `yaml:"config"`
}

// FirewallConfig represents firewall configuration
type FirewallConfig struct {
	Global         []FirewallRule            `yaml:"global"`
	ServerSpecific map[string][]FirewallRule `yaml:"server_specific"`
}

// FirewallRule represents a firewall rule in configuration
type FirewallRule struct {
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"`
	From     string `yaml:"from"`
	Comment  string `yaml:"comment"`
}

// RuntimeConfig represents runtime configuration and state
type RuntimeConfig struct {
	CurrentNexus string            `yaml:"current_nexus"`
	StateDir     string            `yaml:"state_dir"`
	CacheDir     string            `yaml:"cache_dir"`
	PluginDir    string            `yaml:"plugin_dir"`
	LogLevel     string            `yaml:"log_level"`
	ConfigFile   string            `yaml:"config_file"`
	Secrets      map[string]string `yaml:"secrets,omitempty"`
}