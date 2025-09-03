# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

MAH (Modern Application Hub) is a next-generation infrastructure orchestration tool that provides a unified interface for managing distributed containerized applications across multiple servers with automatic SSL, firewall management, and extensible plugin architecture.

## Key Commands

### Build and Development
```bash
# Build the binary
make build

# Build for development (faster, no optimization)
make dev

# Install dependencies
make deps

# Run tests
make test

# Format and lint code
make fmt
make lint

# Cross-compile for multiple platforms
make cross-compile

# Create release packages with version
make release VERSION=2.0.5

# Development quick cycle
make quick   # Format + build
```

### Configuration and Secrets Management
```bash
# Create sample configuration with environment variables
mah config init

# Initialize encrypted secrets for team collaboration
mah config secrets init -p "team-key-32chars"

# Validate configuration
mah config validate

# View configuration (with masked secrets)
mah config show
```

### Version Management
```bash
# Check VERSION file for current version
cat VERSION

# Update version for releases
echo "2.0.5" > VERSION
```

### Git Operations
```bash
# Always commit changes, push to GitHub, and create releases
git add .
git commit -m "Description 🤖 Generated with Claude Code"
git push origin main
gh release create v$(cat VERSION) --generate-notes
```

## Architecture Overview

### Core Design Principles
- **Nexus-Centric Management**: Logical groupings of servers and services for environment management
- **Plugin Architecture**: Extensible system for DNS providers, container orchestration, SSL, monitoring
- **Multi-Server Support**: Deploy and manage applications across multiple servers simultaneously  
- **Declarative Configuration**: Single `mah.yaml` file defines entire infrastructure
- **Security-First**: Environment variable substitution with optional AES-256 encryption

### Module Structure

**cmd/mah/** - CLI application entry point
- `main.go` - Root command setup with Cobra CLI framework
- `nexus.go` - Nexus management commands (list, switch, status)
- `server.go` - Server management commands (init, status, update)
- `service.go` - Service deployment and management commands
- `config.go` - Configuration management commands
- `secrets.go` - Secret management and encryption commands

**internal/config/** - Configuration management system
- Environment variable substitution (`${VAR}` patterns)
- AES-256-GCM encryption for sensitive values
- YAML parsing with validation
- Runtime configuration merging

**internal/nexus/** - Nexus coordination system
- Multi-server command execution
- Server grouping and targeting logic
- Cross-nexus operation coordination
- State persistence in `~/.mah/state/`

**internal/server/** - Server abstraction layer
- SSH connection management with key-based authentication
- Multi-distribution support (Ubuntu, Debian, Rocky Linux, Alpine)
- System information gathering and health checks
- Distro-specific package managers and firewalls

**internal/plugins/** - Plugin framework
- DNS providers (name.com primary, extensible for others)
- Container orchestration (Docker Compose primary)
- SSL/reverse proxy (Traefik integration)
- Monitoring and backup providers

### Configuration Schema

Primary configuration in `mah.yaml`:
```yaml
version: "1.0"
project: "infrastructure-name"

# Server definitions with environment variable substitution
servers:
  thor:
    host: "${SERVER_HOST}"
    ssh_user: "${SSH_USER}"
    nexus: "production"

# Logical groupings
nexuses:
  production:
    servers: ["thor"]
    environment: "production"

# Service definitions
services:
  blog:
    image: "wordpress:latest"
    domains:
      thor: "${BLOG_DOMAIN}"
    environment:
      WORDPRESS_DB_PASSWORD: "${DB_PASSWORD}"

# Plugin configurations
plugins:
  dns:
    provider: "name.com"
    config:
      username: "${NAMECOM_USERNAME}"
      token: "${NAMECOM_TOKEN}"
```

### Secret Management

MAH supports multiple secure approaches:
1. **Environment Variables** (recommended for CI/CD): `export VAR=value`
2. **Encrypted Secrets** (recommended for teams): `mah config secrets init -p "key"`
3. **Template + .gitignore**: `mah config secrets sanitize`

Secrets use AES-256-GCM encryption with unique nonces and authenticated encryption.

## Development Guidelines

### Adding New Features
1. Follow existing modular plugin pattern in `internal/plugins/`
2. Add configuration options to YAML schema in `internal/config/types.go`
3. Implement interface-first design (see existing providers)
4. Add comprehensive error handling with context
5. Update VERSION file and create git tag for releases

### Testing and Validation
1. Run `make test` for unit tests
2. Use `make fmt` and `make lint` before commits
3. Test with `mah config validate` for configuration changes
4. Use `--dry-run` modes where available

### Release Process
1. Update VERSION file with semantic version
2. Run `make release VERSION=$(cat VERSION)` to create cross-platform binaries
3. Commit all changes with descriptive message
4. Push to GitHub: `git push origin main`
5. Create release: `gh release create v$(cat VERSION) --generate-notes`

## Security Considerations

- Never commit plain-text secrets to git
- Use SSH keys for server authentication (no passwords)
- Implement least-privilege firewall rules via `firewall.global` config
- Enable automatic security updates during server initialization
- All secrets support environment variable substitution: `"${SECRET_VAR}"`
- Team secrets encrypted with AES-256-GCM in `~/.mah/secrets.yaml`

## Multi-Distribution Support

| Distribution | Firewall | Package Manager | Status |
|-------------|----------|-----------------|--------|
| Ubuntu 20.04+ | ufw | apt | Full Support |
| Debian 11+ | ufw | apt | Full Support |
| Rocky Linux 8+ | firewalld | dnf | Full Support |
| Alpine Linux | iptables | apk | Basic Support |
| Fedora 36+ | firewalld | dnf | Full Support |

Server abstraction handles distro-specific differences automatically via factory pattern in `internal/server/factory.go`.