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
  equipment?: EquipmentState; // 玩家已裝備的 loadout
  escape: boolean;
}

// Player equipment state from server (matches Go types.EquipmentState JSON tags).
// Note: backend uses chest/gloves/legs; frontend UI uses body/hands/feet — map when consuming.
export interface EquipmentState {
  weapon: ItemState | null;
  head: ItemState | null;
  chest: ItemState | null;
  gloves: ItemState | null;
  legs: ItemState | null;
  ring_1: ItemState | null;
  ring_2: ItemState | null;
  consumable_1: ItemState | null;
  consumable_2: ItemState | null;
  consumable_3: ItemState | null;
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
  durability?: number;
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
  house_id?: UUID;
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
  escaped_count: number; // Number of players who have escaped
}

// Type guard to check if a message is a game state update
export function isGameState(data: any): data is ClientGameState {
  return (
    data &&
    typeof data.session_id === "string" &&
    (data.current_player !== undefined || data.other_players !== undefined)
  );
}

// Equipment types
export type EquipmentSlot = 'weapon' | 'head' | 'body' | 'hands' | 'feet' | 'ring_1' | 'ring_2' | 'consumable_1' | 'consumable_2' | 'consumable_3';

// Matches backend types.ArmorSlot — game-server/game-service/internal/types/game.go
export type ArmorSlot = 'head' | 'chest' | 'gloves' | 'legs';

export interface EquippedItems {
  weapon: ItemState | null;
  head: ItemState | null;
  body: ItemState | null;
  hands: ItemState | null;
  feet: ItemState | null;
  ring_1: ItemState | null;
  ring_2: ItemState | null;
  consumable_1: ItemState | null;
  consumable_2: ItemState | null;
  consumable_3: ItemState | null;
}

export type ItemType = 'weapon' | 'armor' | 'consumable' | 'unknown';

export function getItemType(item: ItemState): ItemType {
  if (item.attack_power || item.weapon_type) return 'weapon';
  if (item.defense_rating !== undefined || item.armor_slot) return 'armor';
  if (item.healing_amount || item.mana_amount) return 'consumable';
  return 'unknown';
}

export function getValidSlotsForItem(item: ItemState): EquipmentSlot[] {
  const type = getItemType(item);
  switch (type) {
    case 'weapon':
      return ['weapon'];
    case 'armor': {
      const slot = item.armor_slot as ArmorSlot | undefined;
      switch (slot) {
        case 'head': return ['head'];
        case 'chest': return ['body'];
        case 'gloves': return ['hands'];
        case 'legs': return ['feet'];
        default: return [];
      }
    }
    case 'consumable':
      return ['consumable_1', 'consumable_2', 'consumable_3'];
    default:
      return [];
  }
}

export function getSlotDisplayName(slot: EquipmentSlot): string {
  const names: Record<EquipmentSlot, string> = {
    weapon: 'Weapon',
    head: 'Head',
    body: 'Body',
    hands: 'Hands',
    feet: 'Feet',
    ring_1: 'Ring 1',
    ring_2: 'Ring 2',
    consumable_1: 'Consumable 1',
    consumable_2: 'Consumable 2',
    consumable_3: 'Consumable 3',
  };
  return names[slot];
}

// Helper to format position for display
export function formatPosition(pos: Position): string {
  return `(${pos.x.toFixed(1)}, ${pos.y.toFixed(1)})`;
}

// Helper to format velocity for display
export function formatVelocity(dir: PlayerDirection): string {
  return `(${dir.vx.toFixed(1)}, ${dir.vy.toFixed(1)})`;
}
