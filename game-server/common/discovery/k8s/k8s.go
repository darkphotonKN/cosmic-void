package k8s

import (
	"context"
	"fmt"
)

// Registry implements discovery.Registry on top of Kubernetes DNS.
//
// Register / Deregister / HealthCheck are intentionally no-ops: in k8s the
// Service object + readinessProbe already do the equivalent of registering
// pods and gating traffic on health. Discover returns a single deterministic
// in-cluster DNS name; kube-proxy handles L4 load-balancing across the
// healthy Endpoints of that Service.
type Registry struct {
	namespace string
}

// serviceEntry maps a code-level service name (the argument callers pass to
// Discover) onto the k8s Service name and gRPC port. Keep this aligned with
// each service's <svc>/k8s/service.yml manifest.
type serviceEntry struct {
	service string
	port    int
}

var serviceMap = map[string]serviceEntry{
	"auth":         {"auth-service", 7003},
	"payments":     {"payment-service", 7021},
	"items":        {"items-service", 7013},
	"stats":        {"stats-service", 7011},
	"notification": {"notification-service", 7077},
	"examples":     {"example-service", 7010},
	"game":         {"game-service", 7004},
	"api-gateway":  {"api-gateway", 7001},
}

func NewRegistry(namespace string) (*Registry, error) {
	if namespace == "" {
		namespace = "default"
	}
	return &Registry{namespace: namespace}, nil
}

func (r *Registry) Register(ctx context.Context, instanceID, serviceName, hostPort string) error {
	return nil
}

func (r *Registry) Deregister(ctx context.Context, instanceID, serviceName string) error {
	return nil
}

func (r *Registry) HealthCheck(instanceID, serviceName string) error {
	return nil
}

func (r *Registry) Discover(ctx context.Context, serviceName string) ([]string, error) {
	entry, ok := serviceMap[serviceName]
	if !ok {
		return nil, fmt.Errorf("k8s discovery: unknown service %q", serviceName)
	}
	return []string{fmt.Sprintf("%s.%s.svc.cluster.local:%d", entry.service, r.namespace, entry.port)}, nil
}
