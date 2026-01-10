# Game Server Architecture

## Overview

The game server follows a **microservices architecture** with domain-driven design principles. Each service is isolated with its own database and communicates through well-defined interfaces using gRPC and message queues.

## Core Architecture Principles

- **Domain-Driven Design (DDD Lite)**: Each service owns its domain logic and data
- **Interface Segregation Principle (ISP)**: Consumers define their own interfaces
- **Dependency Inversion Principle (DIP)**: Depend on abstractions, not concrete implementations
- **Database per Service**: Each service has isolated data storage
- **Event-Driven Communication**: Services communicate through events via RabbitMQ

## Service Architecture

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Client    │      │   Client    │      │   Client    │
└─────┬───────┘      └─────┬───────┘      └─────┬───────┘
      │                    │                    │
      └────────────────────┼────────────────────┘
                          │
                    WebSocket/HTTP
                          │
                   ┌──────▼──────┐
                   │ API Gateway │
                   └──────┬──────┘
                          │
                        gRPC
      ┌───────────────────┼───────────────────┐
      │                   │                   │
┌─────▼─────┐      ┌─────▼─────┐      ┌─────▼─────┐
│   Auth    │      │   Game    │      │   Stats   │
│  Service  │      │  Service  │      │  Service  │
└─────┬─────┘      └─────┬─────┘      └─────┬─────┘
      │                   │                   │
      │              WebSocket                │
      │              (Real-time)              │
      │                   │                   │
┌─────▼─────┐      ┌─────▼─────┐      ┌─────▼─────┐
│ PostgreSQL│      │ PostgreSQL│      │ PostgreSQL│
└───────────┘      └───────────┘      └───────────┘
                          │
                     RabbitMQ
                   (Event Broker)
