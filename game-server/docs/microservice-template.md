# Microservice 模板產生器

## 目標
根據 example-service 結構，產生新的 microservice 骨架，只包含初始化函數（NewService、NewHandler、NewRepository），不含具體業務邏輯。

## 現有 Port 配置（掃描自專案 .env 及 docker-compose.yml）

### gRPC/HTTP Ports（70xx 系列）
| 服務 | Port | 來源 |
|------|------|------|
| api-gateway | 7001 | `PORT` |
| item-service | 7002 | `GRPC_ITEM_SERVICE_ADDR` |
| auth-service | 7003 | `GRPC_AUTH_ADDR` |
| game-service | 7004 | `GRPC_GAME_ADDR` |
| example-service | 7010 | `GRPC_EXAMPLE_ADDR` |
| stats-service | 7011 | `GRPC_STATS_ADDR` |
| auth-service (alt) | 7020 | `PORT` in auth-service/.env |
| payment-service | 7021 | `GRPC_PAYMENT_ADDR` |

**最高 gRPC Port：7021**

### DB Ports（5xxx 系列）
| 服務 | Port | 來源 |
|------|------|------|
| game-service | 5030 | docker-compose |
| api-gateway | 5101 | docker-compose |
| auth-service | 5103 | docker-compose |
| stats-service | 5104 | docker-compose |
| example-service | 5110 | docker-compose |
| payment-service | 5111 | docker-compose |

**最高 DB Port：5111**

### Test DB Ports（6xxx 系列）
| 服務 | Port |
|------|------|
| api-gateway | 6101 |
| stats-service | 6104 |

**最高 Test DB Port：6104**

---

## 下一個服務 Port 規則

```
gRPC Port = 最高 gRPC Port + 1 = 7022
DB Port   = 最高 DB Port + 1   = 5112
Test DB   = 最高 Test DB + 1   = 6105（如需要）
```

---

## 如何找出下一個可用 Port

在創建新服務前，執行以下指令掃描現有 port：

```bash
# 掃描所有 gRPC/HTTP port（70xx 系列）
grep -rh "PORT\|ADDR" game-server/**/.env 2>/dev/null | grep -E "[0-9]{4}" | sort -u

# 掃描所有 DB port（從 docker-compose）
grep -rh "ports:" -A1 game-server/**/docker-compose.yml | grep -E "51[0-9]{2}|50[0-9]{2}" | sort -u

# 或使用 ripgrep 更快速
rg "GRPC.*ADDR.*[0-9]{4}|PORT.*[0-9]{4}" game-server --no-filename | sort -u
rg '"5[01][0-9]{2}:5432"' game-server --no-filename | sort -u
```

找到最高數字後 +1 即為新服務的 port。

---

## Common 套件引用參考

| 功能 | Import 路徑 | 別名 | 用法 |
|-----|-------------|------|------|
| 環境變數 | `common/utils` | `commonhelpers` | `commonhelpers.GetEnvString("KEY", "default")` |
| 服務發現 | `common/discovery` | `discovery` | `discovery.GenerateInstanceID(serviceName)` |
| Consul | `common/discovery/consul` | `consul` | `consul.NewRegistry(addr, name)` |
| 訊息佇列 | `common/broker` | `broker` | `broker.Connect(user, pass, host, port)` |
| 常數 | `common/constants` | `commonconstants` | `commonconstants.ExampleCreatedEvent` |
| Proto | `common/api/proto/{service}` | `pb` | `pb.Register{Service}ServiceServer(...)` |

### 完整 Import 範例

```go
import (
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/{service}"
	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
)
```

---

## 要產生的目錄結構

```
{service-name}/
├── cmd/
│   └── server/
│       └── main.go
├── config/
│   └── db.go
├── internal/
│   └── {service}/
│       ├── handler.go
│       ├── service.go
│       ├── repository.go
│       └── model.go
├── migrations/
│   └── .gitkeep
├── .env
├── .air.toml
├── Makefile
├── go.mod
└── docker-compose.yml
```

---

## 各檔案內容模板

### 1. `cmd/server/main.go`

```go
package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/{service}"
	"github.com/darkphotonKN/cosmic-void-server/common/broker"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery/consul"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/{service-name}/config"
	"github.com/darkphotonKN/cosmic-void-server/{service-name}/internal/{service}"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

var (
	serviceName  = "{service}s"
	grpcAddr     = commonhelpers.GetEnvString("GRPC_{SERVICE_UPPER}_ADDR", "{GRPC_PORT}")
	consulAddr   = commonhelpers.GetEnvString("CONSUL_ADDR", "localhost:8510")
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
	repo := {service}.NewRepository(db)
	svc := {service}.NewService(repo, ch)
	handler := {service}.NewHandler(svc)

	pb.Register{Service}ServiceServer(grpcServer, handler)

	log.Printf("gRPC {Service} Server started on PORT: %s\n", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Can't connect to grpc server. Error:", err.Error())
	}
}
```

