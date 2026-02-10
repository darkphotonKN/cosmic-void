package config

import (
	"log/slog"
	"net"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/notification"
	"github.com/darkphotonKN/cosmic-void-server/notification-service/internal/notification"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

// SetupServices initializes all services and their dependencies
func SetupServices(db *sqlx.DB, amqpChannel *amqp.Channel) *grpc.Server {
	// Create repository
	repo := notification.NewRepository(db)

	// Create service with repository and AMQP channel
	// publishCh := commonbroker.NewAmqpPublisher(amqpChannel)
	service := notification.NewService(repo)

	// Create gRPC handler with service
	handler := notification.NewHandler(service)

	// Create AMQP consumer with service
	consumer := notification.NewConsumer(service, amqpChannel)

	// Set up AMQP infrastructure
	if err := notification.SetupAMQPInfrastructure(amqpChannel); err != nil {
		slog.Error("Failed to setup AMQP infrastructure", "error", err)
	}

	// Start listening for AMQP events
	consumer.Listen()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register notification service with gRPC server
	pb.RegisterNotificationServiceServer(grpcServer, handler)

	slog.Info("Notification service initialized successfully")

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
