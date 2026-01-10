# Stats Service Implementation Plan

## Overview

This document outlines the implementation plan for the stats-service microservice, focusing on creating the basic CRUD operations for player statistics tracking that matches our database schema.

## Current State Analysis

### Database Schema (Already Created)

We have 3 tables defined in `/stats-service/migrations/000001_create_stats_tables.up.sql`:

1. **player_match_stats** - Aggregated match statistics

   - `id` (UUID)
   - `member_id` (UUID, unique)
   - `games_played`, `wins`, `losses`, `kills`, `deaths`
   - `times_placed_top_three`
   - `created_at`, `updated_at`

2. **player_ranking_stats** - Leaderboard/ranking data

   - `id` (UUID)
   - `member_id` (UUID, unique)
   - `username` (denormalized for performance)
   - `wins`, `top_threes`
   - `rating` (ELO/MMR)
   - `rank_position` (nullable)
   - `last_calculated_at`, `created_at`, `updated_at`

3. **match_history** - Individual match records (event sourcing)
   - `id` (UUID)
   - `session_id`, `member_id`
   - `win` (boolean), `final_position`
   - `kills`, `deaths`
   - `rating_before`, `rating_after`, `rating_change`
   - `match_started_at`, `created_at`

### Current Proto Definition Issues

The existing `stats.proto` file has:

- Complex operations we don't need yet (GetLeaderboard, UpdatePlayerStats)
- Fields that don't exist in our database (assists, damage_dealt, damage_taken, xp, items_collected)
- Missing alignment with actual table structure

## Implementation Steps

### Phase 1: Proto Definition Update ✅ COMPLETED

**File**: `/game-server/common/api/proto/stats/stats.proto`

**Changes Made**:

- Simplified to only CREATE operations
- Removed complex querying methods
- Aligned message fields exactly with database columns
- Added proper comments for documentation

**New Service Methods**:

```protobuf
service StatsService {
  rpc CreatePlayerMatchStats(CreatePlayerMatchStatsRequest) returns (PlayerMatchStats);
  rpc CreatePlayerRankingStats(CreatePlayerRankingStatsRequest) returns (PlayerRankingStats);
  rpc CreateMatchHistory(CreateMatchHistoryRequest) returns (MatchHistory);
}
```

### Phase 2: Generate Proto Files (PENDING)

**Action**: Run proto generation

```bash
cd game-server/common
make proto  # or make generate depending on Makefile
```

This will generate:

- `stats.pb.go` - Message definitions
- `stats_grpc.pb.go` - gRPC service interfaces

### Phase 3: Update Model Layer (PENDING)

**File**: `/game-server/stats-service/internal/stats/model.go`

**Changes Required**:

1. Update struct definitions to match new proto exactly:
   - Remove non-existent fields (Assists, DamageDealt, etc.)
   - Add new models for `PlayerMatchStats`, `PlayerRankingStats`, `MatchHistory`
   - Ensure field tags match database columns

**Example Structure**:

```go
type PlayerMatchStats struct {
    ID                   uuid.UUID `db:"id"`
    MemberID            uuid.UUID `db:"member_id"`
    GamesPlayed         int32     `db:"games_played"`
    Wins                int32     `db:"wins"`
    Losses              int32     `db:"losses"`
    Kills               int32     `db:"kills"`
    Deaths              int32     `db:"deaths"`
    TimesPlacedTopThree int32     `db:"times_placed_top_three"`
    CreatedAt           time.Time `db:"created_at"`
    UpdatedAt           time.Time `db:"updated_at"`
}
```

### Phase 4: Implement Repository Layer (PENDING)

**File**: `/game-server/stats-service/internal/stats/repository.go`

**Implementation Pattern** (following auth-service pattern):

```go
type Repository struct {
    DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
    return &Repository{DB: db}
}

// Only CREATE methods, no updates or complex queries
func (r *Repository) CreatePlayerMatchStats(ctx context.Context, stats *PlayerMatchStats) error
func (r *Repository) CreatePlayerRankingStats(ctx context.Context, stats *PlayerRankingStats) error
func (r *Repository) CreateMatchHistory(ctx context.Context, history *MatchHistory) error
```

**Key Points**:

- Use prepared statements
- Return database errors properly wrapped
- Use context for cancellation
- Follow existing error handling patterns

