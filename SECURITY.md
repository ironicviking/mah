# MAH Security & Secret Management

MAH provides multiple secure approaches for managing sensitive configuration data while keeping your infrastructure code safe to commit to version control.

## 🔐 Secret Management Approaches

### 1. **Environment Variables (Recommended for CI/CD)**

The simplest and most widely supported approach:

```yaml
# mah.yaml - Safe to commit to git
servers:
  thor:
    host: "${SERVER_HOST}"
    ssh_user: "${SSH_USER}"
    ssh_key: "~/.ssh/id_rsa"

plugins:
  dns:
    provider: "name.com"
    config:
      username: "${NAMECOM_USERNAME}"
      token: "${NAMECOM_TOKEN}"
```

Set environment variables:
```bash
export SERVER_HOST="185.x.x.x"
export SSH_USER="deploy"
export NAMECOM_USERNAME="your-username"
export NAMECOM_TOKEN="your-api-token"
```

**Pros:**
- ✅ Universal support (works everywhere)
- ✅ Perfect for CI/CD pipelines
- ✅ No additional tooling required
- ✅ Config file is safe to commit

**Cons:**
- ❌ Environment variables visible in process lists
- ❌ Need to manage env vars across environments

### 2. **Encrypted Secrets File (Recommended for Teams)**

MAH can encrypt secrets using AES-256 encryption:

```bash
# Initialize secrets management
mah config secrets init

# Edit ~/.mah/secrets.yaml with actual values
vim ~/.mah/secrets.yaml

# Encrypt the secrets file
export MAH_MASTER_KEY="your-32-character-encryption-key"
mah config secrets encrypt

# Now secrets.yaml contains encrypted values - safe to commit!
```

**Pros:**
- ✅ Encrypted file is safe to commit to git
- ✅ Centralizes secret management
- ✅ Team can share encrypted secrets
- ✅ Works offline

**Cons:**
- ❌ Need to manage encryption key securely
- ❌ Additional setup complexity

### 3. **Template + .gitignore (Simple)**

Create sanitized template for git:

```bash
# Create git-safe template
mah config secrets sanitize mah.yaml mah.template.yaml

# Add actual config to .gitignore
echo "mah.yaml" >> .gitignore

# Commit template, ignore real config
git add mah.template.yaml .gitignore
git commit -m "Add MAH config template"
```

**Pros:**
- ✅ Very simple to understand
- ✅ No additional tools needed
- ✅ Template shows required variables

**Cons:**
- ❌ Easy to accidentally commit secrets
- ❌ No centralized secret management

### 4. **External Secret Management (Enterprise)**

Integrate with enterprise secret management:

```yaml
# Use external secret manager
plugins:
  dns:
    provider: "name.com"
    config:
      username: "${vault:secret/namecom#username}"
      token: "${vault:secret/namecom#token}"
```

**Pros:**
- ✅ Enterprise-grade security
- ✅ Centralized audit trails
- ✅ Role-based access control
- ✅ Automatic rotation support

**Cons:**
- ❌ Requires external infrastructure
- ❌ More complex setup

## 🚀 Quick Start Guide

### For Individual Projects

```bash
# 1. Initialize with environment variables (recommended)
mah config init

# 2. Set your environment variables
cat > .env << EOF
export SERVER_HOST="your.server.ip"
export SSH_USER="your-username"
export NAMECOM_USERNAME="your-namecom-user"
export NAMECOM_TOKEN="your-namecom-token"
export MYSQL_PASSWORD="secure-database-password"
EOF

# 3. Source variables (add to your shell profile)
source .env

# 4. Add .env to .gitignore
echo ".env" >> .gitignore

# 5. Validate configuration
mah config validate
```

### For Team Projects

