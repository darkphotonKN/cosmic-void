import { ActionType } from "@/assets/types/client";
import { socketManager, ConnectionStatus } from "@/utils/class/SocketManager";
import Phaser from "phaser";

export class MainMenuScene extends Phaser.Scene {
  private unsubscribeConnectionStatus?: () => void;
  private buttonBg?: Phaser.GameObjects.Rectangle;
  private startButtonText?: Phaser.GameObjects.Text;
  private connectionStatusText?: Phaser.GameObjects.Text;
  private isConnected: boolean = false;
  private dotAnimation?: Phaser.Time.TimerEvent;
  private dotCount: number = 0;

  constructor() {
    super({ key: "MainMenuScene" });
  }

  create(): void {
    const width = this.cameras.main.width;
    const height = this.cameras.main.height;

    // 背景
    this.cameras.main.setBackgroundColor("#1a1a2e");

    // 創建格子背景
    const graphics = this.add.graphics();
    graphics.lineStyle(1, 0x333355, 0.3);
    for (let x = 0; x <= width; x += 50) {
      graphics.lineBetween(x, 0, x, height);
    }
    for (let y = 0; y <= height; y += 50) {
      graphics.lineBetween(0, y, width, y);
    }

    // Title
    const title = this.add.text(width / 2, height / 4, "🗺️ COSMIC VOID", {
      fontSize: "42px",
      color: "#4ecca3",
      fontStyle: "bold",
    });
    title.setOrigin(0.5);

    // Subtitle
    const subtitle = this.add.text(
      width / 2,
      height / 4 + 50,
      "Multiplayer Treasure Hunt",
      {
        fontSize: "20px",
        color: "#888888",
      },
    );
    subtitle.setOrigin(0.5);

    // Description
    const desc = this.add.text(
      width / 2,
      height / 2 - 30,
      "Fog of War System\nBuilding Collision + Indoor Visibility",
      {
        fontSize: "16px",
        color: "#aaaaaa",
        align: "center",
      },
    );
    desc.setOrigin(0.5);

    // Start button - 初始為禁用狀態
    this.buttonBg = this.add.rectangle(
      width / 2,
      height / 2 + 60,
      200,
      50,
      0x666666, // 灰色表示禁用
    );

    this.startButtonText = this.add.text(
      width / 2,
      height / 2 + 60,
      "Connecting...",
      {
        fontSize: "24px",
        color: "#1a1a2e",
        fontStyle: "bold",
      },
    );
    this.startButtonText.setOrigin(0.5);

    // 連線狀態文字
    this.connectionStatusText = this.add.text(
      width / 2,
      height / 2 + 100,
      "Connecting to server...",
      {
        fontSize: "14px",
        color: "#ffcc00",
      },
    );
    this.connectionStatusText.setOrigin(0.5);

    // 點點動畫
    this.dotAnimation = this.time.addEvent({
      delay: 500,
      callback: () => {
        if (!this.isConnected && this.connectionStatusText) {
          this.dotCount = (this.dotCount + 1) % 4;
          const dots = ".".repeat(this.dotCount);
          this.connectionStatusText.setText(`Connecting to server${dots}`);
        }
      },
      loop: true,
    });

    // 監聽連線狀態
    this.unsubscribeConnectionStatus = socketManager.onConnectionStatusChange(
      (status: ConnectionStatus) => {
        this.handleConnectionStatusChange(status);
      },
    );

    // Controls info
    const controlsText = this.add.text(
      width / 2,
      height - 80,
      "🎮 WASD/Arrows to Move  |  ⚔️ Space to Attack  |  📦 E to Collect",
      {
        fontSize: "14px",
        color: "#666666",
      },
    );
    controlsText.setOrigin(0.5);
  }