### Phase 5: Implement Service Layer (PENDING)

**File**: `/game-server/stats-service/internal/stats/service.go`

**Interface Definition** (Consumer-owned):

```go
type Repository interface {
    CreatePlayerMatchStats(ctx context.Context, stats *PlayerMatchStats) error
    CreatePlayerRankingStats(ctx context.Context, stats *PlayerRankingStats) error
    CreateMatchHistory(ctx context.Context, history *MatchHistory) error
}

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}
```

**Methods to Implement**:

- Convert proto requests to domain models
- Call repository methods
- Convert domain models back to proto responses
- Handle business logic (if any)

### Phase 6: Update gRPC Handler (PENDING)

**File**: `/game-server/stats-service/internal/stats/handler.go`

**Current State**: File exists but needs updating

**Changes Required**:

1. Define Service interface (what handler needs):

```go
type Service interface {
    CreatePlayerMatchStats(ctx context.Context, req *pb.CreatePlayerMatchStatsRequest) (*pb.PlayerMatchStats, error)
    CreatePlayerRankingStats(ctx context.Context, req *pb.CreatePlayerRankingStatsRequest) (*pb.PlayerRankingStats, error)
    CreateMatchHistory(ctx context.Context, req *pb.CreateMatchHistoryRequest) (*pb.MatchHistory, error)
}
```

2. Implement gRPC server interface:

```go
type Handler struct {
    pb.UnimplementedStatsServiceServer
    service Service
}

func NewHandler(service Service) *Handler {
    return &Handler{service: service}
}
```

3. Implement each RPC method

### Phase 7: Update AMQP Consumer (PENDING)

**File**: `/game-server/stats-service/internal/stats/amqp_consumer.go`

**Current State**: Basic skeleton exists

**Required Implementation**:

1. Set up queue consumers for events:

   - `match.completed` - Create match history record
   - `player.stats.update` - Update aggregated stats

2. Message handling pattern:

```go
func (c *Consumer) consumeMatchCompleted() {
    msgs, err := c.channel.Consume(
        "stats.match.completed", // queue name
        "",                       // consumer
        true,                     // auto-ack
        false,                    // exclusive
        false,                    // no-local
        false,                    // no-wait
        nil,                      // args
    )

    for msg := range msgs {
        // Parse message
        // Call service method
        // Log result
    }
}
```

### Phase 8: Wire Dependencies (PENDING)

**File**: `/game-server/stats-service/config/routes.go`

**Setup Required**:

```go
// Create repository
repo := stats.NewRepository(db)

// Create service
service := stats.NewService(repo)

// Create handler
handler := stats.NewHandler(service)

// Create AMQP consumer
consumer := stats.NewConsumer(service, amqpChannel)
consumer.Listen()

// Register gRPC server
pb.RegisterStatsServiceServer(grpcServer, handler)
```

## Testing Strategy

### Unit Tests

- Repository: Mock database calls
- Service: Mock repository interface
- Handler: Mock service interface

### Integration Tests

- Test full flow with real database (test DB)
- Verify proto contract compliance
- Test AMQP message handling

## Rollback Plan

If issues arise:

1. Revert proto changes
2. Regenerate old proto files
3. Revert code changes
4. Restart services

## Success Criteria

- [x] Proto file simplified and aligned with DB schema
- [ ] Proto files successfully generated
- [ ] All 3 create methods working via gRPC
- [ ] AMQP consumer processing events
- [ ] No compilation errors
- [ ] Existing services unaffected

## Notes & Considerations

### Why Only CREATE Methods?

- Following YAGNI principle (You Aren't Gonna Need It)
- Simplifies initial implementation
- Updates can be added later when requirements are clear
- Reduces complexity and potential bugs

### Interface Segregation Principle (ISP)

- Handler defines what it needs from Service
- Service defines what it needs from Repository
- No fat interfaces with unused methods

### Future Enhancements (NOT in this phase)

- Update methods for existing records
- Complex queries (leaderboards, analytics)
- Batch operations
- Caching layer
- Rate limiting

## Review Checklist

Before proceeding with implementation:

- [ ] Proto changes align with database schema
- [ ] No unnecessary complexity added
- [ ] Follows existing project patterns
- [ ] Dependencies properly injected
- [ ] Error handling consistent with other services
- [ ] Logging uses slog (not log package)
