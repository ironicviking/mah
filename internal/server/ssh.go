package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/jonas-jonas/mah/internal/config"
	"github.com/jonas-jonas/mah/pkg"
)

// SSHServer implements the pkg.Server interface using SSH
type SSHServer struct {
	config *config.Server
	client *ssh.Client
	conn   net.Conn
	id     string
}

// NewSSHServer creates a new SSH server instance
func NewSSHServer(id string, config *config.Server) *SSHServer {
	return &SSHServer{
		config: config,
		id:     id,
	}
}

// Connect establishes SSH connection to the server
func (s *SSHServer) Connect(ctx context.Context) error {
	if s.client != nil {
		return nil // Already connected
	}

	// Load SSH private key
	keyPath := s.config.SSHKey
	if strings.HasPrefix(keyPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		keyPath = filepath.Join(home, keyPath[2:])
	}

	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read SSH private key from %s: %w", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		// Check if key needs a passphrase
		if strings.Contains(err.Error(), "cannot decode encrypted private keys") {
			return fmt.Errorf("SSH private key requires a passphrase (not yet supported): %w", err)
		}
		return fmt.Errorf("failed to parse SSH private key at %s: %w", keyPath, err)
	}

	// Create SSH client config
	port := s.config.SSHPort
	if port == 0 {
		port = 22
	}

	sshConfig := &ssh.ClientConfig{
		User: s.config.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
		Timeout:         30 * time.Second,
	}

	// Connect
	addr := net.JoinHostPort(s.config.Host, strconv.Itoa(port))
	s.conn, err = net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(s.conn, addr, sshConfig)
	if err != nil {
		s.conn.Close()
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	s.client = ssh.NewClient(sshConn, chans, reqs)
	return nil
}

// Execute runs a command on the remote server
func (s *SSHServer) Execute(ctx context.Context, cmd string, sudo bool) (*pkg.Result, error) {
	if s.client == nil {
		return nil, fmt.Errorf("not connected to server")
	}

	session, err := s.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Prepare command with sudo if needed
	if sudo && s.config.Sudo {
		cmd = fmt.Sprintf("sudo -n %s", cmd)
	}

	start := time.Now()

	// Set up command execution with context
	done := make(chan error, 1)
	var stdout, stderr strings.Builder
	
	session.Stdout = &stdout
	session.Stderr = &stderr

	go func() {
		done <- session.Run(cmd)
	}()

	// Wait for command completion or context cancellation
	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return nil, ctx.Err()
	case err := <-done:
		duration := time.Since(start).Milliseconds()
		
		result := &pkg.Result{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Duration: duration,
		}

		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				result.ExitCode = exitErr.ExitStatus()
			} else {
				result.ExitCode = 1
				if result.Stderr == "" {
					result.Stderr = err.Error()
				}
			}
		} else {
			result.ExitCode = 0
		}

		return result, nil
	}
}

// TransferFile transfers a file to the remote server using SFTP
func (s *SSHServer) TransferFile(ctx context.Context, local, remote string) error {
	if s.client == nil {
		return fmt.Errorf("not connected to server")
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Open local file
	localFile, err := os.Open(local)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", local, err)
	}
	defer localFile.Close()

	// Get local file info for permissions
	localInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}

	// Create remote directory if needed
	remoteDir := filepath.Dir(remote)
	if remoteDir != "." {
		err = sftpClient.MkdirAll(remoteDir)
		if err != nil {
			return fmt.Errorf("failed to create remote directory %s: %w", remoteDir, err)
		}
	}

	// Create remote file
	remoteFile, err := sftpClient.Create(remote)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remote, err)
	}
	defer remoteFile.Close()

	// Copy file contents
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	// Set permissions
	err = sftpClient.Chmod(remote, localInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// GetDistro detects the Linux distribution
func (s *SSHServer) GetDistro(ctx context.Context) (string, error) {
	// Try /etc/os-release first (modern standard)
	result, err := s.Execute(ctx, "cat /etc/os-release", false)
	if err == nil && result.ExitCode == 0 {
		lines := strings.Split(result.Stdout, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ID=") {
				distro := strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
				return strings.ToLower(distro), nil
			}
		}
	}

	// Fallback: Try specific release files
	distroFiles := map[string]string{
		"ubuntu": "/etc/lsb-release",
		"debian": "/etc/debian_version",
		"centos": "/etc/centos-release",
		"rhel":   "/etc/redhat-release",
		"rocky":  "/etc/rocky-release",
		"alpine": "/etc/alpine-release",
	}

	for distro, file := range distroFiles {
		result, err := s.Execute(ctx, fmt.Sprintf("test -f %s", file), false)
		if err == nil && result.ExitCode == 0 {
			return distro, nil
		}
	}

	return "unknown", nil
}

// GetResources gets server resource information
func (s *SSHServer) GetResources(ctx context.Context) (*pkg.ResourceInfo, error) {
	info := &pkg.ResourceInfo{}

	// Get CPU information
	cpuInfo, err := s.getCPUInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}
	info.CPU = cpuInfo

	// Get memory information
	memInfo, err := s.getMemoryInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}
	info.Memory = memInfo

	// Get disk information
	diskInfo, err := s.getDiskInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}
	info.Disk = diskInfo

	// Get load information
	loadInfo, err := s.getLoadInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get load info: %w", err)
	}
	info.Load = loadInfo

	return info, nil
}

