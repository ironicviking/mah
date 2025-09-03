# MAH Development TODO

This TODO focuses on implementation tasks only. Testing, deployment, and documentation tasks will be added in later phases.

## Phase 1: Core Foundation

### 1.1 Project Structure Setup
- [ ] Create Go module structure (`mah/`)
- [ ] Initialize `go.mod` with dependencies (cobra, viper, yaml.v3, ssh, etc.)
- [ ] Set up directory structure:
  - [ ] `cmd/mah/` - CLI entry point
  - [ ] `internal/core/` - Core MAH logic
  - [ ] `internal/config/` - Configuration management
  - [ ] `internal/nexus/` - Nexus management
  - [ ] `internal/server/` - Server abstraction
  - [ ] `internal/plugins/` - Plugin framework
  - [ ] `pkg/` - Public interfaces
- [ ] Create `Makefile` for build automation
- [ ] Set up cross-compilation targets (Linux, macOS, Windows)

### 1.2 Configuration System
- [ ] Implement YAML configuration parser with validation
- [ ] Create configuration schema structs:
  - [ ] `ServerConfig` struct
  - [ ] `NexusConfig` struct  
  - [ ] `ServiceConfig` struct
  - [ ] `PluginConfig` struct
  - [ ] `FirewallConfig` struct
- [ ] Implement environment variable substitution (`${VAR}`)
- [ ] Add configuration validation with detailed error messages
- [ ] Implement configuration merging (defaults + user config)
- [ ] Add encrypted secret support for sensitive values

### 1.3 CLI Framework
- [ ] Set up Cobra CLI framework with command structure:
  - [ ] `mah nexus` commands
  - [ ] `mah server` commands
  - [ ] `mah service` commands
  - [ ] `mah certs` commands
  - [ ] `mah fw` commands
  - [ ] `mah plugin` commands
  - [ ] `mah config` commands
- [ ] Implement global flags (`--nexus`, `--all`, `--config`)
- [ ] Add colored output support with status indicators
- [ ] Implement interactive nexus selection
- [ ] Add progress bars for long-running operations

## Phase 2: Server Management

### 2.1 Server Abstraction Layer
- [ ] Create `Server` interface with core operations
- [ ] Implement SSH connection management with key-based auth
- [ ] Add sudo command execution support
- [ ] Implement file transfer capabilities (SCP/SFTP)
- [ ] Create server health check functionality
- [ ] Add system information gathering (distro, resources, Docker status)
- [ ] Implement connection pooling and reuse

### 2.2 Multi-Distribution Support
- [ ] Create distro detection logic
- [ ] Implement Ubuntu/Debian support:
  - [ ] Package management via `apt`
  - [ ] Firewall management via `ufw`
  - [ ] Service management via `systemctl`
- [ ] Implement Rocky/RHEL/Fedora support:
  - [ ] Package management via `dnf/yum`
  - [ ] Firewall management via `firewalld`
  - [ ] Service management via `systemctl`
- [ ] Implement Alpine Linux support:
  - [ ] Package management via `apk`
  - [ ] Firewall management via `iptables`
  - [ ] Service management via `rc-service`
- [ ] Create distro-specific factory pattern

### 2.3 Server Initialization
- [ ] Implement Docker CE installation for each distro
- [ ] Add Docker daemon configuration management
- [ ] Implement basic firewall setup
- [ ] Add automatic security updates configuration
- [ ] Create SSH hardening functionality
- [ ] Implement fail2ban setup where supported
- [ ] Add system monitoring agent installation

## Phase 3: Nexus Management

### 3.1 Nexus Core
- [ ] Implement `Nexus` struct and manager
- [ ] Create nexus state persistence (`~/.mah/state/`)
- [ ] Add nexus switching and current tracking
- [ ] Implement nexus validation and health checks
- [ ] Create nexus listing and status display
- [ ] Add cross-nexus operation coordination

### 3.2 Nexus Operations
- [ ] Implement multi-server command execution
- [ ] Add parallel operation support with error handling
- [ ] Create server grouping and targeting logic
- [ ] Implement nexus-specific configuration overrides
- [ ] Add nexus backup and restore functionality

## Phase 4: Plugin Framework

### 4.1 Plugin Core System
- [ ] Design and implement `Plugin` interface
- [ ] Create plugin registration and discovery system
- [ ] Implement plugin lifecycle management (init, execute, cleanup)
- [ ] Add plugin configuration management
- [ ] Create plugin dependency resolution
- [ ] Implement plugin versioning and compatibility checks

### 4.2 Built-in Plugins

#### 4.2.1 DNS Plugin Framework
- [ ] Create `DNSProvider` interface
- [ ] Implement Name.com DNS provider:
  - [ ] API client for Name.com
  - [ ] Record creation, update, deletion
  - [ ] Domain validation
  - [ ] A, AAAA, CNAME, TXT record support
- [ ] Add DNS provider factory and registration
- [ ] Implement DNS challenge support for Let's Encrypt

#### 4.2.2 Container Orchestration Plugin
- [ ] Create `ContainerProvider` interface
- [ ] Implement Docker Compose provider:
  - [ ] Compose file generation from service configs
  - [ ] Service deployment and management
  - [ ] Volume and network management
  - [ ] Container health monitoring
- [ ] Add Docker context management integration
- [ ] Implement service scaling functionality

