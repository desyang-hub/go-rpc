package registry

import (
	"context"
	"fmt"
	"path"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdConfig contains etcd-specific configuration.
type EtcdConfig struct {
	Endpoints    []string
	TLSEnabled   bool
	CACert       string
	ClientCert   string
	ClientKey    string
	User         string
	Password     string
	ServiceName  string
	Prefix       string
	TTL          time.Duration
	TickInterval time.Duration
}

// EtcdRegistry provides etcd-based service discovery.
type EtcdRegistry struct {
	client *clientv3.Client
	config *EtcdConfig
	leases map[string]int64
}

// NewEtcdRegistry creates an etcd registry from configuration.
func NewEtcdRegistry(cfg *EtcdConfig) (*EtcdRegistry, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &EtcdRegistry{
		client: client,
		config: cfg,
		leases: make(map[string]int64),
	}, nil
}

func (r *EtcdRegistry) getServicePrefix() string {
	return path.Join(r.config.Prefix, r.config.ServiceName)
}

func (r *EtcdRegistry) getServiceKey(service ServiceInfo) string {
	return path.Join(r.getServicePrefix(), service.Addr)
}

func (r *EtcdRegistry) setLeaseID(key string, leaseID int64) {
	r.leases[key] = leaseID
}

// Register registers a service instance with etcd using lease-based TTL.
func (r *EtcdRegistry) Register(ctx context.Context, service ServiceInfo) error {
	key := r.getServiceKey(service)
	value := service.ToJSON()

	ttlSec := int64(r.config.TTL.Seconds())
	if ttlSec <= 0 {
		ttlSec = 30
	}
	lease, err := r.client.Grant(ctx, ttlSec)
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}

	r.setLeaseID(key, int64(lease.ID))

	_, err = r.client.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	tick := r.config.TickInterval
	if tick == 0 {
		tick = time.Duration(int64(ttlSec)/3) * time.Second
	}
	go func() {
		ticker := time.NewTicker(tick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				if id, ok := r.leases[key]; ok {
					_, _ = r.client.Revoke(ctx, clientv3.LeaseID(id))
				}
				return
			case <-ticker.C:
				_, err := r.client.KeepAliveOnce(ctx, lease.ID)
				if err != nil {
					return
				}
			}
		}
	}()

	return nil
}

// Deregister removes a service instance from etcd.
func (r *EtcdRegistry) Deregister(ctx context.Context, service ServiceInfo) error {
	key := r.getServiceKey(service)
	_, err := r.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}
	if id, ok := r.leases[key]; ok {
		_, _ = r.client.Revoke(ctx, clientv3.LeaseID(id))
		delete(r.leases, key)
	}
	return nil
}

// Watch returns a channel for service discovery.
func (r *EtcdRegistry) Watch(ctx context.Context) (<-chan ServiceUpdate, error) {
	prefix := r.getServicePrefix()
	serviceUpdates := make(chan ServiceUpdate, 100)

	go func() {
		defer close(serviceUpdates)
		watchChan := r.client.Watch(ctx, prefix+"/", clientv3.WithPrefix())

		for wresp := range watchChan {
			for _, ev := range wresp.Events {
				onUpdate := ServiceUpdate{Type: UpdateTypeRegister}
				if ev.Type == clientv3.EventTypePut {
					onUpdate.Type = UpdateTypeRegister
				} else if ev.Type == clientv3.EventTypeDelete {
					onUpdate.Type = UpdateTypeDeregister
				}
				serviceUpdates <- onUpdate
			}
		}
	}()

	return serviceUpdates, nil
}

// Close closes the etcd client.
func (r *EtcdRegistry) Close() error {
	return r.client.Close()
}
