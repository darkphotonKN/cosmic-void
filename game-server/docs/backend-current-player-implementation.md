# Backend Current Player Implementation Plan

## Overview

Modify the game state serialization to differentiate between the current player and other players, allowing clients to identify which player entity represents them.

## Problem Statement

Currently, `ClientGameState` broadcasts the same state to all players without identifying which player is the recipient. Clients need to know:

- Their own player ID (permanent user ID from signup)
- Which entity in the game state represents them

## Proposed Solution

Use a **per-player serialization** approach where each player receives a personalized view of the game state with `current_player` separated from `other_players`.

## Implementation Changes

### 1. Update `types.ClientGameState` Structure

**File**: `game-server/game-service/internal/types/messages.go`

```go
type ClientGameState struct {
    SessionID     uuid.UUID      `json:"session_id"`
    CurrentPlayer *PlayerState   `json:"current_player"` // NEW: The recipient's player state
    OtherPlayers  []*PlayerState `json:"other_players"`  // NEW: All other players
    Items         []string       `json:"items"`
    Doors         []*DoorState   `json:"doors"`
}
```

**Rationale**: Separating current player from others follows game development best practices:

- Clear ownership identification
- Easier client-side processing
- Supports future features like different rendering for self vs others
- Enables client-side prediction for current player only

### 2. Modify StateSerializer

**File**: `game-server/game-service/internal/serializer/state_serializer.go`

Change signature to accept the recipient's player ID:

```go
func (s *StateSerializer) Serialize(sessionID uuid.UUID, recipientPlayerID uuid.UUID, entities []*ecs.Entity) (*types.ClientGameState, error)
```

Implementation:

- Identify which entity belongs to `recipientPlayerID`
- Separate into `CurrentPlayer` and `OtherPlayers`
- Maintain existing serialization logic for each player

### 3. Update Session Broadcasting

**File**: `game-server/game-service/internal/game/session.go`

Modify `broadcastFullState()` to create personalized states:

```go
func (s *Session) broadcastFullState() error {
    entities := s.EntityManager.GetAllEntities()

    // Create personalized state for each player
    for entityID, playerID := range s.playerEntityIDToPlayerID {
        clientState, err := s.stateSerializer.Serialize(s.ID, playerID, entities)
        if err != nil {
            continue
        }
        s.sender.SendStateToPlayer(playerID, clientState)
    }
    return nil
}
```

## Benefits of Proposed Approach

1. **Clear Separation**: Client immediately knows which player is theirs
2. **Performance**: No client-side searching required
3. **Future-Proof**: Supports features like:
   - Different update rates for self vs others
   - Client-side prediction
   - Lag compensation
   - Privacy features (hiding certain data from other players)
4. **Industry Standard**: Most multiplayer games separate current player state

## Testing Considerations

- Verify each player receives correct `current_player` data
- Test with multiple players in same session
- Ensure state consistency across all clients
- Validate player ID mapping correctness

## Migration Path

1. Update types first (backwards compatible by keeping Players field temporarily)
2. Update serializer to populate both old and new fields
3. Update clients to use new fields
4. Remove old Players field in future version

## Next Steps

After implementing backend changes:

1. Update frontend types to match new structure
2. Modify client message handling
3. Update game rendering to use separated player data
