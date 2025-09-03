package pkg

import "context"

// Server represents a remote server that MAH can manage
type Server interface {
	// Core server operations
	Connect(ctx context.Context) error
	Execute(ctx context.Context, cmd string, sudo bool) (*Result, error)
	TransferFile(ctx context.Context, local, remote string) error
	Disconnect() error

	// System information
	GetDistro(ctx context.Context) (string, error)
	GetResources(ctx context.Context) (*ResourceInfo, error)
	HealthCheck(ctx context.Context) error

	// Server identification
	ID() string
	Host() string
}

// Plugin represents a MAH plugin that provides specific functionality
type Plugin interface {
	Name() string
	Version() string
	Initialize(config map[string]interface{}) error
	Execute(action string, params map[string]interface{}) (*Result, error)
	Cleanup() error
}

// DNSProvider interface for DNS management plugins
type DNSProvider interface {
	CreateRecord(domain, name, recordType, value string) error
	UpdateRecord(domain, name, recordType, value string) error
	DeleteRecord(domain, name, recordType string) error
	ListRecords(domain string) ([]DNSRecord, error)
	ValidateDomain(domain string) error
}

// ContainerProvider interface for container orchestration plugins
type ContainerProvider interface {
	Deploy(config *ServiceConfig) error
	Scale(serviceName string, replicas int) error
	Status(serviceName string) (*ServiceStatus, error)
	Logs(serviceName string, follow bool) (<-chan string, error)
	Remove(serviceName string) error
}

// ProxyProvider interface for reverse proxy plugins
type ProxyProvider interface {
	Configure(services []*ServiceConfig) error
	GetCertificates() ([]Certificate, error)
	RenewCertificate(domain string) error
	UpdateRoutes(routes []Route) error
}

// FirewallProvider interface for firewall management plugins
type FirewallProvider interface {
	ApplyRules(rules []FirewallRule) error
	GetRules() ([]FirewallRule, error)
	RemoveRule(rule FirewallRule) error
	Status() (*FirewallStatus, error)
}

// Result represents the result of a command execution
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration int64 // milliseconds
}

// ResourceInfo contains server resource information
type ResourceInfo struct {
	CPU    CPUInfo    `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
	Disk   DiskInfo   `json:"disk"`
	Load   LoadInfo   `json:"load"`
}

type CPUInfo struct {
	Cores   int     `json:"cores"`
	Usage   float64 `json:"usage"`   // percentage
	Model   string  `json:"model"`
	Arch    string  `json:"arch"`
}

type MemoryInfo struct {
	Total     int64   `json:"total"`     // bytes
	Available int64   `json:"available"` // bytes
	Used      int64   `json:"used"`      // bytes
	Usage     float64 `json:"usage"`     // percentage
}

type DiskInfo struct {
	Total     int64   `json:"total"`     // bytes
	Available int64   `json:"available"` // bytes
	Used      int64   `json:"used"`      // bytes
	Usage     float64 `json:"usage"`     // percentage
}

type LoadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

// ServiceStatus represents the status of a deployed service
type ServiceStatus struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"` // running, stopped, error, etc.
	Replicas  int               `json:"replicas"`
	Health    string            `json:"health"`
	Ports     []int             `json:"ports"`
	Variables map[string]string `json:"variables"`
}

// Certificate represents an SSL certificate
type Certificate struct {
	Domain    string   `json:"domain"`
	Domains   []string `json:"domains"` // SANs
	Issuer    string   `json:"issuer"`
	ExpiresAt int64    `json:"expires_at"`
	IssuedAt  int64    `json:"issued_at"`
	Status    string   `json:"status"`
}

// Route represents a proxy route configuration
type Route struct {
	Domain      string            `json:"domain"`
	Service     string            `json:"service"`
	Port        int               `json:"port"`
	Path        string            `json:"path"`
	TLS         bool              `json:"tls"`
	Middlewares []string          `json:"middlewares"`
	Headers     map[string]string `json:"headers"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"` // tcp, udp, tcp/udp
	Source   string `json:"source"`   // IP, CIDR, or "any"
	Action   string `json:"action"`   // allow, deny
	Comment  string `json:"comment"`
}

// FirewallStatus represents current firewall status
type FirewallStatus struct {
	Active bool           `json:"active"`
	Rules  []FirewallRule `json:"rules"`
	Policy string         `json:"policy"` // default policy
}

// ServiceConfig represents a service configuration for deployment
type ServiceConfig struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Servers     []string          `json:"servers"`
	Domains     map[string]string `json:"domains"`
	Public      bool              `json:"public"`
	Internal    bool              `json:"internal"`
	Ports       []int             `json:"ports"`
	Environment map[string]string `json:"environment"`
	Volumes     []string          `json:"volumes"`
	Networks    []string          `json:"networks"`
	Depends     []string          `json:"depends_on"`
	Labels      map[string]string `json:"labels"`
	Replicas    int               `json:"replicas"`
}