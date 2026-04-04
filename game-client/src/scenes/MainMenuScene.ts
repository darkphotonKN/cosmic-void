import { ActionType } from "@/assets/types/client";
import { socketManager, ConnectionStatus } from "@/utils/class/SocketManager";
import Phaser from "phaser";

export class MainMenuScene extends Phaser.Scene {
  private unsubscribeConnectionStatus?: () => void;
  private buttonBg?: Phaser.GameObjects.Graphics;
  private buttonGlow?: Phaser.GameObjects.Graphics;
  private startButtonText?: Phaser.GameObjects.Text;
  private connectionStatusText?: Phaser.GameObjects.Text;
  private isConnected: boolean = false;
  private dotAnimation?: Phaser.Time.TimerEvent;
  private dotCount: number = 0;
  private scanlineGraphics?: Phaser.GameObjects.Graphics;
  private glowTween?: Phaser.Tweens.Tween;
  private loadoutBtnBg?: Phaser.GameObjects.Graphics;
  private queuePopupActive: boolean = false;
  private queueTitle?: Phaser.GameObjects.Text;
  private queuePeopleText?: Phaser.GameObjects.Text;
  private queueOverlay?: Phaser.GameObjects.Rectangle;
  private queuePopupContainer?: Phaser.GameObjects.Container;

  constructor() {
    super({ key: "MainMenuScene" });
  }