#### 4.2.3 Reverse Proxy Plugin
- [ ] Create `ProxyProvider` interface
- [ ] Implement Traefik provider:
  - [ ] Dynamic configuration generation
  - [ ] Service discovery via Docker labels
  - [ ] Let's Encrypt integration (HTTP and DNS challenges)
  - [ ] Automatic certificate renewal
  - [ ] Load balancing configuration
  - [ ] Access control and middleware support

#### 4.2.4 Firewall Plugin
- [ ] Create `FirewallProvider` interface
- [ ] Implement UFW provider (Ubuntu/Debian)
- [ ] Implement firewalld provider (RHEL/Fedora)
- [ ] Implement iptables provider (Alpine/minimal)
- [ ] Add rule validation and conflict detection
- [ ] Implement rule templating and generation

### 4.3 Plugin Extensions
- [ ] Create plugin loading from external binaries
- [ ] Add plugin configuration validation
- [ ] Implement plugin health monitoring
- [ ] Create plugin marketplace/registry support
- [ ] Add plugin update mechanism

## Phase 5: Service Management

### 5.1 Service Core
- [ ] Implement service configuration parsing and validation
- [ ] Create service dependency resolution
- [ ] Add service state management and persistence
- [ ] Implement service health monitoring
- [ ] Create service scaling and resource management

### 5.2 Service Deployment
- [ ] Implement single-server service deployment
- [ ] Add multi-server deployment coordination
- [ ] Create rolling deployment support
- [ ] Implement deployment rollback functionality
- [ ] Add deployment history and audit trail

### 5.3 Service Operations
- [ ] Implement service start/stop/restart operations
- [ ] Add service log aggregation and streaming
- [ ] Create service metrics collection
- [ ] Implement service backup and restore
- [ ] Add service migration between servers

## Phase 6: SSL Certificate Management

### 6.1 Certificate Core
- [ ] Implement certificate tracking and storage
- [ ] Create certificate expiration monitoring
- [ ] Add automatic renewal scheduling
- [ ] Implement certificate validation and health checks

### 6.2 Let's Encrypt Integration
- [ ] Implement HTTP-01 challenge support
- [ ] Add DNS-01 challenge support with provider integration
- [ ] Create certificate provisioning automation
- [ ] Implement wildcard certificate support
- [ ] Add certificate deployment to services

## Phase 7: State Management

### 7.1 Local State
- [ ] Create state directory management (`~/.mah/`)
- [ ] Implement nexus state persistence
- [ ] Add certificate tracking database
- [ ] Create deployment history storage
- [ ] Implement configuration caching

### 7.2 Remote State Sync
- [ ] Add server state synchronization
- [ ] Implement conflict resolution for state changes
- [ ] Create state backup and recovery
- [ ] Add state migration tools

## Phase 8: Advanced Features

### 8.1 Monitoring Integration
- [ ] Create monitoring plugin interface
- [ ] Implement basic metrics collection
- [ ] Add health check aggregation
- [ ] Create alerting integration points

### 8.2 Backup Integration
- [ ] Create backup plugin interface
- [ ] Implement volume backup scheduling
- [ ] Add configuration backup automation
- [ ] Create disaster recovery procedures

### 8.3 Security Enhancements
- [ ] Implement secret management with encryption
- [ ] Add audit logging for all operations
- [ ] Create security policy enforcement
- [ ] Implement access control and RBAC

## Phase 9: Quality & Performance

### 9.1 Error Handling
- [ ] Implement comprehensive error handling with context
- [ ] Add operation retry logic with exponential backoff
- [ ] Create graceful degradation for network failures
- [ ] Implement operation timeout management

### 9.2 Performance Optimization
- [ ] Add concurrent operation support where safe
- [ ] Implement connection pooling and reuse
- [ ] Create caching for expensive operations
- [ ] Add operation progress tracking

### 9.3 Logging and Observability
- [ ] Implement structured logging with levels
- [ ] Add operation tracing and debugging
- [ ] Create performance metrics collection
- [ ] Implement log rotation and management

## Development Priorities

### Milestone 1: Basic Functionality (Weeks 1-3)
- Phase 1: Core Foundation
- Phase 2.1: Server Abstraction Layer
- Phase 3.1: Nexus Core
- Basic CLI commands working

### Milestone 2: Server Management (Weeks 4-6)
- Phase 2.2: Multi-Distribution Support
- Phase 2.3: Server Initialization
- Phase 4.2.4: Firewall Plugin
- Server provisioning working end-to-end

### Milestone 3: Service Deployment (Weeks 7-9)
- Phase 4.2.2: Container Orchestration Plugin
- Phase 4.2.3: Reverse Proxy Plugin
- Phase 5.1-5.2: Service Management
- Basic service deployment working

### Milestone 4: SSL and DNS (Weeks 10-12)
- Phase 4.2.1: DNS Plugin Framework
- Phase 6: SSL Certificate Management
- Fully automated SSL certificate provisioning

### Milestone 5: Production Ready (Weeks 13-16)
- Phase 7: State Management
- Phase 9: Quality & Performance
- Phase 8.3: Security Enhancements
- Production-ready release

## Implementation Notes

- Use Go 1.21+ for generics and latest features
- Leverage existing libraries: cobra, viper, yaml.v3, crypto/ssh
- Implement interfaces first, then concrete implementations
- Use dependency injection for testability
- Follow Go best practices for project structure
- Use context for cancellation and timeouts
- Implement graceful shutdown for long-running operations
- Use structured logging (logrus or zap)
- Follow semantic versioning for releases