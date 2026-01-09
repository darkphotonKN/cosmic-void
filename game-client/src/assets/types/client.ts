<<<<<<< HEAD
// Base payload that all game actions must include
export interface PlayerSessionPayload {
  session_id: string;
  player_id: string;
=======
export interface gameInfo {
  session_id: string;
  player_id: string;
}
export interface MovePayload {
  vx: number;
  vy: number;
>>>>>>> 70c74326b4474a1d9777f47e2b9c1ba046940f80
}

// Individual action payloads (matching backend expectations)
export interface MovePayload extends PlayerSessionPayload {
  vx: number;  // velocity x (not position)
  vy: number;  // velocity y (not position)
}

export interface InteractPayload extends PlayerSessionPayload {
  entity_id: string;
}

export interface AttackPayload extends PlayerSessionPayload {
  target_id: string;
}

export interface PickupPayload extends PlayerSessionPayload {
  item_id: string;
}

export interface UsePayload extends PlayerSessionPayload {
  item_id: string;
  target_id?: string; // optional: who to use it on
}

export interface ChatPayload extends PlayerSessionPayload {
  message: string;
}

export interface FindGamePayload {
  player_id: string;  // This one doesn't need session_id since it's for finding a game
}

// ====== 動作類型對應 Payload ======

export interface ActionMap {
  move: MovePayload;
  interact: InteractPayload;
  attack: AttackPayload;
  pickup: PickupPayload;
  use: UsePayload;
  chat: ChatPayload;
  find_game: FindGamePayload;
}

export const ActionType = {
  Move: "move",
  Interact: "interact",
  Attack: "attack",
  Pickup: "pickup",
  Use: "use",
  Chat: "chat",
  Find_Game: "find_game",
} as const;

export type ActionType = (typeof ActionType)[keyof typeof ActionType];

// ====== Client → Server 訊息（泛型版）======

export interface ClientMessage<T extends keyof ActionMap> {
  action: T;
  payload: ActionMap[T];
  seq: number;
}

// ====== 或是用 Union Type（更直接）======

export type ClientAction =
  | { action: "move"; payload: MovePayload; seq: number }
  | { action: "interact"; payload: InteractPayload; seq: number }
  | { action: "attack"; payload: AttackPayload; seq: number }
  | { action: "pickup"; payload: PickupPayload; seq: number }
  | { action: "use"; payload: UsePayload; seq: number }
  | { action: "chat"; payload: ChatPayload; seq: number }
  | { action: "find_game"; payload: FindGamePayload; seq: number };
