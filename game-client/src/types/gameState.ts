// Game state types matching Go server structures

// Matching Go uuid.UUID type
export type UUID = string;

// Player position in 2D space
export interface Position {
  x: number;
  y: number;
}

// Player movement direction and velocity
export interface PlayerDirection {
  vx: number;
  vy: number;
  speed: number;
}

// Individual player state
export interface PlayerState {
  id: UUID;           // Player's permanent user ID (from signup)
  entity_id: UUID;    // Temporary entity ID in game session
  username: string;
  position: Position;
  direction: PlayerDirection;
}

// Door/interactable state
export interface DoorState {
  entity_id: UUID;
  position: Position;
  is_open: boolean;
}

// Item state
export interface ItemState {
  item_id: UUID;
  entity_id: UUID;
  name: string;
  quantity: number;
}

// Container/chest state
export interface ContainerState {
  container_id: UUID;
  entity_id: UUID;
  position: Position;
  is_open: boolean;
  items: ItemState[];
}

// Complete game state received from server
export interface ClientGameState {
  session_id: UUID;
  current_player: PlayerState | null;  // This client's player state
  other_players: PlayerState[];        // Other players in session
  items: string[];                     // TODO: Update when items are structured
  doors: DoorState[];
  containers: ContainerState[];
}

// Type guard to check if a message is a game state update
export function isGameState(data: any): data is ClientGameState {
  return (
    data &&
    typeof data.session_id === 'string' &&
    (data.current_player !== undefined || data.other_players !== undefined)
  );
}

// Helper to format position for display
export function formatPosition(pos: Position): string {
  return `(${pos.x.toFixed(1)}, ${pos.y.toFixed(1)})`;
}

// Helper to format velocity for display
export function formatVelocity(dir: PlayerDirection): string {
  return `(${dir.vx.toFixed(1)}, ${dir.vy.toFixed(1)})`;
}