# Frontend Game State Handling Implementation Plan

## Overview
Implement TypeScript types and client-side message handling to receive, parse, and log game state updates from the Go server that broadcasts every second.

## Current State Analysis
- Server broadcasts `ClientGameState` every second via WebSocket
- Client has `SocketManager` but doesn't handle game state messages
- `TreasureHuntScene` uses direct WebSocket instead of `SocketManager`
- No TypeScript types matching server's game state structure

## Implementation Plan

### Phase 1: Type Definitions

#### 1. Create Game State Types
**File**: `game-client/src/types/gameState.ts` (NEW)

```typescript
// Matching Go uuid.UUID type
type UUID = string;

// Player position in 2D space
interface Position {
  x: number;
  y: number;
}

// Player movement direction
interface PlayerDirection {
  vx: number;
  vy: number;
  speed: number;
}

// Individual player state
interface PlayerState {
  id: UUID;           // Player's permanent user ID
  entity_id: UUID;    // Temporary entity ID in game session
  username: string;
  position: Position;
  direction: PlayerDirection;
}

// Door/interactable state
interface DoorState {
  entity_id: UUID;
  position: Position;
  is_open: boolean;
}

// Complete game state (matching backend structure)
interface ClientGameState {
  session_id: UUID;
  current_player: PlayerState;  // This player's state
  other_players: PlayerState[]; // Other players in session
  items: string[];              // TODO: Update when items are structured
  doors: DoorState[];
}

// Server message wrapper (if backend wraps the state)
interface GameStateMessage {
  action?: string;
  payload?: any;
  // Direct state if not wrapped
  session_id?: UUID;
  current_player?: PlayerState;
  other_players?: PlayerState[];
}
```

### Phase 2: Message Handling

#### 2. Update SocketManager
**File**: `game-client/src/utils/class/SocketManager.ts` (MODIFY)

Add specialized game state handling:
```typescript
import { ClientGameState } from '@/types/gameState';

class SocketManager {
  private gameStateListeners: Set<(state: ClientGameState) => void> = new Set();

  // Add method to subscribe to game state updates
  onGameStateUpdate(callback: (state: ClientGameState) => void): () => void {
    this.gameStateListeners.add(callback);
    return () => this.gameStateListeners.delete(callback);
  }

  // Update onmessage handler
  socket.onmessage = (event) => {
    const data = JSON.parse(event.data);

    // Check if it's a game state update (has session_id and players)
    if (this.isGameState(data)) {
      this.handleGameStateUpdate(data as ClientGameState);
    } else if (data.action && this.listeners.has(data.action)) {
      this.listeners.get(data.action)?.(data.payload);
    }
  };

  private isGameState(data: any): boolean {
    return data.session_id &&
           (data.current_player || data.other_players || data.Players);
  }

  private handleGameStateUpdate(state: ClientGameState): void {
    // Enhanced logging
    this.logGameState(state);

    // Notify all listeners
    this.gameStateListeners.forEach(listener => listener(state));
  }
}
```

#### 3. Console Logging Implementation
**File**: `game-client/src/utils/gameStateLogger.ts` (NEW)

