package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// SecretManager handles encrypted secrets and secure configuration
type SecretManager struct {
	secretsFile string
	masterKey   []byte
	gcm         cipher.AEAD
}

// SecretConfig represents encrypted secrets
type SecretConfig struct {
	Secrets    map[string]string `yaml:"secrets"`
	Encrypted  bool              `yaml:"encrypted"`
	KeySource  string            `yaml:"key_source"` // env, file, prompt
}

// NewSecretManager creates a new secret manager
func NewSecretManager(mahDir string) (*SecretManager, error) {
	secretsFile := filepath.Join(mahDir, "secrets.yaml")
	
	sm := &SecretManager{
		secretsFile: secretsFile,
	}
	
	return sm, nil
}

// InitializeEncryption sets up encryption with a master key
func (sm *SecretManager) InitializeEncryption(keySource string) error {
	var key []byte
	var err error
	
	switch keySource {
	case "env":
		keyEnv := os.Getenv("MAH_MASTER_KEY")
		if keyEnv == "" {
			return fmt.Errorf("MAH_MASTER_KEY environment variable not set")
		}
		key = []byte(keyEnv)
	case "prompt":
		fmt.Print("Enter master key for secrets encryption: ")
		keyBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // Add newline after password input
		key = keyBytes
	case "file":
		keyFile := filepath.Join(filepath.Dir(sm.secretsFile), ".mah-key")
		key, err = os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}
	default:
		return fmt.Errorf("unknown key source: %s", keySource)
	}
	
	// Ensure key is 32 bytes for AES-256
	if len(key) < 32 {
		// Pad with zeros (in production, use proper key derivation)
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		key = key[:32]
	}
	
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}
	
	sm.masterKey = key
	sm.gcm = gcm
	
	return nil
}

