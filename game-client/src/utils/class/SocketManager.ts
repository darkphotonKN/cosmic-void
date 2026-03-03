import { ActionMap, ClientMessage } from "@/assets/types/client";
import { ClientGameState, isGameState } from "@/types/gameState";
import { GameStateLogger } from "@/utils/gameStateLogger";
import { useGameStore } from "@/stores/gameStore";
import { useAuthStore } from "@/stores/authStore";
import { decodeGameState, initProtobuf, isProtobufReady, uuidBytesToString } from "@/utils/protobuf/gameState";
import type { ClientGameState as ProtobufGameState } from "@/utils/protobuf/gameState";

export type ConnectionStatus =
  | "disconnected"
  | "connecting"
  | "connected"
  | "error";

// SocketManager.js
class SocketManager {
  private socket: WebSocket | null = null;
  private listeners: Map<string, (data: any) => void> = new Map();
  // Status callback
  private onStatusChange?: (status: string, color: string) => void;
  // Auth error callback
  private onAuthError?: () => void;
  // Connection status
  private connectionStatus: ConnectionStatus = "disconnected";
  private connectionStatusListeners: Set<(status: ConnectionStatus) => void> =
    new Set();
  private seq: number = 0;
  // Game state listeners
  private gameStateListeners: Set<(state: ClientGameState) => void> = new Set();

  constructor() {
    this.socket = null;
    this.listeners = new Map();
    // Initialize protobuf on construction
    initProtobuf().catch(err => {
      console.error('Failed to initialize protobuf:', err);
    });
  }

  getConnectionStatus(): ConnectionStatus {
    return this.connectionStatus;
  }

  isConnected(): boolean {
    return this.connectionStatus === "connected";
  }

  onConnectionStatusChange(
    callback: (status: ConnectionStatus) => void,
  ): () => void {
    this.connectionStatusListeners.add(callback);
    // 立即觸發一次當前狀態
    callback(this.connectionStatus);
    // 返回取消訂閱函數
    return () => {
      this.connectionStatusListeners.delete(callback);
    };
  }

  private setConnectionStatus(status: ConnectionStatus): void {
    this.connectionStatus = status;
    this.connectionStatusListeners.forEach((listener) => listener(status));
  }

  setOnAuthError(callback: () => void) {
    this.onAuthError = callback;
  }

  connect(url: string) {
    if (this.socket) return; // 避免重複連接

    this.setConnectionStatus("connecting");
    this.socket = new WebSocket(url);

    this.socket.onopen = () => {
      console.log("WebSocket connected");
      this.setConnectionStatus("connected");
      this.updateStatus("WebSocket Connected", "#4ecca3");
    };

    this.socket.onerror = (error) => {
      console.error("WebSocket error:", error);
      this.setConnectionStatus("error");
      this.updateStatus("WebSocket Error", "#ff4444");
      // Don't trigger auth error on WebSocket errors - we handle these separately
      // if (this.onAuthError) {
      //   this.onAuthError();
      // }
    };

    this.socket.onclose = (event) => {
      console.log("WebSocket disconnected, code:", event.code);
      this.setConnectionStatus("disconnected");
      this.updateStatus("WebSocket Disconnected", "#ffcc00");
      this.socket = null;
    };

    this.socket.onmessage = (event) => {
      try {
        // Check if message is binary (Protobuf) or text (JSON)
        if (event.data instanceof ArrayBuffer || event.data instanceof Blob) {
          // Binary message - Protobuf game state
          this.handleBinaryMessage(event.data);
        } else {
          // Text message - JSON (legacy or non-game-state messages)
          const data = JSON.parse(event.data);

          // Check if it's a game state update (JSON format - legacy)
          if (isGameState(data)) {
            this.handleGameStateUpdate(data);
          } else if (data.action === "game_found") {
            // Store session_id when game is found
            const sessionId = data.payload?.session_id;
            if (sessionId) {
              useGameStore.getState().setSessionId(sessionId);
              console.log("Game found, session_id:", sessionId);
            }
            // Also notify listeners
            this.listeners.get(data.action)?.(data.payload);
          } else if (data.action && this.listeners.has(data.action)) {
            // Handle action-based messages
            console.log("Received action message:", data.action);
            this.listeners.get(data.action)?.(data.payload);
          } else {
            // Log unhandled messages for debugging
            console.log("Received unhandled message:", data);
          }
        }
      } catch (e) {
        console.error("Failed to parse message:", e);
        GameStateLogger.logError("Failed to parse WebSocket message", e);
      }
    };
  }

  private updateStatus(status: string, color: string): void {
    if (this.onStatusChange) {
      this.onStatusChange(status, color);
    }
  }

  // websocket send message
  sendMessage<T extends keyof ActionMap>(
    action: T,
    payload: ActionMap[T],
  ): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      // Auto-inject session_id and player_id
      const sessionId = useGameStore.getState().sessionId;
      const playerId = useAuthStore.getState().memberInfo?.id;

      const enrichedPayload = {
        ...payload,
        session_id: sessionId,
        player_id: playerId,
      };

      const message: ClientMessage<T> = {
        action,
        payload: enrichedPayload,
        seq: ++this.seq,
      };

      // Debug logging for outgoing messages
      console.log(
        `%c[WebSocket Send] Action: ${action}`,
        'color: #ff69b4; font-weight: bold'
      );
      console.log('Payload:', payload);
      console.log('Full message:', message);

