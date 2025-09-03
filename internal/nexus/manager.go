package nexus

import (
	"context"
	"fmt"
	"sync"

	"github.com/jonas-jonas/mah/internal/config"
	"github.com/jonas-jonas/mah/pkg"
)

// Manager handles nexus operations and management
type Manager struct {
	configMgr    *config.Manager
	servers      map[string]pkg.Server
	currentNexus string
	mu           sync.RWMutex
}

// NewManager creates a new nexus manager
func NewManager(configMgr *config.Manager) *Manager {
	return &Manager{
		configMgr: configMgr,
		servers:   make(map[string]pkg.Server),
	}
}

// Nexus represents a nexus with its configuration and servers
type Nexus struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Environment string              `json:"environment"`
	Servers     []pkg.Server        `json:"servers"`
	Config      *config.Nexus       `json:"config"`
	Status      *Status             `json:"status"`
}

// Status represents the status of a nexus
type Status struct {
	Healthy        bool                      `json:"healthy"`
	ServersOnline  int                       `json:"servers_online"`
	ServersTotal   int                       `json:"servers_total"`
	ServicesTotal  int                       `json:"services_total"`
	ServerStatuses map[string]*ServerStatus  `json:"server_statuses"`
}

// ServerStatus represents the status of a server in a nexus
type ServerStatus struct {
	Online    bool                 `json:"online"`
	Resources *pkg.ResourceInfo    `json:"resources"`
	Services  []*pkg.ServiceStatus `json:"services"`
	Error     string               `json:"error,omitempty"`
}

// List returns all configured nexuses
func (m *Manager) List() ([]*Nexus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg := m.configMgr.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("no configuration loaded")
	}

	var nexuses []*Nexus
	for name, nexusConfig := range cfg.Nexuses {
		nexus := &Nexus{
			Name:        name,
			Description: nexusConfig.Description,
			Environment: nexusConfig.Environment,
			Config:      nexusConfig,
		}

		// Load servers for this nexus
		servers, err := m.getServersForNexus(name)
		if err != nil {
			return nil, fmt.Errorf("failed to load servers for nexus '%s': %w", name, err)
		}
		nexus.Servers = servers

		nexuses = append(nexuses, nexus)
	}

	return nexuses, nil
}

// Get returns a specific nexus by name
func (m *Manager) Get(name string) (*Nexus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg := m.configMgr.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("no configuration loaded")
	}

	nexusConfig := cfg.Nexuses[name]
	if nexusConfig == nil {
		return nil, fmt.Errorf("nexus '%s' not found", name)
	}

	nexus := &Nexus{
		Name:        name,
		Description: nexusConfig.Description,
		Environment: nexusConfig.Environment,
		Config:      nexusConfig,
	}

	// Load servers for this nexus
	servers, err := m.getServersForNexus(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load servers for nexus '%s': %w", name, err)
	}
	nexus.Servers = servers

	return nexus, nil
}

// GetCurrent returns the currently active nexus
func (m *Manager) GetCurrent() (*Nexus, error) {
	currentNexusName := m.configMgr.GetCurrentNexus()
	if currentNexusName == "" {
		return nil, fmt.Errorf("no current nexus set")
	}

	return m.Get(currentNexusName)
}

// Switch switches to a different nexus
func (m *Manager) Switch(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate nexus exists
	_, err := m.Get(name)
	if err != nil {
		return fmt.Errorf("cannot switch to nexus: %w", err)
	}

	// Update current nexus
	if err := m.configMgr.SetCurrentNexus(name); err != nil {
		return fmt.Errorf("failed to set current nexus: %w", err)
	}

	m.currentNexus = name
	return nil
}

