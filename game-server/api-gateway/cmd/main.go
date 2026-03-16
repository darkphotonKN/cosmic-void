package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/api-gateway/config"
	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commontelemetry "github.com/darkphotonKN/cosmic-void-server/common/telemetry"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	_ "github.com/joho/godotenv/autoload"
)

var (
	environment = commonhelpers.GetEnvString("ENVIRONMENT", "development")

	// observability
	collectorEndpoint = commonhelpers.GetEnvString("COLLECTOR_ENDPOINT", "localhost:4317")
	serviceVersion    = commonhelpers.GetEnvString("SERVICE_VERSION", "1.0.0")

	serviceName = "api-gateway"

	// grpc
	httpAddr   = commonhelpers.GetEnvString("PORT", "7001")
	consulAddr = commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")

	// rabbitmq
	amqpUser     = commonhelpers.GetEnvString("RABBITMQ_USER", "guest")
	amqpPassword = commonhelpers.GetEnvString("RABBITMQ_PASS", "guest")
	amqpHost     = commonhelpers.GetEnvString("RABBITMQ_HOST", "localhost")
	amqpPort     = commonhelpers.GetEnvString("RABBITMQ_PORT", "5672")
)

/**
* Main entry point to entire application.
* NOTE: Keep code here as clean and little as possible.
**/
func main() {
	ctx := context.Background()
	fmt.Println("testerson")

	// --- observability ---
	shutdown, err := commontelemetry.Init(ctx, commontelemetry.Config{
		ServiceName:       serviceName,
		ServiceVersion:    serviceVersion,
		Environment:       environment,
		CollectorEndpoint: collectorEndpoint,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer shutdown(ctx)
	// --- database setup removed - api gateway is stateless ---

	// --- service discovery setup ---

	// -- consul client --
	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		log.Fatal("Failed to create Consul registry")
	}

	instanceID := discovery.GenerateInstanceID(serviceName)

	// -- discovery --
	if err := registry.Register(ctx, instanceID, serviceName, "localhost:"+httpAddr); err != nil {
		fmt.Printf("\nError when registering service:\n\n%s\n\n", err)
		panic(err)
	}

	// -- health check --
	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				log.Fatal("Health check failed.")
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	// --- message broker - rabbit mq ---
	ch, closeCh := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)
	broker.DeclareExchange(ch, commonconstants.AuthEventsExchange, "topic")
	defer func() {
		closeCh()
		ch.Close()
	}()

	// --- router setup ---
	router := config.SetupRouter(registry, ch)

	// -- start server --
	if err := router.Run(fmt.Sprintf(":%s", httpAddr)); err != nil {
		log.Fatal("Failed to start server")
	}
}