```

## Service Descriptions

### 1. API Gateway
- **Purpose**: Single entry point for all client requests
- **Technology**: Go, HTTP/WebSocket
- **Responsibilities**:
  - Request routing
  - Authentication validation
  - Protocol translation (HTTP → gRPC)
  - WebSocket proxy for game connections

### 2. Auth Service
- **Purpose**: User authentication and authorization
- **Technology**: Go, gRPC, PostgreSQL
- **Domain Structure**: Standard DDD
  ```
  internal/
  ├── auth/       # Authentication logic
  └── member/     # User management
      ├── model.go      # Domain entities
      ├── repository.go # Data access
      ├── service.go    # Business logic
      └── handler.go    # gRPC handlers
  ```
- **Key Features**:
  - JWT token generation/validation
  - Password hashing (bcrypt)
  - Member CRUD operations

### 3. Game Service (Unique Architecture)
- **Purpose**: Real-time multiplayer game server
- **Technology**: Go, WebSocket, ECS Pattern
- **Architecture**: Entity Component System (ECS)
  ```
  internal/
  ├── components/   # Pure data structures
  ├── systems/      # Game logic (movement, interaction)
  ├── ecs/          # Core ECS framework
  ├── game/         # Game session management
  ├── gameserver/   # WebSocket server & client handling
  ├── messaging/    # Message routing & broadcasting
  ├── queue/        # Matchmaking queue
  └── serializer/   # State serialization
  ```

#### ECS Architecture Details
- **Entities**: Unique IDs representing game objects
- **Components**: Pure data containers (Position, Health, Inventory)
- **Systems**: Stateless processors that act on components
- **Game Loop**: Fixed timestep at 30 FPS
- **State Management**: Authoritative server with client prediction

#### Key Game Service Components:
- **Session Manager**: Handles game instances
- **Entity Manager**: Manages ECS entities and components
- **Message Sender**: Broadcasts game state to clients
- **Queue Service**: Player matchmaking
- **Serializer**: Optimizes state for network transmission

### 4. Stats Service
- **Purpose**: Track player statistics and match history
- **Technology**: Go, gRPC, PostgreSQL
- **Domain Structure**: Standard DDD
  ```
  internal/
  └── stats/
      ├── model.go          # Stats entities
      ├── repository.go     # Database operations
      ├── service.go        # Business logic
      ├── handler.go        # gRPC handlers
      └── amqp_consumer.go  # Event consumer
  ```
- **Database Tables**:
  - `player_match_stats`: Aggregated player statistics
  - `player_ranking_stats`: Leaderboard data
  - `match_history`: Individual match records

## Communication Patterns

### 1. Synchronous Communication (gRPC)
- Used for request-response patterns
- Service-to-service calls
- Strong typing via Protocol Buffers
- Example: Auth validation, user lookups

### 2. Asynchronous Communication (RabbitMQ)
- Event-driven architecture
- Service decoupling
- Event examples:
  - `member.signed_up`
  - `match.completed`
  - `stats.updated`

### 3. Real-time Communication (WebSocket)
- Game state broadcasting
- Client-server game messages
- Low latency requirements
- Binary message format for efficiency

## Protocol Buffers (Proto)

All service contracts are defined in `/common/api/proto/`:
```
common/api/proto/
├── common/     # Shared message types
├── auth/       # Auth service definitions
├── game/       # Game service definitions
└── stats/      # Stats service definitions
```

**Generation**: Run `make proto` in `/game-server/common/`

## Database Design

Each service owns its database schema:

### Auth Service
- `members`: User accounts and credentials
- `sessions`: Active user sessions

### Game Service
- No persistent storage (in-memory state)
- Future: match replays, world state

### Stats Service
- `player_match_stats`: Win/loss, K/D ratios
- `player_ranking_stats`: ELO ratings, leaderboards
- `match_history`: Individual match records

## Development Workflow

### Standard Microservice Pattern
1. Define proto contract
2. Generate code: `make proto`
3. Implement layers:
   - Model (entities)
   - Repository (data access)
   - Service (business logic)
   - Handler (gRPC/HTTP endpoints)
4. Wire dependencies in `config/routes.go`
5. Add migrations if needed

### Interface Design Rules
- **Consumer owns the interface**: Handler defines what it needs from Service
- **Narrow interfaces**: Only expose required methods
- **Dependency injection**: Pass interfaces to constructors
- **No circular dependencies**: Use events for cross-domain communication

## Configuration

### Environment Variables
Each service uses environment-specific configs:
- Database connections
- Service ports
- RabbitMQ settings
- JWT secrets

### Service Discovery
- Currently: Hard-coded service addresses
- Future: Consul for dynamic discovery

## Deployment

### Local Development
```bash
# Start infrastructure
docker-compose up -d  # PostgreSQL, RabbitMQ

# Run services
make dev  # Uses air for hot-reload
```

### Production
- Docker containers per service
- Kubernetes orchestration (planned)
- Health checks and monitoring
- Graceful shutdown handling

## Testing Strategy

### Unit Tests
- Test files adjacent to implementation
- Table-driven tests
- Mock interfaces for isolation

### Integration Tests
- Test full service flows
- Use test databases
- Verify gRPC contracts

### E2E Tests
- Full system testing
- WebSocket game flows
- Performance benchmarks

## Monitoring & Observability

### Structured Logging (slog)
- Consistent log format
- Contextual information
- Log levels: Debug, Info, Warn, Error

### Metrics (Future)
- Prometheus metrics
- Service latency
- Game performance stats
- Player activity

### Tracing (Future)
- OpenTelemetry integration
- Distributed request tracing
- Performance bottleneck identification

## Security Considerations

- JWT-based authentication
- Password hashing (bcrypt)
- Input validation at service boundaries
- Rate limiting (planned)
- DDoS protection (planned)

## Future Enhancements

1. **Service Mesh**: Istio for advanced traffic management
2. **Caching Layer**: Redis for session and game state
3. **Analytics Pipeline**: Real-time game analytics
4. **AI Services**: Bot players, matchmaking improvements
5. **CDN Integration**: Asset delivery optimization