// Status returns the status of a nexus
func (m *Manager) Status(ctx context.Context, name string) (*Status, error) {
	nexus, err := m.Get(name)
	if err != nil {
		return nil, err
	}

	status := &Status{
		ServersTotal:   len(nexus.Servers),
		ServerStatuses: make(map[string]*ServerStatus),
	}

	// Check status of each server
	var wg sync.WaitGroup
	statusChan := make(chan struct {
		serverID string
		status   *ServerStatus
	}, len(nexus.Servers))

	for _, server := range nexus.Servers {
		wg.Add(1)
		go func(srv pkg.Server) {
			defer wg.Done()

			serverStatus := &ServerStatus{}

			// Check if server is online
			if err := srv.HealthCheck(ctx); err != nil {
				serverStatus.Online = false
				serverStatus.Error = err.Error()
			} else {
				serverStatus.Online = true
				status.ServersOnline++

				// Get resource information
				if resources, err := srv.GetResources(ctx); err == nil {
					serverStatus.Resources = resources
				}
			}

			statusChan <- struct {
				serverID string
				status   *ServerStatus
			}{srv.ID(), serverStatus}
		}(server)
	}

	go func() {
		wg.Wait()
		close(statusChan)
	}()

	// Collect results
	for result := range statusChan {
		status.ServerStatuses[result.serverID] = result.status
	}

	status.Healthy = status.ServersOnline == status.ServersTotal

	return status, nil
}

// ExecuteOnAll executes a command on all servers in the current nexus
func (m *Manager) ExecuteOnAll(ctx context.Context, cmd string, sudo bool) (map[string]*pkg.Result, error) {
	currentNexus, err := m.GetCurrent()
	if err != nil {
		return nil, fmt.Errorf("no current nexus: %w", err)
	}

	return m.ExecuteOnNexus(ctx, currentNexus.Name, cmd, sudo)
}

// ExecuteOnNexus executes a command on all servers in a specific nexus
func (m *Manager) ExecuteOnNexus(ctx context.Context, nexusName, cmd string, sudo bool) (map[string]*pkg.Result, error) {
	nexus, err := m.Get(nexusName)
	if err != nil {
		return nil, err
	}

	results := make(map[string]*pkg.Result)
	var wg sync.WaitGroup
	resultChan := make(chan struct {
		serverID string
		result   *pkg.Result
		error    error
	}, len(nexus.Servers))

	// Execute command on each server in parallel
	for _, server := range nexus.Servers {
		wg.Add(1)
		go func(srv pkg.Server) {
			defer wg.Done()

			result, err := srv.Execute(ctx, cmd, sudo)
			resultChan <- struct {
				serverID string
				result   *pkg.Result
				error    error
			}{srv.ID(), result, err}
		}(server)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var lastError error
	for result := range resultChan {
		if result.error != nil {
			lastError = result.error
			// Create error result
			results[result.serverID] = &pkg.Result{
				ExitCode: -1,
				Stderr:   result.error.Error(),
			}
		} else {
			results[result.serverID] = result.result
		}
	}

	if lastError != nil {
		return results, fmt.Errorf("command execution failed on some servers: %w", lastError)
	}

	return results, nil
}

// getServersForNexus loads server instances for a nexus
func (m *Manager) getServersForNexus(nexusName string) ([]pkg.Server, error) {
	serverConfigs, err := m.configMgr.GetNexusServers(nexusName)
	if err != nil {
		return nil, err
	}

	var servers []pkg.Server
	for _, serverConfig := range serverConfigs {
		// Create server instance (this will be implemented when we have server package)
		// For now, we'll just create a placeholder
		servers = append(servers, &mockServer{
			id:   fmt.Sprintf("%s-%s", serverConfig.Host, serverConfig.SSHUser),
			host: serverConfig.Host,
		})
	}

	return servers, nil
}

// mockServer is a temporary placeholder until we implement the real server package
type mockServer struct {
	id   string
	host string
}

func (s *mockServer) Connect(ctx context.Context) error         { return nil }
func (s *mockServer) Execute(ctx context.Context, cmd string, sudo bool) (*pkg.Result, error) {
	return &pkg.Result{ExitCode: 0, Stdout: "mock output"}, nil
}
func (s *mockServer) TransferFile(ctx context.Context, local, remote string) error { return nil }
func (s *mockServer) Disconnect() error                        { return nil }
func (s *mockServer) GetDistro(ctx context.Context) (string, error) { return "ubuntu", nil }
func (s *mockServer) GetResources(ctx context.Context) (*pkg.ResourceInfo, error) {
	return &pkg.ResourceInfo{
		CPU:    pkg.CPUInfo{Cores: 4, Usage: 25.0},
		Memory: pkg.MemoryInfo{Total: 8000000000, Used: 2000000000, Usage: 25.0},
	}, nil
}
func (s *mockServer) HealthCheck(ctx context.Context) error { return nil }
func (s *mockServer) ID() string                           { return s.id }
func (s *mockServer) Host() string                         { return s.host }