package docker

import (
	"fmt"
	"strings"
)

// ComposeFile represents a docker-compose.yml file
type ComposeFile struct {
	Version  string                     `yaml:"version"`
	Services map[string]ComposeService  `yaml:"services"`
	Networks map[string]ComposeNetwork  `yaml:"networks,omitempty"`
	Volumes  map[string]ComposeVolume   `yaml:"volumes,omitempty"`
}

// ComposeService represents a service in docker-compose
type ComposeService struct {
	Image       string            `yaml:"image"`
	Ports       []string          `yaml:"ports,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Networks    []string          `yaml:"networks,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Restart     string            `yaml:"restart,omitempty"`
	Command     []string          `yaml:"command,omitempty"`
	HealthCheck *HealthCheck      `yaml:"healthcheck,omitempty"`
}

// ComposeNetwork represents a network in docker-compose
type ComposeNetwork struct {
	Driver   string            `yaml:"driver,omitempty"`
	External bool              `yaml:"external,omitempty"`
	Labels   map[string]string `yaml:"labels,omitempty"`
}

// ComposeVolume represents a volume in docker-compose
type ComposeVolume struct {
	Driver   string            `yaml:"driver,omitempty"`
	External bool              `yaml:"external,omitempty"`
	Labels   map[string]string `yaml:"labels,omitempty"`
}

// HealthCheck represents a health check configuration
type HealthCheck struct {
	Test        []string `yaml:"test,omitempty"`
	Interval    string   `yaml:"interval,omitempty"`
	Timeout     string   `yaml:"timeout,omitempty"`
	Retries     int      `yaml:"retries,omitempty"`
	StartPeriod string   `yaml:"start_period,omitempty"`
}

// ToYAML converts the compose file to YAML format
func (c *ComposeFile) ToYAML() string {
	var sb strings.Builder

	// Note: version field is obsolete in modern Docker Compose, omitting it

	// Services section
	if len(c.Services) > 0 {
		sb.WriteString("services:\n")
		for name, service := range c.Services {
			sb.WriteString(fmt.Sprintf("  %s:\n", name))
			sb.WriteString(fmt.Sprintf("    image: %s\n", service.Image))

			if service.Restart != "" {
				sb.WriteString(fmt.Sprintf("    restart: %s\n", service.Restart))
			}

			if len(service.Ports) > 0 {
				sb.WriteString("    ports:\n")
				for _, port := range service.Ports {
					sb.WriteString(fmt.Sprintf("      - \"%s\"\n", port))
				}
			}

			if len(service.Environment) > 0 {
				sb.WriteString("    environment:\n")
				for key, value := range service.Environment {
					// Handle environment variable references
					if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
						sb.WriteString(fmt.Sprintf("      %s: %s\n", key, value))
					} else {
						sb.WriteString(fmt.Sprintf("      %s: \"%s\"\n", key, value))
					}
				}
			}

			if len(service.Volumes) > 0 {
				sb.WriteString("    volumes:\n")
				for _, volume := range service.Volumes {
					sb.WriteString(fmt.Sprintf("      - %s\n", volume))
				}
			}

			if len(service.Networks) > 0 {
				sb.WriteString("    networks:\n")
				for _, network := range service.Networks {
					sb.WriteString(fmt.Sprintf("      - %s\n", network))
				}
			}

			if len(service.DependsOn) > 0 {
				sb.WriteString("    depends_on:\n")
				for _, dep := range service.DependsOn {
					sb.WriteString(fmt.Sprintf("      - %s\n", dep))
				}
			}

			if len(service.Labels) > 0 {
				sb.WriteString("    labels:\n")
				for key, value := range service.Labels {
					sb.WriteString(fmt.Sprintf("      %s: \"%s\"\n", key, value))
				}
			}

			if len(service.Command) > 0 {
				sb.WriteString("    command:\n")
				for _, cmd := range service.Command {
					sb.WriteString(fmt.Sprintf("      - \"%s\"\n", cmd))
				}
			}

			if service.HealthCheck != nil {
				sb.WriteString("    healthcheck:\n")
				if len(service.HealthCheck.Test) > 0 {
					sb.WriteString("      test: [")
					for i, test := range service.HealthCheck.Test {
						if i > 0 {
							sb.WriteString(", ")
						}
						sb.WriteString(fmt.Sprintf("\"%s\"", test))
					}
					sb.WriteString("]\n")
				}
				if service.HealthCheck.Interval != "" {
					sb.WriteString(fmt.Sprintf("      interval: %s\n", service.HealthCheck.Interval))
				}
				if service.HealthCheck.Timeout != "" {
					sb.WriteString(fmt.Sprintf("      timeout: %s\n", service.HealthCheck.Timeout))
				}
				if service.HealthCheck.Retries > 0 {
					sb.WriteString(fmt.Sprintf("      retries: %d\n", service.HealthCheck.Retries))
				}
				if service.HealthCheck.StartPeriod != "" {
					sb.WriteString(fmt.Sprintf("      start_period: %s\n", service.HealthCheck.StartPeriod))
				}
			}

			sb.WriteString("\n")
		}
	}

	// Networks section
	if len(c.Networks) > 0 {
		sb.WriteString("networks:\n")
		for name, network := range c.Networks {
			sb.WriteString(fmt.Sprintf("  %s:\n", name))
			if network.External {
				sb.WriteString("    external: true\n")
			} else if network.Driver != "" {
				sb.WriteString(fmt.Sprintf("    driver: %s\n", network.Driver))
			}
			if len(network.Labels) > 0 {
				sb.WriteString("    labels:\n")
				for key, value := range network.Labels {
					sb.WriteString(fmt.Sprintf("      %s: \"%s\"\n", key, value))
				}
			}
		}
		sb.WriteString("\n")
	}

	// Volumes section
	if len(c.Volumes) > 0 {
		sb.WriteString("volumes:\n")
		for name, volume := range c.Volumes {
			sb.WriteString(fmt.Sprintf("  %s:\n", name))
			if volume.External {
				sb.WriteString("    external: true\n")
			} else if volume.Driver != "" {
				sb.WriteString(fmt.Sprintf("    driver: %s\n", volume.Driver))
			}
			if len(volume.Labels) > 0 {
				sb.WriteString("    labels:\n")
				for key, value := range volume.Labels {
					sb.WriteString(fmt.Sprintf("      %s: \"%s\"\n", key, value))
				}
			}
		}
	}

	return sb.String()
}

