# MAH Architecture Documentation

MAH (Modern Application Hub) is a next-generation infrastructure orchestration tool that provides a unified interface for managing distributed containerized applications across multiple servers with automatic SSL, firewall management, and extensible plugin architecture.

## Core Philosophy

MAH transforms the complexity of multi-server Docker deployment into a simple, declarative configuration experience. It eliminates the chicken-and-egg problem of deployment infrastructure by providing a single binary that bootstraps and manages everything needed for modern containerized applications.

## Nexus: The Central Coordination Concept

### What is a Nexus?

A **Nexus** represents an elastic connection point to your infrastructure - a logical grouping of servers, services, and configurations that work together as a cohesive unit. Unlike traditional "contexts" which are just connection endpoints, a Nexus encompasses:

- Server connection and authentication details
- Service configurations and inter-dependencies  
- Network topology and security policies
- Resource allocation and scaling parameters
- Monitoring and observability configuration

### Nexus Architecture

```
┌─────────────────────────────────────────────────────┐
│                    MAH Core                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │   Config    │  │    CLI      │  │   Nexus     │  │
│  │   Parser    │  │  Interface  │  │  Manager    │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────┘
           │                    │                    │
           ▼                    ▼                    ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Nexus: thor-prod│    │ Nexus: odin-stage│   │ Nexus: loki-dev │
│  ├─ Server: VPS1 │    │ ├─ Server: VPS2  │   │ ├─ Server: Home │
│  ├─ Services: 5  │    │ ├─ Services: 3   │   │ ├─ Services: 2  │
│  ├─ Networks: 2  │    │ ├─ Networks: 1   │   │ ├─ Networks: 1  │
│  └─ Policies: FW │    │ └─ Policies: FW  │   │ └─ Policies: FW │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Multi-Server Architecture

### Server Abstraction Layer

MAH provides a unified interface across different Linux distributions and server configurations:

```go
type Server interface {
    // Core server operations
    Connect() error
    Execute(cmd string, sudo bool) (*Result, error)
    TransferFile(local, remote string) error
    
    // System information
    GetDistro() string
    GetResources() (*ResourceInfo, error)
    HealthCheck() error
}

type ServerConfig struct {
    Host     string `yaml:"host"`
    SSHUser  string `yaml:"ssh_user"`
    SSHKey   string `yaml:"ssh_key"`
    Sudo     bool   `yaml:"sudo"`
    Distro   string `yaml:"distro"`
    NexusID  string `yaml:"nexus"`
}
```

### Supported Distributions

| Distribution | Firewall | Package Manager | Container Runtime | Status |
|-------------|----------|-----------------|-------------------|--------|
| Ubuntu 20.04+ | ufw | apt | Docker CE | Full Support |
| Debian 11+ | ufw | apt | Docker CE | Full Support |
| Rocky Linux 8+ | firewalld | dnf | Docker CE | Full Support |
| Alpine Linux | iptables | apk | Docker CE | Basic Support |
| Fedora 36+ | firewalld | dnf | Docker CE | Full Support |

## Plugin Architecture

MAH is designed with extensibility at its core. All major functionality is implemented as plugins, allowing easy addition of new providers and capabilities.

### Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string
    Initialize(config map[string]interface{}) error
    Execute(action string, params map[string]interface{}) (*Result, error)
    Cleanup() error
}
```

### Core Plugin Categories

#### 1. DNS Providers
- **name.com** (primary)
- cloudflare
- route53
- digitalocean
- google-dns

#### 2. Container Orchestration
- **docker-compose** (primary)
- docker-swarm
- kubernetes (future)

#### 3. Reverse Proxy & SSL
- **traefik** (primary)
- nginx
- caddy
- envoy (future)

#### 4. Monitoring & Observability
- prometheus + grafana
- loki + grafana
- datadog
- new-relic

#### 5. Backup & Storage
- restic
- duplicity
- s3
- gcs

#### 6. Cloud Providers
- digitalocean
- linode
- aws
- gcp
- hetzner

### Plugin Discovery and Loading

```go
type PluginManager struct {
    plugins map[string]Plugin
    config  *Config
}

// Plugins are automatically discovered and loaded
func (pm *PluginManager) LoadPlugins(pluginDir string) error {
    // Load built-in plugins
    pm.loadBuiltinPlugins()
    
    // Load external plugins from directory
    pm.loadExternalPlugins(pluginDir)
    
    return nil
}
```

## Configuration Schema

### Primary Configuration File: mah.yaml