// HealthCheck performs a basic health check
func (s *SSHServer) HealthCheck(ctx context.Context) error {
	result, err := s.Execute(ctx, "echo 'health_check'", false)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	
	if result.ExitCode != 0 {
		return fmt.Errorf("health check command failed with exit code %d", result.ExitCode)
	}
	
	if strings.TrimSpace(result.Stdout) != "health_check" {
		return fmt.Errorf("unexpected health check response: %s", result.Stdout)
	}

	return nil
}

// Disconnect closes the SSH connection
func (s *SSHServer) Disconnect() error {
	if s.client != nil {
		err := s.client.Close()
		s.client = nil
		if s.conn != nil {
			s.conn.Close()
			s.conn = nil
		}
		return err
	}
	return nil
}

// ID returns the server identifier
func (s *SSHServer) ID() string {
	return s.id
}

// Host returns the server hostname/IP
func (s *SSHServer) Host() string {
	return s.config.Host
}

// Helper methods for resource gathering

func (s *SSHServer) getCPUInfo(ctx context.Context) (pkg.CPUInfo, error) {
	// Get CPU count
	result, err := s.Execute(ctx, "nproc", false)
	if err != nil {
		return pkg.CPUInfo{}, err
	}
	
	cores, err := strconv.Atoi(strings.TrimSpace(result.Stdout))
	if err != nil {
		cores = 1 // fallback
	}

	// Get CPU usage (1-minute average)
	usage := 0.0
	result, err = s.Execute(ctx, "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'", false)
	if err == nil && result.ExitCode == 0 {
		if u, err := strconv.ParseFloat(strings.TrimSpace(result.Stdout), 64); err == nil {
			usage = u
		}
	}

	// Get CPU model
	model := "Unknown"
	result, err = s.Execute(ctx, "cat /proc/cpuinfo | grep 'model name' | head -1 | cut -d: -f2 | sed 's/^ *//'", false)
	if err == nil && result.ExitCode == 0 && result.Stdout != "" {
		model = strings.TrimSpace(result.Stdout)
	}

	// Get architecture
	arch := "Unknown"
	result, err = s.Execute(ctx, "uname -m", false)
	if err == nil && result.ExitCode == 0 {
		arch = strings.TrimSpace(result.Stdout)
	}

	return pkg.CPUInfo{
		Cores: cores,
		Usage: usage,
		Model: model,
		Arch:  arch,
	}, nil
}

func (s *SSHServer) getMemoryInfo(ctx context.Context) (pkg.MemoryInfo, error) {
	result, err := s.Execute(ctx, "free -b", false)
	if err != nil {
		return pkg.MemoryInfo{}, err
	}

	lines := strings.Split(result.Stdout, "\n")
	if len(lines) < 2 {
		return pkg.MemoryInfo{}, fmt.Errorf("unexpected free command output")
	}

	// Parse memory line
	fields := strings.Fields(lines[1])
	if len(fields) < 7 {
		return pkg.MemoryInfo{}, fmt.Errorf("unexpected free command format")
	}

	total, _ := strconv.ParseInt(fields[1], 10, 64)
	used, _ := strconv.ParseInt(fields[2], 10, 64)
	available, _ := strconv.ParseInt(fields[6], 10, 64)

	usage := 0.0
	if total > 0 {
		usage = float64(used) / float64(total) * 100
	}

	return pkg.MemoryInfo{
		Total:     total,
		Used:      used,
		Available: available,
		Usage:     usage,
	}, nil
}

func (s *SSHServer) getDiskInfo(ctx context.Context) (pkg.DiskInfo, error) {
	result, err := s.Execute(ctx, "df -B1 /", false)
	if err != nil {
		return pkg.DiskInfo{}, err
	}

	lines := strings.Split(result.Stdout, "\n")
	if len(lines) < 2 {
		return pkg.DiskInfo{}, fmt.Errorf("unexpected df command output")
	}

	// Parse disk usage line
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return pkg.DiskInfo{}, fmt.Errorf("unexpected df command format")
	}

	total, _ := strconv.ParseInt(fields[1], 10, 64)
	used, _ := strconv.ParseInt(fields[2], 10, 64)
	available, _ := strconv.ParseInt(fields[3], 10, 64)

	usage := 0.0
	if total > 0 {
		usage = float64(used) / float64(total) * 100
	}

	return pkg.DiskInfo{
		Total:     total,
		Used:      used,
		Available: available,
		Usage:     usage,
	}, nil
}

func (s *SSHServer) getLoadInfo(ctx context.Context) (pkg.LoadInfo, error) {
	result, err := s.Execute(ctx, "cat /proc/loadavg", false)
	if err != nil {
		return pkg.LoadInfo{}, err
	}

	fields := strings.Fields(result.Stdout)
	if len(fields) < 3 {
		return pkg.LoadInfo{}, fmt.Errorf("unexpected loadavg format")
	}

	load1, _ := strconv.ParseFloat(fields[0], 64)
	load5, _ := strconv.ParseFloat(fields[1], 64)
	load15, _ := strconv.ParseFloat(fields[2], 64)

	return pkg.LoadInfo{
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}, nil
}