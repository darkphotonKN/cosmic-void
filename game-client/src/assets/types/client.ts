export interface gameInfo {
  session_id: string;
  player_id: string;
}
export interface MovePayload {
  vx: number;
  vy: number;
}

export interface AttackPayload {
  enemy_entity_id: string;
}

export interface PickupPayload {
  itemId: string;
}

export interface UsePayload {
  itemId: string;
  targetId?: string; // 可選：對誰使用
}

export interface ChatPayload {
  message: string;
}

export interface FindGamePayload {
  playerId: string;
}

export interface InteractPayload {
  entity_id: string;
}

export interface EquipPayload {
  item_entity_id: string;
}

export interface UnequipPayload {
  item_entity_id: string;
}

// ====== 動作類型對應 Payload ======

export interface ActionMap {
  move: MovePayload;
  attack: AttackPayload;
  pickup: PickupPayload;
  use: UsePayload;
  chat: ChatPayload;
  find_game: FindGamePayload;
  interact: InteractPayload;
  equip: EquipPayload;
  unequip: UnequipPayload;
}

export const ActionType = {
  Move: "move",
  Attack: "attack",
  Pickup: "pickup",
  Use: "use",
  Chat: "chat",
  Find_Game: "find_game",
  Interact: "interact",
  Equip: "equip",
  Unequip: "unequip",
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
  | { action: "attack"; payload: AttackPayload; seq: number }
  | { action: "pickup"; payload: PickupPayload; seq: number }
  | { action: "use"; payload: UsePayload; seq: number }
  | { action: "chat"; payload: ChatPayload; seq: number };
