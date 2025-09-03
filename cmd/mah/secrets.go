package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/jonas-jonas/mah/internal/config"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage encrypted secrets",
	Long: `Secrets commands allow you to securely manage sensitive configuration data.
You can encrypt secrets, store them in a separate file, and safely commit to git.`,
}

var secretsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize secrets management",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get MAH directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		
		mahDir := filepath.Join(homeDir, ".mah")
		
		// Create secret manager
		secretManager, err := config.NewSecretManager(mahDir)
		if err != nil {
			return fmt.Errorf("failed to create secret manager: %w", err)
		}
		
		// Create template
		if err := secretManager.CreateSecretsTemplate(); err != nil {
			return fmt.Errorf("failed to create secrets template: %w", err)
		}
		
		secretsFile := filepath.Join(mahDir, "secrets.yaml")
		fmt.Printf("%s Created secrets template: %s\n", 
			color.GreenString("✓"), 
			color.CyanString(secretsFile))
		fmt.Println()
		fmt.Printf("%s %s\n", color.YellowString("⚠️"), "Security recommendations:")
		fmt.Println("  1. Edit secrets.yaml with your actual values")
		fmt.Println("  2. Add secrets.yaml to .gitignore (recommended)")
		fmt.Println("  3. Or encrypt with: mah config secrets encrypt")
		fmt.Println("  4. Use environment variables in CI/CD pipelines")
		
		return nil
	},
}

var secretsEncryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt secrets in secrets.yaml",
	Long: `Encrypt all secrets in the secrets.yaml file using AES-256 encryption.
The encrypted file can be safely committed to version control.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get MAH directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		
		mahDir := filepath.Join(homeDir, ".mah")
		secretsFile := filepath.Join(mahDir, "secrets.yaml")
		
		// Check if secrets file exists
		if _, err := os.Stat(secretsFile); os.IsNotExist(err) {
			return fmt.Errorf("secrets file not found. Run 'mah config secrets init' first")
		}
		
		// Create secret manager
		secretManager, err := config.NewSecretManager(mahDir)
		if err != nil {
			return fmt.Errorf("failed to create secret manager: %w", err)
		}
		
		// Load current secrets
		secrets, err := secretManager.LoadSecrets()
		if err != nil {
			return fmt.Errorf("failed to load secrets: %w", err)
		}
		
		if len(secrets) == 0 {
			return fmt.Errorf("no secrets found to encrypt")
		}
		
		// Ask for key source
		keySource := "env" // Default to environment variable
		fmt.Printf("%s Encrypting %d secrets...\n", color.CyanString("🔐"), len(secrets))
		fmt.Println("Using MAH_MASTER_KEY environment variable for encryption key")
		fmt.Println("Set MAH_MASTER_KEY before running this command")
		
		// Save encrypted secrets
		if err := secretManager.SaveSecrets(secrets, true, keySource); err != nil {
			return fmt.Errorf("failed to save encrypted secrets: %w", err)
		}
		
		fmt.Printf("%s Secrets encrypted successfully\n", color.GreenString("✓"))
		fmt.Printf("Encrypted file: %s\n", color.CyanString(secretsFile))
		fmt.Println()
		fmt.Printf("%s This file can now be safely committed to git\n", color.GreenString("✓"))
		
		return nil
	},
}

var secretsDecryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt and show secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get MAH directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		
		mahDir := filepath.Join(homeDir, ".mah")
		
		// Create secret manager
		secretManager, err := config.NewSecretManager(mahDir)
		if err != nil {
			return fmt.Errorf("failed to create secret manager: %w", err)
		}
		
		// Load secrets
		secrets, err := secretManager.LoadSecrets()
		if err != nil {
			return fmt.Errorf("failed to load secrets: %w", err)
		}
		
		if len(secrets) == 0 {
			fmt.Println("No secrets found")
			return nil
		}
		
		fmt.Printf("%s Decrypted secrets:\n", color.CyanString("🔓"))
		for key, value := range secrets {
			// Mask the value for security
			masked := value
			if len(value) > 8 {
				masked = value[:4] + "****" + value[len(value)-4:]
			} else if len(value) > 4 {
				masked = value[:2] + "****"
			} else {
				masked = "****"
			}
			fmt.Printf("  %s: %s\n", key, masked)
		}
		
		return nil
	},
}

var secretsSanitizeCmd = &cobra.Command{
	Use:   "sanitize [input-file] [output-file]",
	Short: "Create a git-safe version of config by removing sensitive data",
	Long: `Create a sanitized version of your configuration file that's safe to commit to git.
This replaces sensitive values with environment variable placeholders.

Examples:
  mah config secrets sanitize mah.yaml mah.template.yaml
  mah config secrets sanitize  # Uses mah.yaml -> mah.template.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := "mah.yaml"
		outputFile := "mah.template.yaml"
		
		if len(args) >= 1 {
			inputFile = args[0]
		}
		if len(args) >= 2 {
			outputFile = args[1]
		}
		
		// Check if input file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("input file not found: %s", inputFile)
		}
		
		// Sanitize config
		if err := config.SanitizeConfigForGit(inputFile, outputFile); err != nil {
			return fmt.Errorf("failed to sanitize config: %w", err)
		}
		
		fmt.Printf("%s Created sanitized config: %s\n", 
			color.GreenString("✓"), 
			color.CyanString(outputFile))
		fmt.Println("This file is safe to commit to version control")
		
		// Add to .gitignore if not already there
		gitignorePath := ".gitignore"
		gitignoreContent := ""
		
		if data, err := os.ReadFile(gitignorePath); err == nil {
			gitignoreContent = string(data)
		}
		
		if gitignoreContent != "" && !contains(gitignoreContent, "mah.yaml") {
			gitignoreContent += "\n# MAH Configuration (contains secrets)\nmah.yaml\n.mah/secrets.yaml\n"
			if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err == nil {
				fmt.Printf("%s Updated .gitignore to exclude sensitive files\n", color.GreenString("✓"))
			}
		}
		
		return nil
	},
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		len(s) > len(substr) && 
		(s[len(s)-len(substr)-1:len(s)-len(substr)] == "\n" || 
		 s[len(s)-len(substr)-1:len(s)-len(substr)] == " ") &&
		s[len(s)-len(substr):] == substr)
}

func init() {
	secretsCmd.AddCommand(secretsInitCmd)
	secretsCmd.AddCommand(secretsEncryptCmd)
	secretsCmd.AddCommand(secretsDecryptCmd)
	secretsCmd.AddCommand(secretsSanitizeCmd)
	
	// Add secrets as a subcommand of config
	configCmd.AddCommand(secretsCmd)
}