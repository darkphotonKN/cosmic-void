package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/stats-service/config"
	"github.com/darkphotonKN/cosmic-void-server/stats-service/internal/stats"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

var (
	serviceName = "stats"
	grpcAddr    = commonhelpers.GetEnvString("GRPC_STATS_ADDR", "7011")
	consulAddr  = commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")

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

	ch, close := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)

	broker.DeclareExchange(ch, commonconstants.StatsUpdatedEvent, "fanout")
	broker.DeclareExchange(ch, commonconstants.MatchCompletedEvent, "fanout")
	broker.DeclareExchange(ch, commonconstants.PlayerActionEvent, "fanout")

	defer func() {
		close()
		ch.Close()
	}()

	repo := stats.NewRepository(db)
	service := stats.NewService(repo, ch)
	handler := stats.NewHandler(service)
	consumer := stats.NewConsumer(service, ch)
	consumer.Listen()

	pb.RegisterStatsServiceServer(grpcServer, handler)

	log.Printf("grpc Stats Server started on PORT: %s\n", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Can't connect to grpc server. Error:", err.Error())
	}
}