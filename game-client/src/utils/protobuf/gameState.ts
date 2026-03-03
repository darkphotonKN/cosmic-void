/**
 * Protobuf definitions for game state
 * This file handles the loading and decoding of protobuf messages from the game server
 */

import protobuf from 'protobufjs';

// Define TypeScript interfaces matching the protobuf schema
export interface Position {
  x: number;
  y: number;
}

export interface PlayerDirection {
  vx: number;
  vy: number;
  speed: number;
}

export interface ItemState {
  itemId: Uint8Array;
  entityId: Uint8Array;
  name: string;
  quantity: number;
  attackPower?: number;
  durability?: number;
  criticalRate?: number;
  weaponType?: string;
  defenseRating?: number;
  armorSlot?: string;
  healingAmount?: number;
  manaAmount?: number;
  description?: string;
}

export interface PlayerState {
  id: Uint8Array;
  entityId: Uint8Array;
  username: string;
  position: Position;
  direction: PlayerDirection;
  inventory: ItemState[];
}

export interface DoorState {
  entityId: Uint8Array;
  position: Position;
  isOpen: boolean;
}

export interface ContainerState {
  containerId: Uint8Array;
  entityId: Uint8Array;
  position: Position;
  isOpen: boolean;
  items: ItemState[];
}

export interface ClientGameState {
  sessionId: Uint8Array;
  currentPlayer: PlayerState | null;
  otherPlayers: PlayerState[];
  items: Uint8Array[];
  doors: DoorState[];
  containers: ContainerState[];
}

// Helper to convert Uint8Array (UUID bytes) to string
export function uuidBytesToString(bytes: Uint8Array): string {
  if (!bytes || bytes.length !== 16) {
    return '00000000-0000-0000-0000-000000000000';
  }

  const hex = Array.from(bytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');

  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20, 32)}`;
}

// Protobuf root and message types (will be initialized on first load)
let protobufRoot: protobuf.Root | null = null;
let ClientGameStateType: protobuf.Type | null = null;

/**
 * Initialize the protobuf schema
 * This should be called once when the app starts
 */
export async function initProtobuf(): Promise<void> {
  if (protobufRoot) {
    return; // Already initialized
  }

  try {
    // Load the proto file
    protobufRoot = await protobuf.load('/proto/game_state.proto');
    ClientGameStateType = protobufRoot.lookupType('game.ClientGameState');
    console.log('Protobuf schema loaded successfully');
  } catch (error) {
    console.error('Failed to load protobuf schema:', error);
    throw error;
  }
}

/**
 * Decode binary protobuf data into a ClientGameState object
 * @param data - Binary data from WebSocket
 * @returns Decoded game state
 */
export function decodeGameState(data: ArrayBuffer): ClientGameState {
  if (!ClientGameStateType) {
    throw new Error('Protobuf not initialized. Call initProtobuf() first.');
  }

  try {
    // Convert ArrayBuffer to Uint8Array
    const bytes = new Uint8Array(data);

    // Decode the message
    const message = ClientGameStateType.decode(bytes);

    // Convert to plain object
    const obj = ClientGameStateType.toObject(message, {
      longs: String,
      enums: String,
      bytes: Array, // Keep as Array instead of Buffer
      defaults: true,
    });

    return obj as unknown as ClientGameState;
  } catch (error) {
    console.error('Failed to decode protobuf message:', error);
    throw error;
  }
}

/**
 * Check if protobuf is initialized
 */
export function isProtobufReady(): boolean {
  return protobufRoot !== null && ClientGameStateType !== null;
}
