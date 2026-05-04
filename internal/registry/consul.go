package registry

import (
	"context"
	"fmt"
	"time"

	consulAPI "github.com/hashicorp/consul/api"
)

// ConsulConfig contains Consul-specific configuration.
type ConsulConfig struct {
	Address       string
	TLSEnabled    bool
	CACert        string
	ClientCert    string
	ClientKey     string
	Token         string
	Datacenter    string
	ServiceName   string
	ServiceID     string
	ServiceTags   []string
	ServicePort   int
	TTL           time.Duration
	UnhealthyAfter time.Duration
}

// ConsulRegistry provides Consul-based service discovery.
type ConsulRegistry struct {
	client *consulAPI.Client
	config *ConsulConfig
}

// NewConsulRegistry creates a Consul registry from configuration.
func NewConsulRegistry(cfg *ConsulConfig) (*ConsulRegistry, error) {
	consulCfg := &consulAPI.Config{
		Address: cfg.Address,
		Scheme:  "http",
	}
	if cfg.TLSEnabled {
		consulCfg.Scheme = "https"
	}
	client, err := consulAPI.NewClient(consulCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ConsulRegistry{
		client: client,
		config: cfg,
	}, nil
}

// Register registers a service instance with Consul.
func (r *ConsulRegistry) Register(ctx context.Context, service ServiceInfo) error {
	port := r.config.ServicePort
	if port == 0 {
		port = 50051
	}

	checkID := fmt.Sprintf("service:%s", r.config.ServiceID)
	register := &consulAPI.AgentServiceRegistration{
		ID:      r.config.ServiceID,
		Name:    r.config.ServiceName,
		Address: service.Addr,
		Port:    port,
		Tags:    r.addMetadataTags(r.config.ServiceTags, service.Metadata),
		Check: &consulAPI.AgentServiceCheck{
			ID:                       checkID,
			TTL:                      toTTLString(r.config.UnhealthyAfter),
			Status:                   string(consulAPI.HealthPassing),
			DeregisterCriticalServiceAfter: toTTLString(r.config.UnhealthyAfter * 2),
		},
	}

	err := r.client.Agent().ServiceRegister(register)
	if err != nil {
		return fmt.Errorf("failed to register service [%s]: %w", r.config.ServiceName, err)
	}

	if r.config.UnhealthyAfter > 0 {
		r.client.Agent().UpdateTTL(checkID, "service registered successfully", string(consulAPI.HealthPassing))
		go r.heartbeat(ctx, checkID)
	}

	return nil
}

func (r *ConsulRegistry) addMetadataTags(tags []string, metadata map[string]string) []string {
	if len(tags) == 0 {
		tags = []string{}
	}
	for key, value := range metadata {
		tags = append(tags, fmt.Sprintf("%s=%s", key, value))
	}
	return tags
}

func (r *ConsulRegistry) heartbeat(ctx context.Context, checkID string) {
	ticker := time.NewTicker(r.config.UnhealthyAfter / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.client.Agent().UpdateTTL(checkID, "service heartbeat", string(consulAPI.HealthPassing))
		}
	}
}

// Deregister removes a service instance from Consul.
func (r *ConsulRegistry) Deregister(ctx context.Context, service ServiceInfo) error {
	err := r.client.Agent().ServiceDeregister(r.config.ServiceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service [%s]: %w", r.config.ServiceName, err)
	}
	return nil
}

// Watch returns a channel for service discovery.
func (r *ConsulRegistry) Watch(ctx context.Context) (<-chan ServiceUpdate, error) {
	serviceUpdates := make(chan ServiceUpdate, 100)

	go func() {
		defer close(serviceUpdates)
		lastIndex := uint64(0)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			entries, meta, err := r.client.Catalog().Service(r.config.ServiceName, "", &consulAPI.QueryOptions{
				WaitTime: time.Second * 30,
				Index:    lastIndex,
			})

			if err != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}
			lastIndex = meta.LastIndex

			for _, entry := range entries {
				serviceUpdates <- ServiceUpdate{
					Service: ServiceInfo{
						Name: entry.Service.Service,
						Addr: fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port),
						Healthy: true,
					},
					Type: UpdateTypeRegister,
				}
			}
		}
	}()

	return serviceUpdates, nil
}

// Close releases Consul resources.
func (r *ConsulRegistry) Close() error {
	return nil
}

func toTTLString(d time.Duration) string {
	return fmt.Sprintf("%.0fs", d.Seconds())
}