// EncryptSecret encrypts a secret value
func (sm *SecretManager) EncryptSecret(plaintext string) (string, error) {
	if sm.gcm == nil {
		return "", fmt.Errorf("encryption not initialized")
	}
	
	// Generate nonce
	nonce := make([]byte, sm.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	// Encrypt
	ciphertext := sm.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// Base64 encode
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts a secret value
func (sm *SecretManager) DecryptSecret(ciphertext string) (string, error) {
	if sm.gcm == nil {
		return "", fmt.Errorf("encryption not initialized")
	}
	
	// Base64 decode
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	
	// Extract nonce
	nonceSize := sm.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	
	nonce := data[:nonceSize]
	cipherData := data[nonceSize:]
	
	// Decrypt
	plaintext, err := sm.gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	
	return string(plaintext), nil
}

// LoadSecrets loads secrets from the secrets file
func (sm *SecretManager) LoadSecrets() (map[string]string, error) {
	if _, err := os.Stat(sm.secretsFile); os.IsNotExist(err) {
		return make(map[string]string), nil
	}
	
	data, err := os.ReadFile(sm.secretsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}
	
	var secretConfig SecretConfig
	if err := yaml.Unmarshal(data, &secretConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets: %w", err)
	}
	
	secrets := make(map[string]string)
	
	if secretConfig.Encrypted {
		// Initialize encryption if not done
		if sm.gcm == nil {
			if err := sm.InitializeEncryption(secretConfig.KeySource); err != nil {
				return nil, fmt.Errorf("failed to initialize encryption: %w", err)
			}
		}
		
		// Decrypt all secrets
		for key, encryptedValue := range secretConfig.Secrets {
			decrypted, err := sm.DecryptSecret(encryptedValue)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt secret '%s': %w", key, err)
			}
			secrets[key] = decrypted
		}
	} else {
		// Plain text secrets (for development)
		secrets = secretConfig.Secrets
	}
	
	return secrets, nil
}

// SaveSecrets saves secrets to the secrets file
func (sm *SecretManager) SaveSecrets(secrets map[string]string, encrypt bool, keySource string) error {
	secretConfig := SecretConfig{
		Secrets:   make(map[string]string),
		Encrypted: encrypt,
		KeySource: keySource,
	}
	
	if encrypt {
		// Initialize encryption if not done
		if sm.gcm == nil {
			if err := sm.InitializeEncryption(keySource); err != nil {
				return fmt.Errorf("failed to initialize encryption: %w", err)
			}
		}
		
		// Encrypt all secrets
		for key, value := range secrets {
			encrypted, err := sm.EncryptSecret(value)
			if err != nil {
				return fmt.Errorf("failed to encrypt secret '%s': %w", key, err)
			}
			secretConfig.Secrets[key] = encrypted
		}
	} else {
		secretConfig.Secrets = secrets
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(&secretConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}
	
	// Write to file with restricted permissions
	return os.WriteFile(sm.secretsFile, data, 0600)
}

// CreateSecretsTemplate creates a template secrets file
func (sm *SecretManager) CreateSecretsTemplate() error {
	template := SecretConfig{
		Secrets: map[string]string{
			"NAMECOM_USERNAME":     "your-name.com-username",
			"NAMECOM_TOKEN":        "your-name.com-api-token",
			"MYSQL_PASSWORD":       "secure-database-password",
			"CLOUDFLARE_API_TOKEN": "your-cloudflare-api-token",
		},
		Encrypted: false,
		KeySource: "env", // env, file, prompt
	}
	
	data, err := yaml.Marshal(&template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}
	
	header := `# MAH Secrets Configuration
# 
# This file contains sensitive configuration values.
# 
# Security Options:
# 1. Environment Variables (recommended for CI/CD):
#    - Set encrypted: false
#    - Set actual values as environment variables
#    - Add this file to .gitignore
# 
# 2. Encrypted Secrets (recommended for teams):
#    - Set encrypted: true
#    - Use 'mah config encrypt' to encrypt values
#    - Safe to commit encrypted version to git
# 
# 3. External Secret Management:
#    - Use ${ENV_VAR} syntax in main config
#    - Manage secrets in HashiCorp Vault, AWS Secrets Manager, etc.
#
# Key Sources for encryption:
# - env: Use MAH_MASTER_KEY environment variable
# - file: Use .mah-key file (add to .gitignore)
# - prompt: Interactive password prompt (future)

`
	
	fullData := header + string(data)
	
	return os.WriteFile(sm.secretsFile, []byte(fullData), 0600)
}

// SanitizeConfigForGit removes sensitive values from config for git commits
func SanitizeConfigForGit(configPath string, outputPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	
	configStr := string(data)
	
	// Replace sensitive patterns with placeholders
	sensitivePatterns := map[string]string{
		`host:\s*"([^"]+)"`:           `host: "YOUR_SERVER_IP"`,
		`ssh_key:\s*"([^"]+)"`:       `ssh_key: "~/.ssh/your_key"`,
		`username:\s*"([^"]+)"`:      `username: "${NAMECOM_USERNAME}"`,
		`token:\s*"([^"]+)"`:         `token: "${NAMECOM_TOKEN}"`,
		`password:\s*"([^"]+)"`:      `password: "${DB_PASSWORD}"`,
		`api_key:\s*"([^"]+)"`:       `api_key: "${API_KEY}"`,
		`secret:\s*"([^"]+)"`:        `secret: "${SECRET}"`,
		`email:\s*"([^@]+@[^"]+)"`:   `email: "admin@example.com"`,
	}
	
	for pattern, replacement := range sensitivePatterns {
		re := regexp.MustCompile(pattern)
		configStr = re.ReplaceAllString(configStr, replacement)
	}
	
	// Add warning header
	header := `# MAH Configuration Template
# 
# This is a sanitized version of the configuration for version control.
# Copy this to mah.yaml and fill in your actual values.
# 
# For sensitive values, either:
# 1. Use environment variables: ${ENV_VAR_NAME}
# 2. Store in encrypted secrets.yaml file
# 3. Use external secret management
#
# DO NOT commit actual credentials to version control!

`
	
	sanitized := header + configStr
	
	return os.WriteFile(outputPath, []byte(sanitized), 0644)
}

// ValidateSecretsSetup checks if secrets are properly configured
func (sm *SecretManager) ValidateSecretsSetup(config *Config) error {
	// Check for obvious secrets in main config
	configData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	
	configStr := strings.ToLower(string(configData))
	
	// Patterns that suggest secrets are in the main config
	suspiciousPatterns := []string{
		"password:",
		"secret:",
		"api_key:",
		"token:",
		"key:",
	}
	
	var warnings []string
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(configStr, pattern) {
			// Check if it's using environment variable syntax
			if !strings.Contains(configStr, "${") {
				warnings = append(warnings, fmt.Sprintf("Found potential secret: %s", pattern))
			}
		}
	}
	
	if len(warnings) > 0 {
		fmt.Println("⚠️  Security Warning: Potential secrets detected in main config:")
		for _, warning := range warnings {
			fmt.Printf("   - %s\n", warning)
		}
		fmt.Println("   Consider using environment variables or encrypted secrets.")
	}
	
	return nil
}