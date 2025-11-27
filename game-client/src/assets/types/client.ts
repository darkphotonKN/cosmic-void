export interface MovePayload {
  x: number;
  y: number;
}

export interface AttackPayload {
  targetId: string;
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

// ====== 動作類型對應 Payload ======

export interface ActionMap {
  move: MovePayload;
  attack: AttackPayload;
  pickup: PickupPayload;
  use: UsePayload;
  chat: ChatPayload;
}

export const ActionType = {
  Move: "move",
  Attack: "attack",
  Pickup: "pickup",
  Use: "use",
  Chat: "chat",
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
