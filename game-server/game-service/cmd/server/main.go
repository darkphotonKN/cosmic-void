package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/config"
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/game"
	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

var (
	// grpc
	serviceName = "game"
	grpcAddr    = commonhelpers.GetEnvString("GRPC_GAME_ADDR", "7004")
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

	broker.DeclareExchange(ch, commonconstants.RoomCreatedEvent, "fanout")
	broker.DeclareExchange(ch, commonconstants.GameStartedEvent, "fanout")
	broker.DeclareExchange(ch, commonconstants.GameEndedEvent, "fanout")
	defer func() {
		close()
		ch.Close()
	}()

	// TODO: Initialize services and handlers
	// This is where you'll add your game service implementation
	// repo := game.NewRepository(db)
	// service := game.NewService(repo, ch)
	// handler := game.NewHandler(service)
	// pb.RegisterGameServiceServer(grpcServer, handler)

	log.Printf("grpc Game Server started on PORT: %s\n", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Can't connect to grpc server. Error:", err.Error())
	}
}