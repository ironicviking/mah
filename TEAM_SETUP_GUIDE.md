# 🚀 MAH Team Setup Guide - Super Easy!

This guide shows the **easiest way** to set up MAH for team collaboration with encrypted secrets.

## ⚡ Quick Team Setup (30 seconds!)

### Team Lead (One-Time Setup)

```bash
# 1. Generate a secure team key (32+ characters)
TEAM_KEY="MyTeam2024SecureKey1234567890AB"

# 2. Initialize and encrypt in one step!
mah config secrets init -p "$TEAM_KEY"

# 3. Your secrets are now encrypted and ready!
git add ~/.mah/secrets.yaml
git commit -m "Add encrypted team secrets"
git push
```

**That's it!** Your encrypted secrets are now safely in git.

### Team Members (30 seconds each)

```bash
# 1. Clone the repo
git clone https://github.com/yourteam/infrastructure.git
cd infrastructure

# 2. Use the shared team key to decrypt
mah config secrets decrypt -p "MyTeam2024SecureKey1234567890AB"

# 3. Ready to deploy!
mah config validate
mah nexus list
```

## 🔐 Secure Team Key Generation

Generate a strong team key:

```bash
# Option 1: Random secure key
TEAM_KEY=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
echo "Your team key: $TEAM_KEY"

# Option 2: Memorable but secure (recommended)
TEAM_KEY="YourCompany2024InfraKey$(date +%m%d)"
echo "Team key: $TEAM_KEY"

# Option 3: Use a password manager to generate 32+ character key
```

## 📝 Complete Team Workflow Example

### Initial Setup by Team Lead

```bash
# Create infrastructure repository
mkdir company-infrastructure
cd company-infrastructure
git init

# Initialize MAH with encrypted secrets
TEAM_KEY="CompanyInfra2024SecureKey123456"
mah config secrets init -p "$TEAM_KEY"

# Edit the encrypted secrets with real values
vim ~/.mah/secrets.yaml
# Change the placeholder values to:
# NAMECOM_USERNAME: actual-username  
# NAMECOM_TOKEN: actual-api-token
# MYSQL_PASSWORD: actual-database-password
# etc.

# Re-encrypt with actual values
mah config secrets encrypt -p "$TEAM_KEY"

# Create main configuration
mah config init

# Commit everything (encrypted secrets are safe!)
git add .
git commit -m "Initial MAH infrastructure setup"
git push -u origin main

# Share the team key securely with team members
echo "Team key: $TEAM_KEY" | gpg -e -r team@company.com
```

### Team Member Onboarding

```bash
# 1. Clone and setup
git clone https://github.com/company/infrastructure.git
cd infrastructure

# 2. Get team key (from secure channel: 1Password, Vault, etc.)
TEAM_KEY="CompanyInfra2024SecureKey123456"

# 3. Verify secrets work
mah config secrets decrypt -p "$TEAM_KEY"

# 4. Validate configuration
mah config validate

# 5. Deploy something!
mah nexus switch production
mah service deploy blog
```

## 🔄 Daily Team Operations

### Adding New Secrets

```bash
# Any team member can add secrets
vim ~/.mah/secrets.yaml
# Add: NEW_API_KEY: your-new-api-key

# Re-encrypt and commit
mah config secrets encrypt -p "$TEAM_KEY"
git add ~/.mah/secrets.yaml
git commit -m "Add new API key"
git push
```

### Using Secrets in Configuration

```yaml
# In mah.yaml - always use environment variables
plugins:
  new_service:
    provider: "some-provider"
    config:
      api_key: "${NEW_API_KEY}"  # References secret
```

### Team Member Updates

```bash
# Pull latest changes
git pull

# Decrypt with team key to get new secrets
mah config secrets decrypt -p "$TEAM_KEY"

# Deploy updated configuration
mah config validate
mah service deploy new_service
```

## 🛡️ Security Best Practices

### ✅ DO
- **Use 32+ character team keys**
- **Store team key in secure password manager**
- **Rotate team keys periodically**
- **Use different keys for different environments**
- **Commit encrypted secrets to git**
- **Share keys through secure channels only**

### ❌ DON'T
- **Share keys in chat/email/Slack**
- **Use weak/short keys**
- **Commit unencrypted secrets**
- **Store keys in code or configs**
- **Use the same key everywhere**

## 🔧 Troubleshooting

### "Failed to decrypt"
```bash
# Verify you have the correct team key
mah config secrets decrypt -p "your-team-key"

# If wrong key, get correct one from team lead
```

### "No secrets found"
```bash
# Ensure secrets file exists
ls ~/.mah/secrets.yaml

# Re-initialize if needed
mah config secrets init -p "your-team-key"
```

### "Encryption key required"
```bash
# Always provide the key with -p flag
mah config secrets encrypt -p "your-team-key"

# Or set environment variable
export MAH_MASTER_KEY="your-team-key"
mah config secrets encrypt
```

## 🚀 Alternative: Environment Variable Workflow

Some teams prefer environment variables:

```bash
# Team lead shares this with everyone
export MAH_MASTER_KEY="CompanyInfra2024SecureKey123456"

# Then use without -p flags
mah config secrets init --auto-encrypt
mah config secrets encrypt  
mah config secrets decrypt
```

## 🎯 Why This Approach Rocks

- **🔐 Secure**: AES-256 encryption with team keys
- **⚡ Fast**: One command setup for team members  
- **🤝 Collaborative**: Everyone uses same encrypted secrets
- **📦 Portable**: Everything in git, works everywhere
- **🛡️ Safe**: Encrypted secrets are safe to commit
- **📖 Simple**: No complex secret management infrastructure

**Your infrastructure secrets are now team-ready and git-safe!** 🎉