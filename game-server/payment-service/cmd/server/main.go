package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/payment"
	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/payment-service/config"
	"github.com/darkphotonKN/cosmic-void-server/payment-service/internal/payment"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/stripe/stripe-go/v82"
	"google.golang.org/grpc"
)

var (
	serviceName    = "payments"
	grpcAddr       = commonhelpers.GetEnvString("GRPC_PAYMENT_ADDR", "7021")
	consulAddr     = commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")
	amqpUser       = commonhelpers.GetEnvString("RABBITMQ_USER", "guest")
	amqpPassword   = commonhelpers.GetEnvString("RABBITMQ_PASS", "guest")
	amqpHost       = commonhelpers.GetEnvString("RABBITMQ_HOST", "localhost")
	amqpPort       = commonhelpers.GetEnvString("RABBITMQ_PORT", "5672")
	stripeSecretKey = commonhelpers.GetEnvString("STRIPE_SECRET_KEY", "")
)

func main() {
	// --- stripe setup ---
	if stripeSecretKey == "" {
		log.Fatal("STRIPE_SECRET_KEY is required")
	}
	stripe.Key = stripeSecretKey

	// --- database setup ---
	db := config.InitDB()
	defer db.Close()

	// --- service discovery setup ---
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

	// --- health check ---
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

	listener, err := net.Listen("tcp", "localhost:"+grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen at port: %s\nError: %s\n", grpcAddr, err)
	}
	defer listener.Close()

	// --- message broker - rabbit mq ---
	ch, close := broker.Connect(amqpUser, amqpPassword, amqpHost, amqpPort)
	defer func() {
		close()
		ch.Close()
	}()

	// Initialize layers
	repo := payment.NewRepository(db)
	processor := payment.NewStripeProcessor()
	svc := payment.NewService(repo, processor, ch, registry)
	handler := payment.NewHandler(svc)

	pb.RegisterPaymentServiceServer(grpcServer, handler)

	log.Printf("gRPC Payment Server started on PORT: %s\n", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Can't connect to grpc server. Error:", err.Error())
	}
}