      this.socket.send(JSON.stringify(message));
    } else {
      console.warn('Cannot send message: WebSocket not connected');
    }
  }

  // 監聽特定 action
  on(action: string, callback: (payload: any) => void): void {
    this.listeners.set(action, callback);
  }

  // 取消監聽
  off(action: string): void {
    this.listeners.delete(action);
  }

  // Subscribe to game state updates
  onGameStateUpdate(callback: (state: ClientGameState) => void): () => void {
    this.gameStateListeners.add(callback);
    // Return unsubscribe function
    return () => {
      this.gameStateListeners.delete(callback);
    };
  }

  // Handle binary protobuf messages
  private async handleBinaryMessage(data: ArrayBuffer | Blob): Promise<void> {
    try {
      // Convert Blob to ArrayBuffer if needed
      const buffer = data instanceof Blob ? await data.arrayBuffer() : data;

      console.log('📦 Received binary message, size:', buffer.byteLength);

      // Check if protobuf is ready
      if (!isProtobufReady()) {
        console.warn('⚠️ Protobuf not ready yet, skipping binary message');
        return;
      }

      // Decode the protobuf message
      const protobufState = decodeGameState(buffer);

      console.log('✅ Protobuf decoded:', {
        sessionId: uuidBytesToString(protobufState.sessionId),
        containerCount: protobufState.containers.length,
        currentPlayerItems: protobufState.currentPlayer?.inventory.length || 0,
        containers: protobufState.containers.map(c => ({
          id: uuidBytesToString(c.entityId),
          itemCount: c.items.length,
          isOpen: c.isOpen,
        })),
      });

      // Convert protobuf format to our existing ClientGameState format
      const gameState = this.convertProtobufToGameState(protobufState);

      console.log('✅ Converted to game state:', {
        containers: gameState.containers.map(c => ({
          id: c.entity_id,
          itemCount: c.items.length,
          isOpen: c.is_open,
        })),
        currentPlayerInventory: gameState.current_player?.inventory?.length || 0,
      });

      // Handle as normal game state
      this.handleGameStateUpdate(gameState);
    } catch (error) {
      console.error('❌ Failed to handle binary message:', error);
      console.error('Error details:', error);
      GameStateLogger.logError('Failed to decode protobuf message', error);
    }
  }

  // Convert Protobuf ClientGameState to our existing ClientGameState format
  private convertProtobufToGameState(pbState: ProtobufGameState): ClientGameState {
    return {
      session_id: uuidBytesToString(pbState.sessionId),
      current_player: pbState.currentPlayer ? {
        id: uuidBytesToString(pbState.currentPlayer.id),
        entity_id: uuidBytesToString(pbState.currentPlayer.entityId),
        username: pbState.currentPlayer.username,
        position: pbState.currentPlayer.position,
        direction: pbState.currentPlayer.direction,
        inventory: pbState.currentPlayer.inventory.map(item => ({
          item_id: uuidBytesToString(item.itemId),
          entity_id: uuidBytesToString(item.entityId),
          name: item.name,
          quantity: item.quantity,
          attack_power: item.attackPower,
          durability: item.durability,
          critical_rate: item.criticalRate,
          weapon_type: item.weaponType,
          defense_rating: item.defenseRating,
          armor_slot: item.armorSlot,
          healing_amount: item.healingAmount,
          mana_amount: item.manaAmount,
          description: item.description,
        })),
      } : null,
      other_players: pbState.otherPlayers.map(player => ({
        id: uuidBytesToString(player.id),
        entity_id: uuidBytesToString(player.entityId),
        username: player.username,
        position: player.position,
        direction: player.direction,
        inventory: player.inventory.map(item => ({
          item_id: uuidBytesToString(item.itemId),
          entity_id: uuidBytesToString(item.entityId),
          name: item.name,
          quantity: item.quantity,
          attack_power: item.attackPower,
          durability: item.durability,
          critical_rate: item.criticalRate,
          weapon_type: item.weaponType,
          defense_rating: item.defenseRating,
          armor_slot: item.armorSlot,
          healing_amount: item.healingAmount,
          mana_amount: item.manaAmount,
          description: item.description,
        })),
      })),
      items: pbState.items.map(itemId => uuidBytesToString(itemId)),
      doors: pbState.doors.map(door => ({
        entity_id: uuidBytesToString(door.entityId),
        position: door.position,
        is_open: door.isOpen,
      })),
      containers: pbState.containers.map(container => ({
        container_id: uuidBytesToString(container.containerId),
        entity_id: uuidBytesToString(container.entityId),
        position: container.position,
        is_open: container.isOpen,
        items: container.items.map(item => ({
          item_id: uuidBytesToString(item.itemId),
          entity_id: uuidBytesToString(item.entityId),
          name: item.name,
          quantity: item.quantity,
          attack_power: item.attackPower,
          durability: item.durability,
          critical_rate: item.criticalRate,
          weapon_type: item.weaponType,
          defense_rating: item.defenseRating,
          armor_slot: item.armorSlot,
          healing_amount: item.healingAmount,
          mana_amount: item.manaAmount,
          description: item.description,
        })),
      })),
    };
  }

  // Handle game state updates
  private handleGameStateUpdate(state: ClientGameState): void {
    // Log the state with formatted output
    GameStateLogger.logGameState(state);

    // Notify all game state listeners
    this.gameStateListeners.forEach((listener) => {
      try {
        listener(state);
      } catch (error) {
        console.error("Error in game state listener:", error);
      }
    });
  }
}

export const socketManager = new SocketManager();