  private handleConnectionStatusChange(status: ConnectionStatus): void {
    if (!this.buttonBg || !this.startButtonText || !this.connectionStatusText) {
      return;
    }

    switch (status) {
      case "connected":
        this.isConnected = true;
        // 停止點點動畫
        if (this.dotAnimation) {
          this.dotAnimation.destroy();
        }
        // 啟用按鈕
        this.buttonBg.setFillStyle(0x4ecca3);
        this.buttonBg.setInteractive({ useHandCursor: true });
        this.startButtonText.setText("Start Game");
        this.connectionStatusText.setText("Connected!");
        this.connectionStatusText.setColor("#4ecca3");
        // 設置按鈕交互
        this.setupButtonInteraction();
        break;

      case "connecting":
        this.isConnected = false;
        this.buttonBg.setFillStyle(0x666666);
        this.buttonBg.disableInteractive();
        this.startButtonText.setText("Connecting...");
        this.connectionStatusText.setColor("#ffcc00");
        break;

      case "error":
        this.isConnected = false;
        if (this.dotAnimation) {
          this.dotAnimation.destroy();
        }
        this.buttonBg.setFillStyle(0xff4444);
        this.buttonBg.disableInteractive();
        this.startButtonText.setText("Error");
        this.connectionStatusText.setText("Connection failed. Please refresh.");
        this.connectionStatusText.setColor("#ff4444");
        break;

      case "disconnected":
        this.isConnected = false;
        if (this.dotAnimation) {
          this.dotAnimation.destroy();
        }
        this.buttonBg.setFillStyle(0x666666);
        this.buttonBg.disableInteractive();
        this.startButtonText.setText("Disconnected");
        this.connectionStatusText.setText("Connection lost. Please refresh.");
        this.connectionStatusText.setColor("#ffcc00");
        break;
    }
  }

  private setupButtonInteraction(): void {
    if (!this.buttonBg) return;

    this.buttonBg.on("pointerover", () => {
      if (this.isConnected) {
        this.buttonBg?.setFillStyle(0x3dbb92);
      }
    });

    this.buttonBg.on("pointerout", () => {
      if (this.isConnected) {
        this.buttonBg?.setFillStyle(0x4ecca3);
      }
    });

    this.buttonBg.on("pointerdown", () => {
      if (this.isConnected) {
        socketManager.sendMessage(ActionType.Find_Game, { playerId: "1" });
        this.queuePopup();
      }
    });
  }

  shutdown(): void {
    // 場景銷毀時取消訂閱
    if (this.unsubscribeConnectionStatus) {
      this.unsubscribeConnectionStatus();
    }
    if (this.dotAnimation) {
      this.dotAnimation.destroy();
    }
  }

  queuePopup() {
    const { width, height } = this.scale;

    // 半透明背景遮罩
    const overlay = this.add.rectangle(
      width / 2,
      height / 2,
      width,
      height,
      0x000000,
      0.7,
    );

    // 彈窗背景
    const popup = this.add.container(width / 2, height / 2);

    const bg = this.add
      .rectangle(0, 0, 300, 200, 0xffffff, 1)
      .setStrokeStyle(2, 0x000000);

    const title = this.add
      .text(0, -60, "Queueing...", {
        fontSize: "28px",
        color: "#000",
      })
      .setOrigin(0.5);

    const closeBtn = this.add
      .text(0, 50, "Close", {
        fontSize: "20px",
        backgroundColor: "#4ecca3",
        padding: { x: 20, y: 10 },
      })
      .setOrigin(0.5)
      .setInteractive({ useHandCursor: true });

    const peopleCountText = this.add
      .text(0, -10, "People in queue: 0 / 2", {
        fontSize: "16px",
        color: "#000",
      })
      .setOrigin(0.5);

    // 監聽排隊狀態更新
    socketManager.on(
      "queue_status",
      (payload: { current: number; total: number }) => {
        console.log("payload", payload);
        if (!payload) return;
        peopleCountText.setText(
          `People in queue: ${payload.current} / ${payload.total}`,
        );
      },
    );

    // 監聽配對成功
    socketManager.on("game_found", (payload: { sessionID: string }) => {
      if (!payload) return;
      console.log("Game found! Session ID:", payload.sessionID);
      title.setText("Game Found!");
      peopleCountText.setText("Starting game...");

      // 1.5 秒後進入遊戲場景
      this.time.delayedCall(1500, () => {
        socketManager.off("queue_status");
        socketManager.off("game_found");
        overlay.destroy();
        popup.destroy();
        this.scene.start("TreasureHuntScene", { sessionID: payload.sessionID });
      });
    });

    closeBtn.on("pointerdown", () => {
      // 取消監聽
      socketManager.off("queue_status");
      socketManager.off("game_found");
      // TODO: 發送離開排隊的訊息給後端
      overlay.destroy();
      popup.destroy();
    });

    popup.add([bg, title, closeBtn, peopleCountText]);
  }
}
