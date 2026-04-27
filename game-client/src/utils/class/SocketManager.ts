import { ActionMap, ClientMessage } from "@/assets/types/client";
import { ClientGameState, isGameState } from "@/types/gameState";
import { GameStateLogger } from "@/utils/gameStateLogger";
import { useGameStore } from "@/stores/gameStore";
import { useAuthStore } from "@/stores/authStore";

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
  // Generic server-side message error callback (top-level `error` field)
  private onMessageError?: (info: {
    action: string;
    error: string;
    message?: string;
  }) => void;
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

  setOnMessageError(
    callback: (info: { action: string; error: string; message?: string }) => void,
  ) {
    this.onMessageError = callback;
  }

  disconnect() {
    if (this.socket) {
      this.socket.close(1000, "client_disconnect");
      this.socket = null;
    }
    this.setConnectionStatus("disconnected");
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
    };

    this.socket.onclose = (event) => {
      console.log("WebSocket disconnected, code:", event.code, "reason:", event.reason);
      this.socket = null;

      // Close code 4001 = auth error (token invalid/expired)
      if (event.code === 4001) {
        console.log("Auth error detected, redirecting to login...");
        if (this.onAuthError) {
          this.onAuthError();
        }
        return;
      }

      this.setConnectionStatus("disconnected");
      this.updateStatus("WebSocket Disconnected", "#ffcc00");
    };

    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        // Top-level server error (e.g. duplicate queue, invalid action). Surface to a global handler.
        if (data.error) {
          console.warn("Server error message received:", data);
          if (this.onMessageError) {
            this.onMessageError({
              action: data.action ?? "",
              error: String(data.error),
              message: data.payload?.message,
            });
          }
          return;
        }

        // Check if it's a game state update
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

      // Debug logging for outgoing messages (level 2+ = info)
      if (GameStateLogger.getLevel() >= 2) {
        console.log(
          `%c[WebSocket Send] Action: ${action}`,
          'color: #ff69b4; font-weight: bold',
          payload,
        );
      }

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
