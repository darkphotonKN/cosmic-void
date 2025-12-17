package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/auth-service/config"
	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/member"
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/common/utils/cache"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

var (
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
	// --- database setup ---

	db := config.InitDB()
	defer db.Close()

	// --- redis setup ---
	err := config.InitRedis(config.RedisConfig{
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

	ctx := context.Background()
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

	// --- message broker - rabbit mq ---
	ch, close := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)

	broker.DeclareExchange(ch, commonconstants.MemberSignedUpEvent, "fanout")
	defer func() {
		close()
		ch.Close()
	}()

	repo := member.NewRepository(db)
	service := member.NewService(repo, ch, cacheService)
	handler := member.NewHandler(service)

	// consumer := member.NewConsumer(service, ch)
	// start goroutine and listen to events from message broker
	// consumer.Listen()

	pb.RegisterAuthServiceServer(grpcServer, handler)

	log.Printf("grpc Order Server started on PORT: %s\n", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Can't connect to grpc server. Error:", err.Error())
	}
}