// AddTraefikLabels adds Traefik labels for automatic reverse proxy configuration
func (s *ComposeService) AddTraefikLabels(serviceName string, domain string, port int, internal bool) {
	if s.Labels == nil {
		s.Labels = make(map[string]string)
	}

	// Enable Traefik
	s.Labels["traefik.enable"] = "true"

	// Set the service
	s.Labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", serviceName)] = fmt.Sprintf("%d", port)

	if !internal && domain != "" {
		// HTTP router
		s.Labels[fmt.Sprintf("traefik.http.routers.%s.rule", serviceName)] = fmt.Sprintf("Host(`%s`)", domain)
		s.Labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", serviceName)] = "web"

		// HTTPS router
		s.Labels[fmt.Sprintf("traefik.http.routers.%s-secure.rule", serviceName)] = fmt.Sprintf("Host(`%s`)", domain)
		s.Labels[fmt.Sprintf("traefik.http.routers.%s-secure.entrypoints", serviceName)] = "websecure"
		s.Labels[fmt.Sprintf("traefik.http.routers.%s-secure.tls", serviceName)] = "true"
		s.Labels[fmt.Sprintf("traefik.http.routers.%s-secure.tls.certresolver", serviceName)] = "letsencrypt"

		// Redirect HTTP to HTTPS
		s.Labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", serviceName)] = fmt.Sprintf("%s-redirect", serviceName)
		s.Labels[fmt.Sprintf("traefik.http.middlewares.%s-redirect.redirectscheme.scheme", serviceName)] = "https"
		s.Labels[fmt.Sprintf("traefik.http.middlewares.%s-redirect.redirectscheme.permanent", serviceName)] = "true"
	}
}

// SetHealthCheck sets a health check for the service
func (s *ComposeService) SetHealthCheck(test []string, interval, timeout string, retries int) {
	s.HealthCheck = &HealthCheck{
		Test:     test,
		Interval: interval,
		Timeout:  timeout,
		Retries:  retries,
	}
}