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

// ====== 基本物件型別 ======

interface Position {
  x: number;
  y: number;
}

interface Treasure {
  id: string;
  x: number;
  y: number;
  type: "gold" | "silver";
  value: number;
}

interface Enemy {
  id: string;
  x: number;
  y: number;
  type: "goblin" | "skeleton";
  hp: number;
}

interface Player {
  id: string;
  x: number;
  y: number;
  hp: number;
  name: string;
}

interface Wall {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
}

interface Building {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
  type: "house" | "tower" | "ruins" | "shrine";
}

// ====== 玩家自己的狀態 ======

interface PlayerState {
  id: string;
  x: number;
  y: number;
  hp: number;
  maxHp: number;
  score: number;
  inventory: InventoryItem[];
  isIndoor: boolean;
  currentBuildingId: string | null;
}

interface InventoryItem {
  id: string;
  type: "potion" | "weapon" | "key";
  name: string;
}

// ====== 可見物件 ======

interface VisibleObjects {
  treasures: Treasure[];
  enemies: Enemy[];
  players: Player[];
  walls: Wall[];
  buildings: Building[];
}

// ====== 動作結果 ======

interface ActionResult<T extends ActionType = ActionType> {
  action: T;
  success: boolean;
  message?: string;
  data?: ActionResultData[T];
}

interface ActionResultData {
  move: null;
  attack: { damage: number; targetHp: number };
  pickup: { itemId: string; value: number };
  use: { effect: string };
  chat: null;
}

// ====== 事件 ======

type GameEvent =
  | { type: "damage_taken"; fromId: string; amount: number }
  | { type: "item_collected"; itemId: string; value: number }
  | { type: "enemy_died"; enemyId: string }
  | { type: "player_entered"; playerId: string; name: string }
  | { type: "player_left"; playerId: string };

// ====== Server → Client 完整回應 ======

interface ServerMessage {
  seq: number;
  timestamp: number;
  player: PlayerState;
  visible: VisibleObjects;
  result?: ActionResult;
  events?: GameEvent[];
}
