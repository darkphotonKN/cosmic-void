import { ActionType } from "@/assets/types/client";
import { socketManager } from "@/utils/class/SocketManager";
import Phaser from "phaser";

export class MainMenuScene extends Phaser.Scene {
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

    // Start button
    const buttonBg = this.add.rectangle(
      width / 2,
      height / 2 + 60,
      200,
      50,
      0x4ecca3,
    );
    buttonBg.setInteractive({ useHandCursor: true });

    const startButtonText = this.add.text(
      width / 2,
      height / 2 + 60,
      "Start Game",
      {
        fontSize: "24px",
        color: "#1a1a2e",
        fontStyle: "bold",
      },
    );
    startButtonText.setOrigin(0.5);

    buttonBg.on("pointerover", () => {
      buttonBg.setFillStyle(0x3dbb92);
    });

    buttonBg.on("pointerout", () => {
      buttonBg.setFillStyle(0x4ecca3);
    });

    buttonBg.on("pointerdown", () => {
      // this.scene.start("TreasureHuntScene");
      socketManager.sendMessage(ActionType.Find_Game, { playerId: "1" });
      this.queuePopup();
    });

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

    closeBtn.on("pointerdown", () => {
      overlay.destroy();
      popup.destroy();
    });

    popup.add([bg, title, closeBtn, peopleCountText]);
  }
}