```typescript
export class GameStateLogger {
  private static updateCount = 0;

  static logGameState(state: ClientGameState): void {
    const timestamp = new Date().toLocaleTimeString();
    this.updateCount++;

    console.group(
      `%c[Game State Update #${this.updateCount}] - ${timestamp}`,
      'color: #4ecca3; font-weight: bold; font-size: 14px'
    );

    console.log('%cSession ID:', 'color: #ffd700', state.session_id);

    // Current player info
    if (state.current_player) {
      console.group('%c👤 Current Player', 'color: #00ff00; font-weight: bold');
      console.table({
        'Username': state.current_player.username,
        'Player ID': state.current_player.id,
        'Entity ID': state.current_player.entity_id,
        'Position': `(${state.current_player.position.x.toFixed(1)}, ${state.current_player.position.y.toFixed(1)})`,
        'Velocity': `(${state.current_player.direction.vx.toFixed(1)}, ${state.current_player.direction.vy.toFixed(1)})`,
        'Speed': state.current_player.direction.speed
      });
      console.groupEnd();
    }

    // Other players
    if (state.other_players?.length > 0) {
      console.group(`%c👥 Other Players (${state.other_players.length})`, 'color: #00bfff; font-weight: bold');
      state.other_players.forEach(player => {
        console.log(
          `  ${player.username}: (${player.position.x.toFixed(1)}, ${player.position.y.toFixed(1)})`
        );
      });
      console.groupEnd();
    }

    // Items and doors summary
    console.log(`%c📦 Items: ${state.items?.length || 0}`, 'color: #ff69b4');
    console.log(`%c🚪 Doors: ${state.doors?.length || 0}`, 'color: #dda0dd');

    // Full state for debugging
    console.group('%c📊 Full State Object', 'color: #808080; font-size: 10px');
    console.log(state);
    console.groupEnd();

    console.groupEnd();
  }
}
```

### Phase 3: Scene Integration

#### 4. Update TreasureHuntScene
**File**: `game-client/src/scenes/TreasureHuntScene.ts` (MODIFY)

Replace direct WebSocket with SocketManager:
```typescript
import { socketManager } from '@/utils/class/SocketManager';
import { ClientGameState } from '@/types/gameState';

export class TreasureHuntScene extends Phaser.Scene {
  private gameStateUnsubscribe?: () => void;
  private currentGameState?: ClientGameState;

  create(): void {
    // Connect via SocketManager instead of direct WebSocket
    if (!socketManager.isConnected()) {
      socketManager.connect('ws://localhost:5555/game/ws');
    }

    // Subscribe to game state updates
    this.gameStateUnsubscribe = socketManager.onGameStateUpdate(
      (state: ClientGameState) => {
        this.handleGameStateUpdate(state);
      }
    );

    // ... rest of create logic
  }

  private handleGameStateUpdate(state: ClientGameState): void {
    this.currentGameState = state;

    // For now, just store the state
    // Phase 2 will implement actual game updates

    // Update status display
    if (state.current_player) {
      const pos = state.current_player.position;
      this.updateStatus(
        `Connected | You: (${pos.x.toFixed(0)}, ${pos.y.toFixed(0)})`,
        '#4ecca3'
      );
    }
  }

  shutdown(): void {
    // Clean up subscription
    this.gameStateUnsubscribe?.();
    super.shutdown();
  }
}
```

## File Structure Summary

```
game-client/
├── docs/
│   └── frontend-game-state-handling.md (THIS FILE)
├── src/
│   ├── types/
│   │   └── gameState.ts (NEW)
│   ├── utils/
│   │   ├── class/
│   │   │   └── SocketManager.ts (MODIFY)
│   │   └── gameStateLogger.ts (NEW)
│   └── scenes/
│       └── TreasureHuntScene.ts (MODIFY)
```

## Testing Plan

1. **Connection Test**: Verify SocketManager connects successfully
2. **Message Reception**: Confirm state updates received every second
3. **Type Safety**: Ensure TypeScript types match server data
4. **Console Output**: Validate formatted logging displays correctly
5. **Memory**: Check for memory leaks from listeners

## Console Output Example

```
[Game State Update #1] - 14:23:45
Session ID: 550e8400-e29b-41d4-a716-446655440000
👤 Current Player
  Username: Player1
  Player ID: 123e4567-e89b-12d3-a456-426614174000
  Position: (100.0, 200.0)
  Velocity: (1.0, 0.0)
👥 Other Players (2)
  Player2: (150.0, 300.0)
  Player3: (200.0, 250.0)
📦 Items: 5
🚪 Doors: 2
```

## Next Steps (Phase 2)
After implementing message handling:
1. Update player positions based on state
2. Implement interpolation between updates
3. Add client-side prediction for current player
4. Handle state reconciliation
5. Add visual indicators for other players

## Notes
- Backend will separate `current_player` from `other_players` (see backend plan)
- Initial implementation focuses on logging/debugging
- Game rendering updates will be implemented after validation