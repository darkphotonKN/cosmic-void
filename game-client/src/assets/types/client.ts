interface MovePayload {
  x: number;
  y: number;
}

interface AttackPayload {
  targetId: string;
}

interface PickupPayload {
  itemId: string;
}

interface UsePayload {
  itemId: string;
  targetId?: string; // 可選：對誰使用
}

interface ChatPayload {
  message: string;
}

// ====== 動作類型對應 Payload ======

interface ActionMap {
  move: MovePayload;
  attack: AttackPayload;
  pickup: PickupPayload;
  use: UsePayload;
  chat: ChatPayload;
}

const ActionType = {
  Move: "move",
  Attack: "attack",
  Pickup: "pickup",
  Use: "use",
  Chat: "chat",
} as const;

type ActionType = (typeof ActionType)[keyof typeof ActionType];

// ====== Client → Server 訊息（泛型版）======

interface ClientMessage<T extends keyof ActionMap> {
  action: T;
  payload: ActionMap[T];
  seq: number;
}

// ====== 或是用 Union Type（更直接）======

type ClientAction =
  | { action: "move"; payload: MovePayload; seq: number }
  | { action: "attack"; payload: AttackPayload; seq: number }
  | { action: "pickup"; payload: PickupPayload; seq: number }
  | { action: "use"; payload: UsePayload; seq: number }
  | { action: "chat"; payload: ChatPayload; seq: number };
