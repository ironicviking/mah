# MAH Configuration Examples

This directory contains example configurations and usage patterns for MAH.

## 🔐 Secret Management Examples

### Example 1: Environment Variables (CI/CD Friendly)

**mah.yaml** (safe to commit):
```yaml
version: "1.0"
project: "production-app"

servers:
  web:
    host: "${WEB_SERVER_HOST}"
    ssh_user: "${SSH_USER}"
    ssh_key: "~/.ssh/deploy_key"
    sudo: true
    distro: "ubuntu"
    nexus: "production"

plugins:
  dns:
    provider: "name.com"
    config:
      username: "${NAMECOM_USERNAME}"
      token: "${NAMECOM_TOKEN}"
```

**.env** file (add to .gitignore):
```bash
export WEB_SERVER_HOST="203.0.113.10"
export SSH_USER="deploy"
export NAMECOM_USERNAME="myusername"
export NAMECOM_TOKEN="abc123xyz789"
```

Usage:
```bash
source .env
mah config validate
mah nexus list
```

### Example 2: Encrypted Secrets (Team Collaboration)

**mah.yaml** (safe to commit):
```yaml
version: "1.0"
project: "team-project"

servers:
  staging:
    host: "staging.example.com"
    ssh_user: "deploy"
    ssh_key: "~/.ssh/staging_key"
    sudo: true
    distro: "ubuntu"
    nexus: "staging"

services:
  database:
    servers: ["staging"]
    image: "postgres:15"
    environment:
      POSTGRES_PASSWORD: "${DB_PASSWORD}"
```

**~/.mah/secrets.yaml** (encrypted, safe to commit):
```yaml
secrets:
    DB_PASSWORD: "encrypted_base64_string_here"
encrypted: true
key_source: env
```

Usage:
```bash
# One-time setup per team member
export MAH_MASTER_KEY="team-shared-32-character-key-here"

# Use normally
mah config validate
mah service deploy database
```

### Example 3: Mixed Approach (Production)

**mah.yaml**:
```yaml
version: "1.0"
project: "enterprise-app"

servers:
  prod-web-1:
    host: "${PROD_WEB_1_HOST}"
    ssh_user: "deploy"
    ssh_key: "~/.ssh/production_key"
    sudo: true
    distro: "ubuntu"
    nexus: "production"
    
  prod-web-2:
    host: "${PROD_WEB_2_HOST}"
    ssh_user: "deploy"
    ssh_key: "~/.ssh/production_key"
    sudo: true
    distro: "ubuntu"
    nexus: "production"

nexuses:
  production:
    description: "Production environment"
    servers: ["prod-web-1", "prod-web-2"]
    environment: "production"

services:
  app:
    servers: ["prod-web-1", "prod-web-2"]
    image: "myapp:latest"
    domains:
      prod-web-1: "app.example.com"
      prod-web-2: "app2.example.com"
    public: true
    environment:
      DATABASE_URL: "${DATABASE_URL}"
      REDIS_URL: "${REDIS_URL}"
      API_SECRET: "${API_SECRET}"

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

firewall:
  global:
    - port: 22
      protocol: tcp
      from: "10.0.0.0/8"  # VPN network only
      comment: "SSH from VPN"
    - port: 80
      protocol: tcp
      from: "any"
      comment: "HTTP traffic"
    - port: 443
      protocol: tcp
      from: "any"
      comment: "HTTPS traffic"
```

## 🚀 Usage Workflows

### Development Workflow

```bash
# 1. Clone repository
git clone https://github.com/company/infrastructure.git
cd infrastructure

# 2. Copy template and configure
cp mah.template.yaml mah.yaml

# 3. Set development environment variables
export SERVER_HOST="dev.internal.company.com"
export SSH_USER="dev-deploy"
# ... other variables

# 4. Validate and deploy
mah config validate
mah nexus switch development
mah service deploy app
```

### Production Deployment (CI/CD)

```bash
# In CI/CD pipeline (GitHub Actions, GitLab CI, etc.)

# 1. Set secrets in CI/CD environment
# SERVER_HOST, SSH_USER, API_KEYS, etc.

# 2. Deploy
mah config validate
mah nexus switch production
mah service deploy --all

# 3. Verify deployment
mah nexus status production
```

### Team Collaboration

```bash
# Team lead setup (one-time) - Easy way with direct password
TEAM_KEY="generate-secure-32-char-key"
mah config secrets init -p "$TEAM_KEY"
# Edit ~/.mah/secrets.yaml with actual values
mah config secrets encrypt -p "$TEAM_KEY"
git add ~/.mah/secrets.yaml
git commit -m "Add encrypted team secrets"

# Team members (each person) - Use same key
TEAM_KEY="same-secure-32-char-key"
mah config secrets decrypt -p "$TEAM_KEY"
mah config validate

# Or use environment variable approach
export MAH_MASTER_KEY="same-secure-32-char-key"
mah config secrets decrypt
mah config validate
```

## 🛡️ Security Patterns

### Multi-Environment Secrets

Create separate secret files for each environment:

```bash
# Development
cp ~/.mah/secrets.yaml ~/.mah/secrets-dev.yaml
# Edit with dev values

# Production  
cp ~/.mah/secrets.yaml ~/.mah/secrets-prod.yaml
# Edit with prod values

# Use environment-specific secrets
export MAH_ENVIRONMENT=production
mah config validate --config production.yaml
```

### Rotating Secrets

```bash
# 1. Update secret values
vim ~/.mah/secrets.yaml

# 2. Re-encrypt with new master key
export MAH_MASTER_KEY="new-32-character-master-key"
mah config secrets encrypt

# 3. Update services
mah service deploy --all

# 4. Share new master key with team securely
```

### Audit Trail

```bash
# Check for potential security issues
mah config validate

# Create sanitized version for documentation
mah config secrets sanitize mah.yaml docs/mah-example.yaml

# Verify no secrets in git history
git log --grep="password\|token\|key" --oneline
```

## 📁 File Organization

```
project/
├── mah.template.yaml          # Template (committed)
├── mah.yaml                   # Actual config (gitignored)
├── .env                       # Environment variables (gitignored)
├── .gitignore                 # Excludes sensitive files
├── ~/.mah/
│   ├── secrets.yaml           # Encrypted secrets (safe to commit)
│   └── config.yaml           # MAH runtime config
└── docs/
    └── deployment.md          # Deployment documentation
```

## 🔧 Troubleshooting

### "Secret validation failed"
```bash
# Check for plain-text secrets in config
mah config validate

# Move secrets to environment variables
mah config secrets sanitize
```

### "Encryption key not found"
```bash
# Verify master key is set
echo $MAH_MASTER_KEY

# Re-initialize if needed
mah config secrets init
```

### "Permission denied"
```bash
# Check file permissions
ls -la ~/.mah/secrets.yaml

# Fix permissions
chmod 600 ~/.mah/secrets.yaml
```

These examples show various patterns for securely managing MAH configurations in different environments and team structures.