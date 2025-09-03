# MAH (Modern Application Hub)

MAH is a next-generation infrastructure orchestration tool that provides a unified interface for managing distributed containerized applications across multiple servers with automatic SSL, firewall management, and extensible plugin architecture.

## 🚀 Features

- **Nexus Management**: Logical groupings of servers and services for easy environment management
- **Multi-Server Support**: Deploy and manage applications across multiple servers simultaneously  
- **Automatic SSL**: Let's Encrypt integration with DNS-01 and HTTP-01 challenges
- **Firewall Management**: Automated firewall configuration across different Linux distributions
- **Plugin Architecture**: Extensible system for DNS providers, monitoring, backups, and more
- **Single Binary**: No dependencies, just download and run
- **Declarative Configuration**: Everything defined in a single `mah.yaml` file

## 🏗️ Architecture

MAH transforms complex multi-server Docker deployment into a simple, declarative experience:

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
│  Nexus: prod    │    │ Nexus: staging  │    │ Nexus: dev      │
│  ├─ Server: VPS1│    │ ├─ Server: VPS2 │    │ ├─ Server: Home │
│  ├─ Services: 5 │    │ ├─ Services: 3  │    │ ├─ Services: 2  │
│  └─ SSL: Auto   │    │ └─ SSL: Auto    │    │ └─ SSL: Auto    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### Installation

Download the latest binary for your platform:

```bash
# Linux/macOS
curl -L https://github.com/ironicviking/mah/releases/latest/download/mah-linux-amd64 -o mah
chmod +x mah
sudo mv mah /usr/local/bin/

# Or build from source
git clone https://github.com/ironicviking/mah.git
cd mah
make build
sudo cp build/mah /usr/local/bin/
```

### Initialize Configuration

```bash
# Create a sample configuration file (uses environment variables)
mah config init

# Set required environment variables
export SERVER_HOST="your.server.ip"
export SSH_USER="your-username"
export NAMECOM_USERNAME="your-namecom-user"
export NAMECOM_TOKEN="your-api-token"
# ... see output for complete list

# Or initialize secret management for teams
mah config secrets init
```

### Basic Usage

```bash
# Validate configuration
mah config validate

# List all nexuses
mah nexus list

# Switch to production nexus
mah nexus switch production

# Show nexus status
mah nexus status

# Deploy a service
mah service deploy blog

# Check service status
mah service status
```

## 📖 Configuration

MAH uses a single `mah.yaml` file to define your entire infrastructure. **Sensitive data is managed securely using environment variables or encrypted secrets**.

```yaml
version: "1.0"
project: "my-infrastructure"

# Server definitions (uses environment variables for security)
servers:
  thor:
    host: "${SERVER_HOST}"          # Environment variable
    ssh_user: "${SSH_USER}"         # Environment variable
    ssh_key: "~/.ssh/id_rsa"
    sudo: true
    distro: "ubuntu"
    nexus: "production"

# Nexus definitions (logical groupings)
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
      thor: "${BLOG_DOMAIN}"         # Environment variable
    public: true
    environment:
      WORDPRESS_DB_HOST: "mysql"
      WORDPRESS_DB_PASSWORD: "${MYSQL_PASSWORD}"

# Plugin configurations
plugins:
  dns:
    provider: "name.com"
    config:
      username: "${NAMECOM_USERNAME}"  # Environment variable
      token: "${NAMECOM_TOKEN}"        # Environment variable
      
  ssl:
    provider: "traefik"
    email: "${ADMIN_EMAIL}"           # Environment variable
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
```

### 🔐 Secret Management

MAH provides secure secret management with multiple options:

```bash
# 1. Environment Variables (recommended for CI/CD)
export SERVER_HOST="185.x.x.x"
export NAMECOM_TOKEN="your-api-token"

# 2. Encrypted Secrets (recommended for teams)
mah config secrets init
mah config secrets encrypt

# 3. Git-safe Templates
mah config secrets sanitize
```

See [SECURITY.md](SECURITY.md) for detailed security practices.

## 🔧 Commands

### Nexus Management
```bash
mah nexus list                    # List all nexuses
mah nexus switch <name>           # Switch active nexus
mah nexus current                 # Show current nexus
mah nexus status [name]           # Show nexus health
```

### Server Management
```bash
mah server list                   # List servers in current nexus
mah server init <name>            # Initialize server
mah server status [name]          # Show server status
```

### Service Management
```bash
mah service list                  # List services
mah service deploy <name>         # Deploy service
mah service status [name]         # Show service status
mah service logs <name> [-f]      # Show service logs
```

### Configuration
```bash
mah config init                   # Create sample config
mah config validate               # Validate configuration
mah config show                   # Show current config

# Secret Management
mah config secrets init           # Initialize secrets management
mah config secrets encrypt        # Encrypt secrets file
mah config secrets decrypt        # View secrets (masked)
mah config secrets sanitize       # Create git-safe template
```

## 🔌 Plugins

MAH's plugin architecture allows easy extension:

### DNS Providers
- **name.com** (built-in)
- cloudflare
- route53
- digitalocean

### SSL Providers
- **traefik** (built-in)
- nginx
- caddy

### Monitoring
- prometheus + grafana
- loki + grafana
- datadog

## 🏗️ Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/jonas-jonas/mah.git
cd mah

# Install dependencies
make deps

# Build binary
make build

# Run tests
make test

# Cross-compile for all platforms
make cross-compile
```

### Project Structure

```
mah/
├── cmd/mah/              # CLI application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── nexus/            # Nexus management
│   ├── server/           # Server abstraction
│   └── plugins/          # Plugin framework
├── pkg/                  # Public interfaces
├── templates/            # Configuration templates
└── Makefile             # Build automation
```

## 📋 Roadmap

- [x] **Phase 1**: Core foundation and CLI
- [x] **Phase 2**: Configuration system  
- [ ] **Phase 3**: Server management and SSH
- [ ] **Phase 4**: Service deployment with Docker
- [ ] **Phase 5**: Traefik integration and SSL
- [ ] **Phase 6**: Name.com DNS plugin
- [ ] **Phase 7**: Monitoring and alerting
- [ ] **Phase 8**: Backup automation

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests.

## 📄 License

MIT License - see LICENSE file for details.

## 🆘 Support

- Documentation: [docs/](docs/)
- Issues: [GitHub Issues](https://github.com/jonas-jonas/mah/issues)
- Discussions: [GitHub Discussions](https://github.com/jonas-jonas/mah/discussions)

---

**MAH** - Making infrastructure orchestration simple, secure, and scalable.