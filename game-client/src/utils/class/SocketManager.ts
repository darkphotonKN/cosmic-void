import { ActionMap, ClientMessage } from "@/assets/types/client";

// SocketManager.js
class SocketManager {
  private socket: WebSocket | null = null;
  private listeners: Map<string, (data: any) => void> = new Map();
  // Status callback
  private onStatusChange?: (status: string, color: string) => void;
  private seq: number = 0;

  constructor() {
    this.socket = null;
    this.listeners = new Map();
  }

  connect(url: string) {
    if (this.socket) return; // 避免重複連接

    this.socket = new WebSocket(url);

    this.socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      // 通知所有監聽者
      if (this.listeners.has(data.type)) {
        this.listeners.get(data.type)?.(data);
      }
    };

    this.socket.onopen = () => {
      console.log("WebSocket connected");
      this.updateStatus("WebSocket Connected", "#4ecca3");
    };

    this.socket.onerror = (error) => {
      console.error("WebSocket error:", error);
      this.updateStatus("WebSocket Error", "#ff4444");
    };

    this.socket.onclose = () => {
      console.log("WebSocket disconnected");
      this.updateStatus("WebSocket Disconnected", "#ffcc00");
    };

    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received server message:", data);
      } catch (e) {
        console.error("Failed to parse message:", e);
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
      const message: ClientMessage<T> = {
        action,
        payload,
        seq: ++this.seq,
      };
      this.socket.send(JSON.stringify(message));
    }
  }

  //   on(type: string, callback) {
  //     this.listeners.set(type, callback);
  //   }

  //   off(type) {
  //     this.listeners.delete(type);
  //   }
}

export const socketManager = new SocketManager();
