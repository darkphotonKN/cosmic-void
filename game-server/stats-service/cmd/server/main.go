package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commonoutbox "github.com/darkphotonKN/cosmic-void-server/common/outbox"
	commontelemetry "github.com/darkphotonKN/cosmic-void-server/common/telemetry"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/stats-service/config"
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

	serviceName = "stats"
	grpcAddr    = commonhelpers.GetEnvString("GRPC_STATS_ADDR", "7011")
	consulAddr  = commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")

	amqpUser     = commonhelpers.GetEnvString("RABBITMQ_USER", "guest")
	amqpPassword = commonhelpers.GetEnvString("RABBITMQ_PASS", "guest")
	amqpHost     = commonhelpers.GetEnvString("RABBITMQ_HOST", "localhost")
	amqpPort     = commonhelpers.GetEnvString("RABBITMQ_PORT", "5672")

	// redis
	redisHost = commonhelpers.GetEnvString("REDIS_HOST", "localhost")
	redisPort = commonhelpers.GetEnvString("REDIS_PORT", "6379")
	redisDB   = commonhelpers.GetEnvString("REDIS_DB", "0")
)

func main() {
	ctx := context.Background()

	// --- logger ---
	commonhelpers.SetupLogger(environment)

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

	// --- database setup ---
	db := config.InitDB()
	defer db.Close()

	// --- redis setup ---
	redisDBInt, err := strconv.Atoi(redisDB)
	if err != nil {
		log.Fatal(err)
	}

	err = config.InitRedis(config.RedisConfig{
		Mode:         "standalone",
		Addrs:        []string{redisHost + ":" + redisPort},
		Password:     "",
		DB:           redisDBInt,
		PoolSize:     10,
		MinIdleConns: 5,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer config.CloseRedis()

	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		log.Fatal("Failed to create Consul registry")
	}

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
		log.Println("Metrics server started on :8083")
		http.ListenAndServe(":8083", nil)
	}()

	ch, close := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)

	defer func() {
		close()
		ch.Close()
	}()

	// --- outbox workers ---
	outboxRepo := commonoutbox.NewRepo(db)
	outboxServ := commonoutbox.NewService(outboxRepo)
	publisher := commonbroker.NewAmqpPublisher(ch)
	outboxWorker := commonoutbox.NewOutboxWorker(time.Minute*60, 20, outboxServ, publisher)

	workerCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()

		outboxWorker.InitiateWork(workerCtx)
	}()

	// use the new config setup to initialize all services
	grpcServer = config.SetupServices(db, ch)

	log.Printf("grpc Stats Server started on PORT: %s\n", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Can't connect to grpc server. Error:", err.Error())
	}
}
