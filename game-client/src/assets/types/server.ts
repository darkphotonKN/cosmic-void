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
