package config

import (
	"log/slog"
	"net"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	"github.com/darkphotonKN/cosmic-void-server/stats-service/internal/stats"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

// SetupServices initializes all services and their dependencies
func SetupServices(db *sqlx.DB, amqpChannel *amqp.Channel) *grpc.Server {
	// create repository
	repo := stats.NewRepository(db)

	// create service with repository and AMQP channel
	publishCh := commonbroker.NewAmqpPublisher(amqpChannel) // adapter
	service := stats.NewService(repo, publishCh)

	// create gRPC handler with service
	handler := stats.NewHandler(service)

	// create AMQP consumer with service
	consumer := stats.NewConsumer(service, amqpChannel)

	// set up AMQP infrastructure
	if err := stats.SetupAMQPInfrastructure(amqpChannel); err != nil {
		slog.Error("Failed to setup AMQP infrastructure", "error", err)
	}

	// start listening for AMQP events
	consumer.Listen()

	// create gRPC server
	grpcServer := grpc.NewServer()

	// Register stats service with gRPC server
	pb.RegisterStatsServiceServer(grpcServer, handler)

	slog.Info("Stats service initialized successfully")

	return grpcServer
}

// StartGRPCServer starts the gRPC server on the specified port
func StartGRPCServer(grpcServer *grpc.Server, port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	slog.Info("Starting gRPC server", "port", port)

	// This blocks until the server is stopped
	if err := grpcServer.Serve(listener); err != nil {
		return err
	}

	return nil
}

// InitializeAMQPConnection establishes connection to RabbitMQ
func InitializeAMQPConnection(amqpURL string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	slog.Info("Connected to RabbitMQ")

	return conn, channel, nil
}