  create(): void {
    const width = this.cameras.main.width;
    const height = this.cameras.main.height;

    // Menu music
    if (!this.sound.get('menuTheme')) {
      this.sound.play('menuTheme', { loop: true, volume: 0.35 });
    } else if (!this.sound.get('menuTheme')?.isPlaying) {
      this.sound.play('menuTheme', { loop: true, volume: 0.35 });
    }
    // Stop game ambient if returning from game
    this.sound.stopByKey('gameAmbient');

    // Deep space background
    this.cameras.main.setBackgroundColor("#0a0a12");

    // Star field
    const stars = this.add.graphics();
    for (let i = 0; i < 120; i++) {
      const x = Phaser.Math.Between(0, width);
      const y = Phaser.Math.Between(0, height);
      const size = Math.random() < 0.1 ? 2 : 1;
      const alpha = Phaser.Math.FloatBetween(0.15, 0.6);
      const color = Math.random() < 0.3 ? 0x00f0ff : Math.random() < 0.5 ? 0xff00aa : 0xffffff;
      stars.fillStyle(color, alpha);
      stars.fillRect(x, y, size, size);
    }

    // Subtle grid overlay
    const grid = this.add.graphics();
    grid.lineStyle(1, 0x00f0ff, 0.04);
    for (let x = 0; x <= width; x += 40) {
      grid.lineBetween(x, 0, x, height);
    }
    for (let y = 0; y <= height; y += 40) {
      grid.lineBetween(0, y, width, y);
    }

    // Scanline effect
    this.scanlineGraphics = this.add.graphics();
    this.scanlineGraphics.fillStyle(0x000000, 0.03);
    for (let y = 0; y < height; y += 4) {
      this.scanlineGraphics.fillRect(0, y, width, 2);
    }
    this.scanlineGraphics.setDepth(100);

    // Horizontal accent line under title area
    const accentLine = this.add.graphics();
    accentLine.lineStyle(1, 0x00f0ff, 0.3);
    accentLine.lineBetween(width * 0.2, height / 4 + 85, width * 0.8, height / 4 + 85);

    // Title
    const title = this.add.text(width / 2, height / 4, "COSMIC VOID", {
      fontSize: "48px",
      color: "#00f0ff",
      fontStyle: "bold",
      letterSpacing: 12,
    });
    title.setOrigin(0.5);

    // Title glow effect via shadow
    title.setShadow(0, 0, "#00f0ff", 8, true, true);

    // Subtitle
    const subtitle = this.add.text(
      width / 2,
      height / 4 + 55,
      "EXTRACTION // SURVIVE THE VOID",
      {
        fontSize: "14px",
        color: "#ff00aa",
        letterSpacing: 6,
      },
    );
    subtitle.setOrigin(0.5);

    // Description
    const desc = this.add.text(
      width / 2,
      height / 2 - 40,
      "Loot. Fight. Extract.\nNothing survives the void forever.",
      {
        fontSize: "15px",
        color: "#556677",
        align: "center",
        lineSpacing: 6,
      },
    );
    desc.setOrigin(0.5);

    // Button glow layer (behind button)
    this.buttonGlow = this.add.graphics();
    this.buttonGlow.setAlpha(0);

    // Button background (rounded rect via graphics)
    const btnX = width / 2;
    const btnY = height / 2 + 50;
    const btnW = 220;
    const btnH = 50;

    this.buttonBg = this.add.graphics();
    this.drawButton(0x333344, 0x556677);

    // Invisible hit area for interaction
    const hitArea = this.add.rectangle(btnX, btnY, btnW, btnH, 0x000000, 0);
    hitArea.setInteractive({ useHandCursor: true });

    // Store ref immediately — setupButtonInteraction needs it when
    // onConnectionStatusChange fires synchronously with "connected"
    (this as Record<string, unknown>)._hitArea = hitArea;

    this.startButtonText = this.add.text(
      btnX,
      btnY,
      "CONNECTING...",
      {
        fontSize: "18px",
        color: "#0a0a12",
        fontStyle: "bold",
        letterSpacing: 3,
      },
    );
    this.startButtonText.setOrigin(0.5);

    // Connection status text — near bottom of screen
    this.connectionStatusText = this.add.text(
      width / 2,
      height - 40,
      "Establishing uplink...",
      {
        fontSize: "12px",
        color: "#ff8800",
        letterSpacing: 2,
      },
    );
    this.connectionStatusText.setOrigin(0.5);

    // Connecting dot animation
    this.dotAnimation = this.time.addEvent({
      delay: 500,
      callback: () => {
        if (!this.isConnected && this.connectionStatusText) {
          this.dotCount = (this.dotCount + 1) % 4;
          const dots = ".".repeat(this.dotCount);
          this.connectionStatusText.setText(`Establishing uplink${dots}`);
        }
      },
      loop: true,
    });

    // Connection status listener
    this.unsubscribeConnectionStatus = socketManager.onConnectionStatusChange(
      (status: ConnectionStatus) => {
        this.handleConnectionStatusChange(status);
      },
    );

    // Reconnection listener
    socketManager.on("reconnected", (payload: { session_id: string; username: string; message: string }) => {
      console.log("Reconnected!", payload);
      if (this.connectionStatusText) {
        this.connectionStatusText.setText(`Uplink restored // ${payload.username}`);
        this.connectionStatusText.setColor("#00f0ff");
      }
    });

    // Game found listener
    socketManager.on("game_found", (payload: { session_id?: string; sessionID?: string }) => {
      console.log("Game found! Payload:", payload);

      const sessionID = payload.session_id || payload.sessionID;

      if (!sessionID) {
        console.error("No session ID in game_found payload:", payload);
        return;
      }

      if (this.queuePopupActive && this.queueTitle && this.queuePeopleText) {
        this.queueTitle.setText("MATCH FOUND");
        this.queuePeopleText.setText("Deploying...");

        this.time.delayedCall(1500, () => {
          this.closeQueuePopup();
          this.scene.start("CosmicVoidScene", { sessionID });
        });
      } else {
        this.scene.start("CosmicVoidScene", { sessionID });
      }
    });

    // Manage Loadout button
    const loadoutBtnX = width / 2;
    const loadoutBtnY = height / 2 + 115;
    const loadoutBtnW = 180;
    const loadoutBtnH = 36;

    this.loadoutBtnBg = this.add.graphics();
    this.loadoutBtnBg.fillStyle(0x112233, 0.8);
    this.loadoutBtnBg.fillRoundedRect(loadoutBtnX - loadoutBtnW / 2, loadoutBtnY - loadoutBtnH / 2, loadoutBtnW, loadoutBtnH, 4);
    this.loadoutBtnBg.lineStyle(1, 0x00f0ff, 0.3);
    this.loadoutBtnBg.strokeRoundedRect(loadoutBtnX - loadoutBtnW / 2, loadoutBtnY - loadoutBtnH / 2, loadoutBtnW, loadoutBtnH, 4);

    const loadoutText = this.add.text(loadoutBtnX, loadoutBtnY, 'MANAGE LOADOUT', {
      fontSize: '12px',
      color: '#00f0ff',
      letterSpacing: 3,
    });
    loadoutText.setOrigin(0.5);

    const loadoutHit = this.add.rectangle(loadoutBtnX, loadoutBtnY, loadoutBtnW, loadoutBtnH, 0x000000, 0);
    loadoutHit.setInteractive({ useHandCursor: true });

    loadoutHit.on('pointerover', () => {
      if (!this.loadoutBtnBg) return;
      this.loadoutBtnBg.clear();
      this.loadoutBtnBg.fillStyle(0x1a2a3a, 0.9);
      this.loadoutBtnBg.fillRoundedRect(loadoutBtnX - loadoutBtnW / 2, loadoutBtnY - loadoutBtnH / 2, loadoutBtnW, loadoutBtnH, 4);
      this.loadoutBtnBg.lineStyle(1, 0x00f0ff, 0.6);
      this.loadoutBtnBg.strokeRoundedRect(loadoutBtnX - loadoutBtnW / 2, loadoutBtnY - loadoutBtnH / 2, loadoutBtnW, loadoutBtnH, 4);
    });

    loadoutHit.on('pointerout', () => {
      if (!this.loadoutBtnBg) return;
      this.loadoutBtnBg.clear();
      this.loadoutBtnBg.fillStyle(0x112233, 0.8);
      this.loadoutBtnBg.fillRoundedRect(loadoutBtnX - loadoutBtnW / 2, loadoutBtnY - loadoutBtnH / 2, loadoutBtnW, loadoutBtnH, 4);
      this.loadoutBtnBg.lineStyle(1, 0x00f0ff, 0.3);
      this.loadoutBtnBg.strokeRoundedRect(loadoutBtnX - loadoutBtnW / 2, loadoutBtnY - loadoutBtnH / 2, loadoutBtnW, loadoutBtnH, 4);
    });

    loadoutHit.on('pointerdown', () => {
      this.scene.start('LoadoutScene');
    });

    // Controls info
    const controlsText = this.add.text(
      width / 2,
      height - 60,
      "WASD Move  //  SPACE Attack  //  E Interact  //  F Loot  //  I Inventory",
      {
        fontSize: "11px",
        color: "#334455",
        letterSpacing: 2,
      },
    );
    controlsText.setOrigin(0.5);

    // Version / flavor text
    const versionText = this.add.text(
      width / 2,
      height - 35,
      "v0.1 // SECTOR 7-G // UNAUTHORIZED ACCESS WILL BE TERMINATED",
      {
        fontSize: "9px",
        color: "#222233",
        letterSpacing: 1,
      },
    );
    versionText.setOrigin(0.5);
  }

