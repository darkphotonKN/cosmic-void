package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/auth-service/config"
	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/member"
	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/upload"
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commonoutbox "github.com/darkphotonKN/cosmic-void-server/common/outbox"
	commontelemetry "github.com/darkphotonKN/cosmic-void-server/common/telemetry"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/common/utils/cache"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"log/slog"
)

var (
	environment = commonhelpers.GetEnvString("ENVIRONMENT", "development")

	// observability
	collectorEndpoint = commonhelpers.GetEnvString("COLLECTOR_ENDPOINT", "localhost:4317")
	serviceVersion    = commonhelpers.GetEnvString("SERVICE_VERSION", "1.0.0")

	// grpc
	serviceName = "auth"
	grpcAddr    = commonhelpers.GetEnvString("GRPC_AUTH_ADDR", "7003")
	consulAddr  = commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")

	// rabbit mq
	amqpUser     = commonhelpers.GetEnvString("RABBITMQ_USER", "guest")
	amqpPassword = commonhelpers.GetEnvString("RABBITMQ_PASS", "guest")
	amqpHost     = commonhelpers.GetEnvString("RABBITMQ_HOST", "localhost")
	amqpPort     = commonhelpers.GetEnvString("RABBITMQ_PORT", "5672")
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

	// --- S3 setup ---
	s3Config := config.LoadS3Config()
	s3Client, err := config.InitS3Client(ctx, s3Config)
	if err != nil {
		log.Printf("Warning: S3 client initialization failed: %v", err)
		log.Println("Avatar upload functionality will be disabled")
		// Continue without S3 avatar uploads will not work but service runs
		s3Client = nil
	}

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
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

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
		log.Println("Metrics server started on :8081")
		http.ListenAndServe(":8081", nil)
	}()

	// --- message broker - rabbit mq ---
	ch, close := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)

	broker.DeclareExchange(ch, commonconstants.AuthEventsExchange, "topic")
	defer func() {
		close()
		ch.Close()
	}()

	// --- member service setup ---
	publishCh := commonbroker.NewAmqpPublisher(ch)
	memberRepo := member.NewRepository(db)

	// --- outbox workers ---
	// The worker drains the outbox table and publishes events to RabbitMQ.
	// Without it, rows written by CreateMember would sit in the DB forever
	// and downstream services (notification, etc.) would never see the
	// event. 5s cycle keeps sign-up → notification latency small while
	// staying cheap for a low-traffic auth DB.
	outboxRepo := commonoutbox.NewRepo(db)
	outboxService := commonoutbox.NewService(outboxRepo)

	memberService := member.NewService(db, memberRepo, publishCh, cacheService, outboxService)
	memberHandler := member.NewHandler(memberService)

	outboxWorker := commonoutbox.NewOutboxWorker(time.Second*5, 20, outboxService, publishCh)
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	// InitiateWork spawns its own goroutine and returns immediately — do
	// not wrap it in another goroutine, that would cancel the context as
	// soon as the wrapper returned and kill the worker.
	go outboxWorker.Run(workerCtx)

	// --- upload service setup ---
	if s3Client != nil {
		uploadRepo := upload.NewRepository(db)
		logger := slog.Default()

		uploadService := upload.NewService(
			uploadRepo,
			s3Client,
			memberService,
			s3Config.BucketName,
			s3Config.CDNUrl,
			logger,
			publishCh,
			db,
		)

		uploadHandler := upload.NewHandler(uploadService)
		pb.RegisterUploadServiceServer(grpcServer, uploadHandler)
	}

	// rabbitmq consumer
	consumer := member.NewConsumer(memberService, ch)
	if err := consumer.SetupConsumer(); err != nil {
		log.Fatalf("Failed to setup auth RPC infrastructure: %v", err)
	}
	consumer.Listen()

	pb.RegisterAuthServiceServer(grpcServer, memberHandler)

	log.Printf("grpc Auth Server started on PORT: %s\n", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Can't connect to grpc server. Error:", err.Error())
	}
}
