package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/game-service/config"
	grpcitems "github.com/darkphotonKN/cosmic-void-server/game-service/grpc/items"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

var (
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
	defer func() {
		close()
		ch.Close()
	}()

	broker.DeclareExchange(ch, commonconstants.GameMatchEndedEvent, "fanout")

	// TODO: Initialize your services and handlers
	// repo := yourpackage.NewRepository(db)
	// service := yourpackage.NewService(repo, ch)
	// handler := yourpackage.NewHandler(service)
	// pb.RegisterGameServiceServer(grpcServer, handler)

	log.Printf("grpc Game Server started on PORT: %s\n", grpcAddr)

	// --- TEST: Items gRPC Client ---
	// 在 grpcServer.Serve() 之前測試
	itemsClient := grpcitems.NewClient(registry)
	go func() {
		time.Sleep(3 * time.Second) // 等待 items-service 啟動
		ListAllWeapons(itemsClient)
	}()
	// --- END TEST ---

	// routes setup
	routes := config.SetupRouter(db, registry, ch)

	fmt.Printf("Server listening on port %s.\n", gamePort)

	go func() {
		err := routes.Run(gamePort)

		if err != nil {
			log.Fatal("Can't connect to game server. Error:", err.Error())
		}

	}()

	// 這個會阻塞，所以測試要放在上面
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Can't connect to grpc server. Error:", err.Error())
	}
}

func ListAllWeapons(itemsClient grpcitems.ItemsClient) {
	log.Println("\n========================================")
	log.Println("=== Testing Items gRPC Client ===")
	log.Println("========================================")

	ctx := context.Background()

	// 調用 items service
	log.Println("📞 Calling itemsClient.ListWeaponsWithTemplate()...")
	response, err := itemsClient.ListWeaponsWithTemplate(ctx)
	if err != nil {
		log.Printf("❌ Failed to list weapons: %v", err)
		log.Println("💡 Possible reasons:")
		log.Println("   1. Items-service is not running")
		log.Println("   2. Items-service not registered in Consul")
		log.Println("   3. Network connection issue")
		return
	}

	log.Printf("✅ Successfully got response!")
	log.Printf("📦 Number of weapons: %d", len(response.Weapons))

	// 使用返回的武器數據
	if len(response.Weapons) == 0 {
		log.Println("⚠️  No weapons found in database")
		log.Println("💡 Try creating some weapons first using items-service API")
	} else {
		log.Println("\n🗡️  Weapon List:")
		for i, weapon := range response.Weapons {
			log.Printf("  [%d] %s", i+1, weapon.ItemName)
			log.Printf("      ├─ ID: %s", weapon.Id)
			log.Printf("      ├─ Attack: %d", weapon.AttackPower)
			log.Printf("      ├─ Durability: %d", weapon.Durability)
			log.Printf("      ├─ Critical Rate: %.2f", weapon.CriticalRate)
			log.Printf("      ├─ Type: %s", weapon.WeaponType)
			log.Printf("      └─ Description: %s", weapon.Description)
		}
	}

	log.Println("\n========================================")
	log.Println("=== Test Complete ===")
	log.Println("========================================\n")
}
