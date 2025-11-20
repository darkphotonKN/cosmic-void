package consul

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	consul "github.com/hashicorp/consul/api"
	"strconv"
	"strings"
)

// Registry provides a wrapper around the Consul client
// to handle service registration and discovery operations
type Registry struct {
	client *consul.Client // The underlying Consul client that communicates with the Consul server
}

// NewRegistry creates a new Registry with a connection to the Consul server
// addr: The address of the Consul server (e.g., "localhost:8500")
// serviceName: The name of the service to be registered (unused in this implementation but common in other patterns)
func NewRegistry(addr, serviceName string) (*Registry, error) {
	// Create default configuration for Consul client
	// DefaultConfig provides sensible defaults for timeouts, etc.
	config := consul.DefaultConfig()

	// Override the default address with the provided one
	config.Address = addr

	// Initialize a new Consul client with our configuration
	// This establishes the connection to the Consul server
	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	// Return a new Registry instance with the configured client
	return &Registry{
		client: client,
	}, nil
}

// Register adds a service instance to Consul's service registry
// ctx: Context for the operation (unused here but allows for future deadline/cancellation)
// instanceID: Unique identifier for this specific instance of the service
// serviceName: The type of service being registered (e.g., "order-service")
// hostPort: The address where this service instance can be reached (e.g., "localhost:2221")
func (r *Registry) Register(ctx context.Context, instanceID, serviceName, hostPort string) error {
	// Split the hostPort string to extract host and port separately
	// Example: "localhost:2221" -> ["localhost", "2221"]
	parts := strings.Split(hostPort, ":")
	if len(parts) < 2 {
		return errors.New("Error occured when splitting port: hostPort format is invalid.")
	}

	// Convert the port from string to integer as required by Consul
	port, err := strconv.Atoi(parts[1]) // port
	host := parts[0]                    // host
	if err != nil {
		return err
	}

	// Register the service with Consul using the Agent API
	// AgentServiceRegistration defines the service properties and health check configuration
	err = r.client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		ID:      instanceID,  // Unique ID for this service instance
		Address: host,        // Host where the service is running
		Port:    port,        // Port the service is listening on
		Name:    serviceName, // Type of service (used for discovery)
		Check: &consul.AgentServiceCheck{ // Health check configuration
			CheckID:                        instanceID, // ID for this health check
			TLSSkipVerify:                  true,       // Skip TLS verification for HTTPS checks
			TTL:                            "5s",       // Time-to-live: service must update health status within this period
			Timeout:                        "1s",       // How long Consul waits for a check before considering it failed
			DeregisterCriticalServiceAfter: "10s",      // Auto-deregister if service remains unhealthy for this duration
		},
	})

	if err != nil {
		fmt.Printf("\nError when attempting to start service register agent: %s\n\n", err)
		return err
	}

	return nil
}

// Deregister removes a service instance from Consul's registry
// This is typically called when a service is shutting down gracefully
func (r *Registry) Deregister(ctx context.Context, instanceID, serverName string) error {
	// Remove the health check, which effectively deregisters the service
	return r.client.Agent().CheckDeregister(instanceID)
}

// HealthCheck updates the TTL check for a service instance
// This must be called periodically (within the TTL time) to keep the service marked as healthy
// Think of this as the service saying "I'm still alive!" to Consul
func (r *Registry) HealthCheck(instanceID, serviceName string) error {
	// UpdateTTL refreshes the TTL timer and sets the service status to "passing"
	return r.client.Agent().UpdateTTL(instanceID, "online", api.HealthPassing)
}

// Discover finds all healthy instances of a particular service type
// Returns a list of host:port strings for all healthy instances
// This is how services find other services they need to communicate with
func (r *Registry) Discover(ctx context.Context, serviceName string) ([]string, error) {
	// Query Consul for healthy instances of the specified service
	// The empty string parameter is for tags (unused here)
	// The 'true' parameter means "only return passing services"
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)

	// TODO: REMOVE AFTER DEBUG
	fmt.Printf("\ncurrent entries: \n\n%+v\n\n", entries)

	if err != nil {
		return nil, err
	}

	// Build a list of host:port strings from the results
	var instances []string
	for _, entry := range entries {
		instances = append(instances, fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port))
	}

	// TODO: REMOVE AFTER DEBUG
	fmt.Printf("\ncurrent instances: \n\n%+v\n\n", instances)

	// Return the list of discovered service instances
	return instances, nil
}
