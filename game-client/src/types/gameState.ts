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
  id: UUID; // Player's permanent user ID (from signup)
  entity_id: UUID; // Temporary entity ID in game session
  username: string;
  position: Position;
  direction: PlayerDirection;
  inventory?: ItemState[]; // 玩家背包
  escape: boolean;
}

// Door/interactable state
export interface DoorState {
  entity_id: UUID;
  position: Position;
  width: number;
  height: number;
  is_open: boolean;
}

// Item state
export interface ItemState {
  item_id: UUID;
  entity_id: UUID;
  name: string;
  quantity: number;
  attack_power?: number;
  critical_rate?: number;
  weapon_type?: string;
  defense_rating?: number;
  armor_slot?: string;
  healing_amount?: number;
  mana_amount?: number;
  description?: string;
  lootedAt?: number; // 本地取得時間戳，用於 pending 判斷
}

// Container/chest state
export interface ContainerState {
  container_id: UUID;
  entity_id: UUID;
  position: Position;
  is_open: boolean;
  items: ItemState[];
}

// Escape door state
export interface EscapeDoorState {
  entity_id: UUID;
  position: Position;
  is_open: boolean;
  is_locked: boolean;
}

// Switch/button state
export interface SwitchState {
  entity_id: UUID;
  position: Position;
  switch_id: number;
  is_activated: boolean;
}

// Wall state
export interface WallState {
  entity_id: UUID;
  position: Position;
  width: number;
  height: number;
}

// Complete game state received from server
export interface ClientGameState {
  session_id: UUID;
  current_player: PlayerState | null; // This client's player state
  other_players: PlayerState[]; // Other players in session
  items: string[]; // TODO: Update when items are structured
  doors: DoorState[];
  walls: WallState[];
  containers: ContainerState[];
  escape_doors: EscapeDoorState[]; // Escape doors with lock state
  switches: SwitchState[]; // Switches/buttons for puzzles
}

// Type guard to check if a message is a game state update
export function isGameState(data: any): data is ClientGameState {
  return (
    data &&
    typeof data.session_id === "string" &&
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
