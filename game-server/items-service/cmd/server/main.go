package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commontelemetry "github.com/darkphotonKN/cosmic-void-server/common/telemetry"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/items-service/config"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
	// observability
	environment       = commonhelpers.GetEnvString("ENVIRONMENT", "development")
	collectorEndpoint = commonhelpers.GetEnvString("COLLECTOR_ENDPOINT", "localhost:4317")

	serviceName    = "items"
	grpcAddr       = commonhelpers.GetEnvString("GRPC_ITEMS_ADDR", "7013")
	consulAddr     = commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")
	serviceVersion = commonhelpers.GetEnvString("SERVICE_VERSION", "1.0.0")

	amqpUser     = commonhelpers.GetEnvString("RABBITMQ_USER", "guest")
	amqpPassword = commonhelpers.GetEnvString("RABBITMQ_PASS", "guest")
	amqpHost     = commonhelpers.GetEnvString("RABBITMQ_HOST", "localhost")
	amqpPort     = commonhelpers.GetEnvString("RABBITMQ_PORT", "5672")
)

func main() {
	db := config.InitDB()
	defer db.Close()

	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		log.Fatal("Failed to create Consul registry")
	}

	ctx := context.Background()

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

	// test
	// repo := items.NewRepository(db)
	// testItemId := uuid.MustParse("aa0e8400-e29b-41d4-a716-446655440001")

	// itemData, err := repo.GetItemTemplateByID(ctx, testItemId)
	// slog.Info("Debugging get item template", "itemData", itemData)
	// end test

	instanceID := discovery.GenerateInstanceID(serviceName)

	if err := registry.Register(ctx, instanceID, serviceName, "localhost:"+grpcAddr); err != nil {
		log.Printf("\nError when registering service:\n\n%s\n\n", err)
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				log.Fatal("Health check failed.")
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	grpcServer := grpc.NewServer()

	listener, err := net.Listen("tcp", "localhost:"+grpcAddr)
	if err != nil {
		log.Fatalf(
			"Failed to listen at port: %s\nError: %s\n", grpcAddr, err,
		)
	}
	defer listener.Close()

	// --- metrics ---

	// setup endpoint for metrics collection
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Metrics server started on :7013")
		http.ListenAndServe(":7013", nil)
	}()

	ch, close := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)

	// Only declare the exchange we actually consume from
	broker.DeclareExchange(ch, commonconstants.GameEventsExchange, "topic")

	defer func() {
		close()
		ch.Close()
	}()

	// Use the new config setup to initialize all services
	grpcServer = config.SetupServices(db, ch)

	log.Printf("grpc Items Server started on PORT: %s\n", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Can't connect to grpc server. Error:", err.Error())
	}
}
