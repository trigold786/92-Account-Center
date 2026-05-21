package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ServiceInstance struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Address   string            `json:"address"`
	Port      int               `json:"port"`
	Metadata  map[string]string `json:"metadata"`
	Healthy   bool              `json:"healthy"`
	RegisteredAt time.Time      `json:"registered_at"`
}

type ServiceRegistry interface {
	Register(ctx context.Context, instance *ServiceInstance) error
	Deregister(ctx context.Context, serviceID string) error
	Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	HealthCheck(ctx context.Context, serviceID string) error
}

type InMemoryRegistry struct {
	mu        sync.RWMutex
	services  map[string]*ServiceInstance
	byName    map[string][]string
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		services: make(map[string]*ServiceInstance),
		byName:   make(map[string][]string),
	}
}

func (r *InMemoryRegistry) Register(_ context.Context, instance *ServiceInstance) error {
	if instance.ID == "" || instance.Name == "" {
		return fmt.Errorf("service ID and name are required")
	}

	instance.RegisteredAt = time.Now()
	instance.Healthy = true

	r.mu.Lock()
	defer r.mu.Unlock()

	r.services[instance.ID] = instance
	r.byName[instance.Name] = append(r.byName[instance.Name], instance.ID)
	return nil
}

func (r *InMemoryRegistry) Deregister(_ context.Context, serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, ok := r.services[serviceID]
	if !ok {
		return fmt.Errorf("service %s not found", serviceID)
	}

	delete(r.services, serviceID)

	ids := r.byName[svc.Name]
	for i, id := range ids {
		if id == serviceID {
			r.byName[svc.Name] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	if len(r.byName[svc.Name]) == 0 {
		delete(r.byName, svc.Name)
	}

	return nil
}

func (r *InMemoryRegistry) Discover(_ context.Context, serviceName string) ([]*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, ok := r.byName[serviceName]
	if !ok {
		return nil, nil
	}

	var instances []*ServiceInstance
	for _, id := range ids {
		if svc, ok := r.services[id]; ok && svc.Healthy {
			instances = append(instances, svc)
		}
	}
	return instances, nil
}

func (r *InMemoryRegistry) HealthCheck(_ context.Context, serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, ok := r.services[serviceID]
	if !ok {
		return fmt.Errorf("service %s not found", serviceID)
	}
	svc.Healthy = true
	return nil
}