```yaml
# Global configuration
version: "1.0"
project: "my-infrastructure"

# Server definitions
servers:
  thor:
    host: "185.x.x.x"
    ssh_user: "jonas"
    ssh_key: "~/.ssh/id_rsa"
    sudo: true
    distro: "ubuntu"
    nexus: "thor-prod"
    
  odin:
    host: "142.x.x.x"
    ssh_user: "deploy"
    ssh_key: "~/.ssh/deploy_key"
    sudo: true
    distro: "debian"
    nexus: "odin-staging"

# Nexus definitions - logical groupings
nexuses:
  thor-prod:
    description: "Production environment"
    servers: ["thor"]
    environment: "production"
    
  odin-staging:
    description: "Staging environment" 
    servers: ["odin"]
    environment: "staging"

# Service definitions
services:
  blog:
    servers: ["thor", "odin"]
    image: "wordpress:latest"
    domains:
      thor: "blog.example.com"
      odin: "staging.blog.example.com" 
    public: true
    environment:
      WORDPRESS_DB_HOST: "mysql"
      WORDPRESS_DB_PASSWORD: "${SECRET_DB_PASS}"
    
  mysql:
    servers: ["thor"]
    image: "mysql:8"
    internal: true
    volumes:
      - "mysql_data:/var/lib/postgresql/data"

# Plugin configurations
plugins:
  dns:
    provider: "name.com"
    config:
      username: "${NAMECOM_USERNAME}"
      token: "${NAMECOM_TOKEN}"
      
  ssl:
    provider: "traefik"
    email: "admin@example.com"
    config:
      dns_challenge: true
      dns_provider: "name.com"
      
  monitoring:
    provider: "prometheus"
    config:
      retention: "30d"
      scrape_interval: "15s"

# Firewall rules
firewall:
  global:
    - port: 22
      from: "any"
    - port: 80  
      from: "any"
    - port: 443
      from: "any"
      
  server_specific:
    thor:
      - port: 5432
        from: "100.0.0.0/8"  # Headscale network
```

## Command Structure

### Primary Commands

```bash
# Nexus Management
mah nexus list                    # List all configured nexuses
mah nexus create <name>           # Create nexus from config
mah nexus switch <name>           # Switch active nexus
mah nexus current                 # Show current nexus
mah nexus status                  # Health check current nexus

# Server Management  
mah server init <server>          # Initialize server (Docker, FW, etc.)
mah server status [server]        # Server status
mah server update [server]        # Update server packages

# Service Management
mah service deploy <name>         # Deploy service to current nexus
mah service status [name]         # Service status  
mah service logs <name> [-f]      # Service logs
mah service scale <name> <count>  # Scale service

# SSL Certificate Management
mah certs status                  # Certificate status
mah certs renew [domain]          # Force renewal
mah certs list                    # List all certificates

# Firewall Management
mah fw status                     # Current firewall rules
mah fw reload                     # Reload from config
mah fw add <rule>                 # Add temporary rule

# Plugin Management
mah plugin list                   # List loaded plugins
mah plugin info <name>            # Plugin information
mah plugin install <name>         # Install external plugin

# Configuration
mah config validate               # Validate configuration
mah config show                   # Show merged configuration
mah config encrypt <key>          # Encrypt sensitive values
```

### Cross-Nexus Operations

```bash
# Target specific nexus
mah --nexus=thor-prod service deploy blog
mah -n odin-staging service status

# Operate on multiple nexuses
mah --all service status          # Status across all nexuses
mah service deploy blog --targets=thor,odin

# Quick nexus switching
mah use thor-prod                 # Switch active nexus
mah @odin-staging service deploy blog  # One-off command
```

## State Management

### Local State Storage

MAH maintains local state for:
- Active nexus configuration
- Certificate expiration tracking
- Service deployment history
- Server resource utilization
- Plugin state and cache

```
~/.mah/
├── config.yaml              # User configuration overrides
├── state/
│   ├── nexus-current.yaml   # Current active nexus
│   ├── certificates.db     # Certificate tracking
│   └── deployment-history.db
├── cache/
│   ├── server-info/         # Cached server information  
│   └── docker-images/       # Image cache metadata
└── plugins/
    ├── external/            # External plugin binaries
    └── config/              # Plugin configurations
```

## Security Model

### Authentication & Authorization

1. **SSH Key-Based Authentication**: All server connections use SSH keys
2. **Sudo Privilege Escalation**: Configurable per server
3. **Secret Management**: Environment variable substitution with optional encryption
4. **TLS Certificate Management**: Automatic Let's Encrypt integration
5. **Network Security**: Firewall rule automation

### Network Security

MAH enforces security best practices:
- Default-deny firewall rules
- Automatic security updates
- SSH hardening
- Container isolation
- TLS-only communication for management

## Extensibility Points

### Adding New DNS Providers

```go
type DNSProvider interface {
    CreateRecord(domain, name, recordType, value string) error
    UpdateRecord(domain, name, recordType, value string) error  
    DeleteRecord(domain, name, recordType string) error
    ListRecords(domain string) ([]DNSRecord, error)
}

// Register provider
func init() {
    RegisterDNSProvider("name.com", &NameComProvider{})
}
```

### Adding New Monitoring Systems

```go
type MonitoringProvider interface {
    Setup(servers []Server) error
    Configure(metrics []MetricConfig) error
    CreateDashboard(name string, config DashboardConfig) error
}
```

### Adding New Cloud Providers

```go
type CloudProvider interface {
    CreateServer(config ServerConfig) (*Server, error)
    ListServers() ([]*Server, error)
    DeleteServer(id string) error
    CreateLoadBalancer(config LBConfig) error
}
```

This modular architecture allows MAH to grow organically while maintaining a consistent user experience across all providers and services.