  private drawButton(fill: number, stroke: number, glowColor?: number): void {
    const width = this.cameras.main.width;
    const btnX = width / 2 - 110;
    const btnY = this.cameras.main.height / 2 + 50 - 25;
    const btnW = 220;
    const btnH = 50;

    if (this.buttonGlow && glowColor) {
      this.buttonGlow.clear();
      this.buttonGlow.fillStyle(glowColor, 0.15);
      this.buttonGlow.fillRoundedRect(btnX - 4, btnY - 4, btnW + 8, btnH + 8, 10);
      this.buttonGlow.setAlpha(1);
    }

    if (this.buttonBg) {
      this.buttonBg.clear();
      this.buttonBg.fillStyle(fill, 1);
      this.buttonBg.fillRoundedRect(btnX, btnY, btnW, btnH, 6);
      this.buttonBg.lineStyle(1, stroke, 0.8);
      this.buttonBg.strokeRoundedRect(btnX, btnY, btnW, btnH, 6);
    }
  }

  private handleConnectionStatusChange(status: ConnectionStatus): void {
    if (!this.buttonBg || !this.startButtonText || !this.connectionStatusText) {
      return;
    }

    const hitArea = (this as Record<string, unknown>)._hitArea as Phaser.GameObjects.Rectangle | undefined;

    switch (status) {
      case "connected":
        this.isConnected = true;
        if (this.dotAnimation) {
          this.dotAnimation.destroy();
        }
        this.drawButton(0x00f0ff, 0x00f0ff, 0x00f0ff);
        this.startButtonText.setText("FIND MATCH");
        this.startButtonText.setColor("#0a0a12");
        this.connectionStatusText.setText("Uplink active");
        this.connectionStatusText.setColor("#00f0ff");
        this.setupButtonInteraction();
        break;

      case "connecting":
        this.isConnected = false;
        this.drawButton(0x333344, 0x556677);
        if (hitArea) hitArea.disableInteractive();
        this.startButtonText.setText("CONNECTING...");
        this.startButtonText.setColor("#556677");
        this.connectionStatusText.setColor("#ff8800");
        break;

      case "error":
        this.isConnected = false;
        if (this.dotAnimation) {
          this.dotAnimation.destroy();
        }
        this.drawButton(0x441111, 0xff2244, 0xff2244);
        if (hitArea) hitArea.disableInteractive();
        this.startButtonText.setText("OFFLINE");
        this.startButtonText.setColor("#ff2244");
        this.connectionStatusText.setText("Uplink failed // Refresh to retry");
        this.connectionStatusText.setColor("#ff2244");
        break;

      case "disconnected":
        this.isConnected = false;
        if (this.dotAnimation) {
          this.dotAnimation.destroy();
        }
        this.drawButton(0x222233, 0x556677);
        if (hitArea) hitArea.disableInteractive();
        this.startButtonText.setText("DISCONNECTED");
        this.startButtonText.setColor("#556677");
        this.connectionStatusText.setText("Uplink lost // Refresh to retry");
        this.connectionStatusText.setColor("#ff8800");
        break;
    }
  }