```bash
# 1. Initialize secrets management
mah config secrets init

# 2. Edit secrets with actual values
vim ~/.mah/secrets.yaml

# 3. Encrypt secrets
export MAH_MASTER_KEY="shared-team-encryption-key-32chars"
mah config secrets encrypt

# 4. Commit encrypted secrets (safe!)
git add ~/.mah/secrets.yaml
git commit -m "Add encrypted secrets"

# 5. Team members decrypt with same key
export MAH_MASTER_KEY="shared-team-encryption-key-32chars" 
mah config secrets decrypt
```

## 🛡️ Security Best Practices

### ✅ DO

- **Use environment variables for CI/CD pipelines**
- **Encrypt secrets for team sharing**
- **Use different secrets for different environments**
- **Rotate secrets regularly**
- **Use strong, unique passwords**
- **Add sensitive files to .gitignore**
- **Use SSH keys instead of passwords**
- **Restrict file permissions (600) for secret files**

### ❌ DON'T

- **Never commit plain-text secrets to git**
- **Don't use weak encryption keys**
- **Don't share secrets in chat/email**
- **Don't use the same password everywhere**
- **Don't store secrets in config files without encryption**
- **Don't expose secrets in logs or error messages**

## 🔧 Secret Management Commands

```bash
# Initialize secrets management
mah config secrets init

# Encrypt secrets file
mah config secrets encrypt

# View encrypted secrets (masked)
mah config secrets decrypt

# Create git-safe template
mah config secrets sanitize

# Create config with environment variables
mah config init
```

## 🗂️ File Structure

```
project/
├── mah.template.yaml          # Git-safe config template
├── .gitignore                 # Excludes sensitive files
├── .env                       # Environment variables (ignored)
└── ~/.mah/
    ├── secrets.yaml           # Encrypted secrets (safe to commit)
    └── config.yaml           # Runtime configuration
```

## 🔐 Encryption Details

MAH uses **AES-256-GCM** encryption with:
- **256-bit encryption key** from `MAH_MASTER_KEY`
- **Unique nonce** for each encrypted value
- **Authenticated encryption** preventing tampering
- **Base64 encoding** for safe storage

## 🌍 Environment Variable Patterns

MAH supports flexible environment variable patterns:

```yaml
# Simple substitution
host: "${SERVER_HOST}"

# With defaults
host: "${SERVER_HOST:-localhost}"

# Nested in structures
database:
  password: "${DB_PASSWORD}"
  host: "${DB_HOST:-localhost}"
  
# Plugin configurations
plugins:
  dns:
    config:
      token: "${NAMECOM_TOKEN}"
```

## 🚨 Security Validation

MAH automatically warns about potential security issues:

- ✅ **Detects secrets in main config**
- ✅ **Validates file permissions**
- ✅ **Checks for common secret patterns**
- ✅ **Warns about unencrypted sensitive data**

## 📋 Migration Guide

### From Plain Text Config

1. **Backup your current config**
2. **Run `mah config secrets sanitize`** to create template
3. **Move secrets to environment variables or secrets.yaml**
4. **Test with `mah config validate`**
5. **Commit sanitized version**

### To Environment Variables

```bash
# Extract secrets from your config
grep -E "(password|token|key|secret)" mah.yaml > secrets.txt

# Create .env file
echo "export MYSQL_PASSWORD=your-password" >> .env
echo "export NAMECOM_TOKEN=your-token" >> .env

# Update config to use variables
sed -i 's/password: "actual-password"/password: "${MYSQL_PASSWORD}"/' mah.yaml
```

### To Encrypted Secrets

```bash
# Option 1: Direct password
mah config secrets init -p "your-32-character-key"
vim ~/.mah/secrets.yaml
mah config secrets encrypt -p "your-32-character-key"

# Option 2: Environment variable
export MAH_MASTER_KEY="your-32-character-key"
mah config secrets init --auto-encrypt
vim ~/.mah/secrets.yaml
mah config secrets encrypt
```

Remember: **Security is a journey, not a destination**. Regularly review and update your secret management practices!