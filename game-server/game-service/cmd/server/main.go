package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commontelemetry "github.com/darkphotonKN/cosmic-void-server/common/telemetry"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/common/utils/cache"
	"github.com/darkphotonKN/cosmic-void-server/game-service/config"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components/metrics"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
	environment = commonhelpers.GetEnvString("ENVIRONMENT", "development")

	// observability
	collectorEndpoint = commonhelpers.GetEnvString("COLLECTOR_ENDPOINT", "localhost:4317")
	serviceVersion    = commonhelpers.GetEnvString("SERVICE_VERSION", "1.0.0")

	// game service
	gamePort = fmt.Sprintf(":%s", commonhelpers.GetEnvString("GAME_PORT", "5555"))

	// grpc
	serviceName  = "game"
	grpcAuthAddr = commonhelpers.GetEnvString("GRPC_AUTH_ADDR", "7003")
	grpcAddr     = commonhelpers.GetEnvString("GRPC_GAME_ADDR", "7004")
	consulAddr   = commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")

	// rabbit mq
	amqpUser     = commonhelpers.GetEnvString("RABBITMQ_USER", "guest")
	amqpPassword = commonhelpers.GetEnvString("RABBITMQ_PASS", "guest")
	amqpHost     = commonhelpers.GetEnvString("RABBITMQ_HOST", "localhost")
	amqpPort     = commonhelpers.GetEnvString("RABBITMQ_PORT", "5672")
)

func main() {
	ctx := context.Background()

	// --- pprof ---

	// TODO: remove after benchmark passes
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(1)

	// --- logger ---
	commonhelpers.SetupLogger(environment)

	// --- observability ---

	// -- default --
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

	// -- custom metrics --
	err = metrics.Init()
	if err != nil {
		log.Printf("\ncustom metrics setup init errored. Error: %w\n\n", err)
	}

	// --- database setup ---

	db := config.InitDB()
	defer db.Close()

	// --- redis setup ---
	err = config.InitRedis(config.RedisConfig{
		Mode:         commonhelpers.GetEnvString("REDIS_MODE", "standalone"),
		Addrs:        []string{commonhelpers.GetEnvString("REDIS_ADDR", "localhost:6379")},
		Password:     commonhelpers.GetEnvString("REDIS_PASSWORD", ""),
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer config.CloseRedis()
	cacheService := cache.NewRedisCache(config.GetClient())

	// --- service discovery setup ---

	// -- consul client --
	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		log.Fatal("Failed to create Consul registry")
	}

	instanceID := discovery.GenerateInstanceID(serviceName)

	// -- discovery --
	if err := registry.Register(ctx, instanceID, serviceName, "localhost:"+grpcAddr); err != nil {
		log.Printf("\nError when registering service:\n\n%s\n\n", err)
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

	// --- grpc ---
	grpcServer := grpc.NewServer()

	// create a network listener to this service
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
		log.Println("Metrics server started on :8082")
		http.ListenAndServe(":8082", nil)
	}()

	// --- message broker - rabbit mq ---
	ch, close := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)
	defer func() {
		close()
		ch.Close()
	}()

	broker.DeclareExchange(ch, commonconstants.GameEventsExchange, "topic")

	// TODO: Initialize your services and handlers
	// repo := yourpackage.NewRepository(db)
	// service := yourpackage.NewService(repo, ch)
	// handler := yourpackage.NewHandler(service)
	// pb.RegisterGameServiceServer(grpcServer, handler)

	log.Printf("grpc Game Server started on PORT: %s\n", grpcAddr)

	// routes setup
	routes := config.SetupRouter(db, registry, ch , cacheService)

	fmt.Printf("Server listening on port %s.\n", gamePort)

	go func() {
		err := routes.Run(gamePort)

		if err != nil {
			log.Fatal("Can't connect to game server. Error:", err.Error())
		}

	}()

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Can't connect to grpc server. Error:", err.Error())
	}
}