### 2. `config/db.go`

```go
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitDB() *sqlx.DB {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")
	return db
}
```

### 3. `internal/{service}/repository.go`

```go
package {service}

import "github.com/jmoiron/sqlx"

type Repository interface {
	// 定義 repository 方法
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}
```

### 4. `internal/{service}/service.go`

```go
package {service}

import "github.com/rabbitmq/amqp091-go"

type Service interface {
	// 定義 service 方法
}

type service struct {
	repo      Repository
	publishCh *amqp.Channel
}

func NewService(repo Repository, ch *amqp.Channel) Service {
	return &service{
		repo:      repo,
		publishCh: ch,
	}
}
```

### 5. `internal/{service}/handler.go`

```go
package {service}

import (
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/{service}"
)

type Handler struct {
	service Service
	pb.Unimplemented{Service}ServiceServer
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}
```

### 6. `internal/{service}/model.go`

```go
package {service}

import (
	"time"

	"github.com/google/uuid"
)

type {Service} struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
```

### 7. `.env`

```properties
# Database
DB_HOST=localhost
DB_PORT={DB_PORT}
DB_USER=user
DB_PASSWORD=password
DB_NAME=cosmic_void_{service}_service_db

# gRPC
GRPC_{SERVICE_UPPER}_ADDR={GRPC_PORT}

# RabbitMQ
RABBITMQ_USER=cosmicvoid
RABBITMQ_PASS=cosmicvoid
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5682

# Service Discovery
CONSUL_ADDR=localhost:8510
```

> **Port 計算規則**：創建新服務前，掃描專案內所有 `.env` 和 `docker-compose.yml`，找出最高的 gRPC Port 和 DB Port，各 +1 作為新服務的 port。

### 8. `docker-compose.yml`

```yaml
services:
  db:
    image: postgres:17
    container_name: cosmic_void_{service}_service_db
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: cosmic_void_{service}_service_db
    ports:
      - "{DB_PORT}:5432"
    volumes:
      - {service}_db_data:/var/lib/postgresql/data
    networks:
      - cosmic-void-network

volumes:
  {service}_db_data:

networks:
  cosmic-void-network:
    external: true
    name: cosmic-void-network
```

### 9. `go.mod`

```
module github.com/darkphotonKN/cosmic-void-server/{service-name}

go 1.23.0

require (
	github.com/darkphotonKN/cosmic-void-server/common v0.0.0
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/rabbitmq/amqp091-go v1.10.0
	google.golang.org/grpc v1.72.0
)

replace github.com/darkphotonKN/cosmic-void-server/common => ../common
```

### 10. `Makefile`

```makefile
.PHONY: dev build test migrate-up migrate-down migrate-create

dev:
	@air

build:
	go build -o bin/{service-name} cmd/server/main.go

test:
	@go test -v ./... -cover

migrate-up:
	@migrate -path ./migrations -database "postgres://user:password@localhost:{DB_PORT}/cosmic_void_{service}_service_db?sslmode=disable" up

migrate-down:
	@migrate -path ./migrations -database "postgres://user:password@localhost:{DB_PORT}/cosmic_void_{service}_service_db?sslmode=disable" down

migrate-create:
	migrate create -ext sql -dir ./migrations -seq $(NAME)
```

### 11. `.air.toml`

```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/server/main.go"
bin = "./tmp/main"
delay = 1000
exclude_dir = ["assets", "tmp", "vendor", "migrations"]
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_regex = ["_test.go"]
```

---

## 變數替換規則

| 變數 | 說明 | 範例 |
|-----|------|------|
| `{service-name}` | 服務目錄名（含 -service） | `inventory-service` |
| `{service}` | 服務名（不含 -service） | `inventory` |
| `{Service}` | 服務名首字母大寫 | `Inventory` |
| `{SERVICE_UPPER}` | 服務名全大寫 | `INVENTORY` |
| `{GRPC_PORT}` | gRPC 端口 | `7022` |
| `{DB_PORT}` | 資料庫端口 | `5112` |

---

## 額外步驟

1. **更新 `go.work`**：加入新服務
   ```
   use (
       ...
       ./{service-name}
   )
   ```

2. **產生 Proto 檔案**（如需要）：
   ```
   common/api/proto/{service}/{service}.proto
   ```

3. **更新 `common/Makefile`**：加入新的 proto 編譯指令

---

## 驗證方式

1. 進入新服務目錄：`cd {service-name}`
2. 啟動資料庫：`docker-compose up -d`
3. 執行服務：`make dev` 或 `air`
4. 確認 gRPC 服務啟動：查看 log 顯示 port
5. 確認 Consul 註冊：訪問 `http://localhost:8510` 查看服務列表