  private setupButtonInteraction(): void {
    const hitArea = (this as Record<string, unknown>)._hitArea as Phaser.GameObjects.Rectangle | undefined;
    if (!hitArea) return;

    hitArea.setInteractive({ useHandCursor: true });

    hitArea.on("pointerover", () => {
      if (this.isConnected) {
        this.drawButton(0x33ffdd, 0x00f0ff, 0x00f0ff);
      }
    });

    hitArea.on("pointerout", () => {
      if (this.isConnected) {
        this.drawButton(0x00f0ff, 0x00f0ff, 0x00f0ff);
      }
    });

    hitArea.on("pointerdown", () => {
      if (this.isConnected) {
        socketManager.sendMessage(ActionType.Find_Game, { playerId: "1" });
        this.queuePopup();
      }
    });
  }

  shutdown(): void {
    if (this.unsubscribeConnectionStatus) {
      this.unsubscribeConnectionStatus();
    }
    if (this.dotAnimation) {
      this.dotAnimation.destroy();
    }
    if (this.glowTween) {
      this.glowTween.destroy();
    }
  }

  queuePopup() {
    const { width, height } = this.scale;

    this.queuePopupActive = true;

    // Dark overlay
    this.queueOverlay = this.add.rectangle(
      width / 2,
      height / 2,
      width,
      height,
      0x000000,
      0.8,
    );

    this.queuePopupContainer = this.add.container(width / 2, height / 2);

    // Popup background
    const popupW = 320;
    const popupH = 200;
    const bg = this.add.graphics();
    bg.fillStyle(0x0a0a18, 1);
    bg.fillRoundedRect(-popupW / 2, -popupH / 2, popupW, popupH, 8);
    bg.lineStyle(1, 0x00f0ff, 0.4);
    bg.strokeRoundedRect(-popupW / 2, -popupH / 2, popupW, popupH, 8);

    this.queueTitle = this.add
      .text(0, -60, "SEARCHING FOR MATCH", {
        fontSize: "18px",
        color: "#00f0ff",
        letterSpacing: 4,
      })
      .setOrigin(0.5);

    this.queuePeopleText = this.add
      .text(0, -15, "Operators in queue: 0 / 2", {
        fontSize: "14px",
        color: "#556677",
      })
      .setOrigin(0.5);

    // Cancel button
    const cancelBg = this.add.graphics();
    cancelBg.fillStyle(0x221122, 1);
    cancelBg.fillRoundedRect(-60, 35, 120, 36, 4);
    cancelBg.lineStyle(1, 0xff00aa, 0.5);
    cancelBg.strokeRoundedRect(-60, 35, 120, 36, 4);

    const cancelText = this.add
      .text(0, 53, "CANCEL", {
        fontSize: "13px",
        color: "#ff00aa",
        letterSpacing: 3,
      })
      .setOrigin(0.5);

    const cancelHit = this.add.rectangle(0, 53, 120, 36, 0x000000, 0);
    cancelHit.setInteractive({ useHandCursor: true });

    // Queue status listener
    socketManager.on(
      "queue_status",
      (payload: { current: number; total: number }) => {
        console.log("Queue status payload:", payload);
        if (!payload || !this.queuePeopleText) return;
        this.queuePeopleText.setText(
          `Operators in queue: ${payload.current} / ${payload.total}`,
        );
      },
    );

    cancelHit.on("pointerdown", () => {
      this.closeQueuePopup();
      // TODO: send leave queue message to backend
    });

    this.queuePopupContainer.add([bg, this.queueTitle, this.queuePeopleText, cancelBg, cancelText, cancelHit]);
  }

  private closeQueuePopup() {
    this.queuePopupActive = false;

    socketManager.off("queue_status");

    if (this.queueOverlay) {
      this.queueOverlay.destroy();
      this.queueOverlay = undefined;
    }
    if (this.queuePopupContainer) {
      this.queuePopupContainer.destroy();
      this.queuePopupContainer = undefined;
    }

    this.queueTitle = undefined;
    this.queuePeopleText = undefined;
  }
}
