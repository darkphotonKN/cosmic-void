/**
 * CosmicVoidScene - 簡化版遊戲場景
 * 移動邏輯 + WebSocket + 建築（進入後看不到外面）
 */

import Phaser from "phaser";
import { ActionType } from "@/assets/types/client";
import { socketManager } from "@/utils/class/SocketManager";
import {
  ClientGameState,
  ContainerState,
  ItemState,
  EscapeDoorState,
  SwitchState,
  WallState,
  DoorState,
  EquippedItems,
  EquipmentState,
} from "@/types/gameState";
import { EquipmentPanel } from "@/ui/EquipmentPanel";
import { GameStateLogger } from "@/utils/gameStateLogger";

interface Building {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
  doorSide: "top" | "bottom" | "left" | "right";
  wallGroup: Phaser.Physics.Arcade.StaticGroup;
  roof: Phaser.GameObjects.Graphics;
  floor: Phaser.GameObjects.Graphics;
  doorMarker: Phaser.GameObjects.Graphics;
  // Door properties
  door: Phaser.GameObjects.Graphics;
  doorCollider: Phaser.GameObjects.Rectangle;
  isOpen: boolean;
}

export class CosmicVoidScene extends Phaser.Scene {
  private player?: Phaser.Physics.Arcade.Sprite;
  private otherPlayers: Map<string, Phaser.Physics.Arcade.Sprite> = new Map();
  private otherPlayersEntityIds: Map<string, string> = new Map(); // player_id → entity_id
  private otherPlayersTargets: Map<string, { x: number; y: number }> =
    new Map();

  // leg graphics for walking animation
  private playerLegs?: Phaser.GameObjects.Graphics;
  private otherPlayersLegs: Map<string, Phaser.GameObjects.Graphics> =
    new Map();
  private playerFacing: "up" | "down" | "left" | "right" = "down";
  private walkPhase = 0;
  private otherPlayersFacing: Map<string, "up" | "down" | "left" | "right"> =
    new Map();
  private otherPlayersWalkPhase: Map<string, number> = new Map();

  // username labels
  private playerNameText?: Phaser.GameObjects.Text;
  private otherPlayersNameTexts: Map<string, Phaser.GameObjects.Text> =
    new Map();
  private hoveredPlayerId?: string; // survives game state rerenders

  // Controls
  private cursors!: Phaser.Types.Input.Keyboard.CursorKeys;
  private wasd!: {
    up: Phaser.Input.Keyboard.Key;
    down: Phaser.Input.Keyboard.Key;
    left: Phaser.Input.Keyboard.Key;
    right: Phaser.Input.Keyboard.Key;
  };

  // In-game controls panel
  private controlsPanel?: Phaser.GameObjects.Container;

  // End-of-game overlay (shown when server sends end_game action)
  private gameEndOverlay?: Phaser.GameObjects.Container;

  // Game state
  private gameStateUnsubscribe?: () => void;
  private targetPosition: { x: number; y: number } | null = null;

  // 地圖大小
  private mapWidth = 1440;
  private mapHeight = 960;

  // 建築
  private buildings: Building[] = [];
  private currentBuilding: Building | null = null;
  private outsideObjects: Phaser.GameObjects.GameObject[] = [];
  private indoorMask!: Phaser.GameObjects.Graphics;

  // 寶箱 (從後端同步)
  private chests: Map<
    string,
    { sprite: Phaser.GameObjects.Sprite; entityId: string }
  > = new Map();

  // 逃脫門 (從後端同步)
  private escapeDoors: Map<
    string,
    { sprite: Phaser.GameObjects.Sprite; entityId: string }
  > = new Map();

  // 開關/按鈕 (從後端同步)
  private switches: Map<
    string,
    { sprite: Phaser.GameObjects.Sprite; entityId: string }
  > = new Map();

  // 牆壁 (從後端同步)
  private walls: Map<
    string,
    { graphics: Phaser.GameObjects.Graphics; entityId: string }
  > = new Map();

  // 門 (從後端同步)
  private serverDoors: Map<
    string,
    { rect: Phaser.GameObjects.Rectangle; entityId: string; isOpen: boolean }
  > = new Map();
  private serverBuildingsCreated = false;

  // 寶箱跳窗
  private chestPopup?: Phaser.GameObjects.Container;
  private isPopupOpen = false;
  private openedChestEntityId?: string;
  private popupItemsText?: Phaser.GameObjects.Text;

  // 道具欄 + 裝備面板
  private equipmentPanel?: EquipmentPanel;
  private equippedItems: EquippedItems = {
    weapon: null,
    head: null,
    body: null,
    hands: null,
    feet: null,
    ring_1: null,
    ring_2: null,
    consumable_1: null,
    consumable_2: null,
    consumable_3: null,
  };
  private inventoryItems: ItemState[] = [];

  // Item row grid system (manual hit testing — Phaser input is broken with scrollFactor 0)
  private itemRows: {
    screenRect: { x: number; y: number; w: number; h: number };
    item: ItemState;
    label: Phaser.GameObjects.Text;
    rowBg: Phaser.GameObjects.Graphics;
    source: "chest";
  }[] = [];
  private hoveredRowIndex = -1;
  private hoveredItemEntityId?: string; // Survives row rebuilds
  private lastPointerX = 0;
  private lastPointerY = 0;
  private itemTooltip?: Phaser.GameObjects.Container;
  private chestItemFingerprint = "";

  // 當前寶箱的物品（用於 F 鍵取得）
  private currentChestItems: ItemState[] = [];
  private chestLootedAtMap = new Map<string, number>(); // entityId → loot 時間戳
  private canAttack = true;
  private readonly PENDING_DURATION = 1000; // 1 秒內不比對剛拿的物品
  private lastGameState?: ClientGameState;

  // 狀態追蹤：避免重複通知（每秒 33 幀會重複收到相同狀態）
  private previousEscapeDoorLocked: boolean | null = null;
  private previousEscapeDoorOpened: boolean | null = null;
  private previousSwitchActivated: boolean | null = null;
  private escapedPlayers: Set<string> = new Set();
  private escapedCountText?: Phaser.GameObjects.Text;

  constructor() {
    super({ key: "CosmicVoidScene" });
  }

  private toggleControlsPanel(): void {
    if (this.controlsPanel) {
      this.controlsPanel.destroy();
      this.controlsPanel = undefined;
      return;
    }

    const cam = this.cameras.main;
    const panelW = 260;
    const panelH = 220;
    const x = cam.width / 2;
    const y = cam.height / 2;

    const children: Phaser.GameObjects.GameObject[] = [];

    const bg = this.add.graphics();
    bg.fillStyle(0x0a0a12, 0.92);
    bg.fillRoundedRect(-panelW / 2, -panelH / 2, panelW, panelH, 8);
    bg.lineStyle(1, 0x00f0ff, 0.5);
    bg.strokeRoundedRect(-panelW / 2, -panelH / 2, panelW, panelH, 8);
    children.push(bg);

    const title = this.add.text(0, -panelH / 2 + 16, "CONTROLS", {
      fontSize: "16px",
      color: "#00f0ff",
      letterSpacing: 5,
    });
    title.setOrigin(0.5);
    children.push(title);

    const controls = [
      ["WASD", "Move"],
      ["E", "Interact"],
      ["F", "Take Item"],
      ["I", "Equipment"],
      ["Q", "Close Panel"],
      ["CLICK", "Attack Enemy"],
      ["ESC", "Main Menu"],
      ["H", "Toggle Controls"],
    ];

    let curY = -panelH / 2 + 44;
    for (const [key, action] of controls) {
      const keyText = this.add.text(-panelW / 2 + 20, curY, key, {
        fontSize: "11px",
        color: "#00f0ff",
        letterSpacing: 2,
      });
      const actionText = this.add.text(panelW / 2 - 20, curY, action, {
        fontSize: "11px",
        color: "#556677",
      });
      actionText.setOrigin(1, 0);
      children.push(keyText, actionText);
      curY += 20;
    }

    const hint = this.add.text(0, panelH / 2 - 16, "H to close", {
      fontSize: "10px",
      color: "#334455",
    });
    hint.setOrigin(0.5);
    children.push(hint);

    this.controlsPanel = this.add.container(x, y, children);
    this.controlsPanel.setDepth(1200);
    this.controlsPanel.setScrollFactor(0);
  }

  private showGameEndOverlay(position: number, result?: string): void {
    if (this.gameEndOverlay) return; // already shown

    const cam = this.cameras.main;

    // full-screen backdrop — eats pointer input from anything beneath
    const backdrop = this.add.rectangle(
      cam.width / 2,
      cam.height / 2,
      cam.width,
      cam.height,
      0x0a0a12,
      0.85,
    );
    // centered card
    const cardW = 460;
    const cardH = 280;
    const card = this.add.graphics();
    card.fillStyle(0x0a0a12, 0.95);
    card.fillRoundedRect(-cardW / 2, -cardH / 2, cardW, cardH, 10);
    card.lineStyle(1, 0x00f0ff, 0.5);
    card.strokeRoundedRect(-cardW / 2, -cardH / 2, cardW, cardH, 10);

    let titleStr: string;
    let subtitleStr: string;

    switch (result) {
      case "escaped":
        titleStr = "VICTORY!";
        subtitleStr = "OPERATOR EXTRACTED";
        break;
      case "survived":
        titleStr = "VICTORY!";
        subtitleStr = "LAST ONE STANDING";
        break;
      default: // "eliminated"
        titleStr = `YOU PLACED #${position}`;
        subtitleStr = "BETTER LUCK ON YOUR NEXT DEPLOY";
        break;
    }

    const title = this.add.text(0, -60, titleStr, {
      fontSize: "44px",
      color: "#00f0ff",
      fontStyle: "bold",
      letterSpacing: 6,
    });
    title.setOrigin(0.5);

    const subtitle = this.add.text(0, 0, subtitleStr, {
      fontSize: "13px",
      color: "#556677",
      letterSpacing: 3,
    });
    subtitle.setOrigin(0.5);

    // --- action buttons ---
    const btnW = 180;
    const btnH = 38;
    const btnY = 70;
    const btnGap = 16;

    // RE-DEPLOY button (re-queue) — cyan fill
    const redeployBtn = this.add.graphics();
    redeployBtn.fillStyle(0x00f0ff, 1);
    redeployBtn.fillRoundedRect(
      -btnW / 2 - btnW / 2 - btnGap / 2,
      btnY - btnH / 2,
      btnW,
      btnH,
      6,
    );
    const redeployText = this.add.text(
      -btnW / 2 - btnGap / 2,
      btnY,
      "RE-DEPLOY",
      {
        fontSize: "14px",
        color: "#0a0a12",
        fontStyle: "bold",
        letterSpacing: 3,
      },
    );
    redeployText.setOrigin(0.5);
    // hit areas live outside containers so pointer events work reliably
    const cx = cam.width / 2;
    const cy = cam.height / 2;
    const redeployHit = this.add
      .rectangle(cx - btnW / 2 - btnGap / 2, cy + btnY, btnW, btnH, 0x000000, 0)
      .setInteractive({ useHandCursor: true })
      .setScrollFactor(0)
      .setDepth(2001);
    redeployHit.on("pointerover", () => {
      redeployBtn.clear();
      redeployBtn.fillStyle(0x33f5ff, 1);
      redeployBtn.fillRoundedRect(
        -btnW / 2 - btnW / 2 - btnGap / 2,
        btnY - btnH / 2,
        btnW,
        btnH,
        6,
      );
    });
    redeployHit.on("pointerout", () => {
      redeployBtn.clear();
      redeployBtn.fillStyle(0x00f0ff, 1);
      redeployBtn.fillRoundedRect(
        -btnW / 2 - btnW / 2 - btnGap / 2,
        btnY - btnH / 2,
        btnW,
        btnH,
        6,
      );
    });
    redeployHit.on("pointerdown", () => {
      this.sound.stopByKey("gameAmbient");
      socketManager.sendMessage(ActionType.Find_Game, { playerId: "1" });
      this.scene.start("MainMenuScene");
    });

    // RETURN TO BASE button — outlined
    const returnBtn = this.add.graphics();
    returnBtn.lineStyle(1, 0x00f0ff, 0.6);
    returnBtn.strokeRoundedRect(btnGap / 2, btnY - btnH / 2, btnW, btnH, 6);
    const returnText = this.add.text(
      btnGap / 2 + btnW / 2,
      btnY,
      "RETURN TO BASE",
      {
        fontSize: "13px",
        color: "#00f0ff",
        letterSpacing: 2,
      },
    );
    returnText.setOrigin(0.5);
    const returnHit = this.add
      .rectangle(cx + btnGap / 2 + btnW / 2, cy + btnY, btnW, btnH, 0x000000, 0)
      .setInteractive({ useHandCursor: true })
      .setScrollFactor(0)
      .setDepth(2001);
    returnHit.on("pointerover", () => {
      returnBtn.clear();
      returnBtn.fillStyle(0x00f0ff, 0.1);
      returnBtn.fillRoundedRect(btnGap / 2, btnY - btnH / 2, btnW, btnH, 6);
      returnBtn.lineStyle(1, 0x00f0ff, 0.8);
      returnBtn.strokeRoundedRect(btnGap / 2, btnY - btnH / 2, btnW, btnH, 6);
    });
    returnHit.on("pointerout", () => {
      returnBtn.clear();
      returnBtn.lineStyle(1, 0x00f0ff, 0.6);
      returnBtn.strokeRoundedRect(btnGap / 2, btnY - btnH / 2, btnW, btnH, 6);
    });
    returnHit.on("pointerdown", () => {
      this.sound.stopByKey("gameAmbient");
      this.scene.start("MainMenuScene");
    });

    const cardContainer = this.add.container(cam.width / 2, cam.height / 2, [
      card,
      title,
      subtitle,
      redeployBtn,
      redeployText,
      returnBtn,
      returnText,
    ]);

    this.gameEndOverlay = this.add.container(0, 0, [backdrop, cardContainer]);
    this.gameEndOverlay.setDepth(2000);
    this.gameEndOverlay.setScrollFactor(0);

    // disable game input — keyboard movement, hover, clicks
    if (this.input.keyboard) {
      this.input.keyboard.enabled = false;
    }
  }

  preload(): void {
    // create soldier textures (4 directions each) — same spacesuit colors for all
    const suitBody = 0x8899aa;
    const suitVisor = 0x00f0ff;
    const suitHighlight = 0xbccdd8;
    const suitDark = 0x667788;
    this.createSoldierTextures(
      "player",
      suitBody,
      suitVisor,
      suitHighlight,
      suitDark,
    );
    this.createSoldierTextures(
      "otherPlayer",
      suitBody,
      suitVisor,
      suitHighlight,
      suitDark,
    );
    this.createChestTextures();
    this.createEscapeDoorTextures();
    this.createSwitchTextures();
    this.createMetalFloorTexture();
    this.createHullTexture();
    this.createEscapeParticleTexture();
  }

  private createHullTexture(): void {
    const size = 128;
    const canvas = document.createElement("canvas");
    canvas.width = size;
    canvas.height = size;
    const ctx = canvas.getContext("2d")!;

    // dark rusty brown-green industrial wall
    ctx.fillStyle = "#2e2a22";
    ctx.fillRect(0, 0, size, size);

    // heavy wall grain
    for (let i = 0; i < 5000; i++) {
      const x = Math.random() * size;
      const y = Math.random() * size;
      const brightness = 30 + Math.random() * 30;
      const r = brightness + Math.random() * 8;
      const g = brightness - 2 + Math.random() * 5;
      const b = brightness - 8;
      ctx.fillStyle = `rgba(${r}, ${g}, ${b}, ${0.15 + Math.random() * 0.2})`;
      ctx.fillRect(x, y, 1, 1);
    }

    // rust streaks - vertical drips
    for (let i = 0; i < 5; i++) {
      const sx = Math.random() * size;
      const sy = Math.random() * size * 0.3;
      const len = 20 + Math.random() * 40;
      ctx.strokeStyle = `rgba(60, 40, 25, ${0.1 + Math.random() * 0.12})`;
      ctx.lineWidth = 1 + Math.random() * 2;
      ctx.beginPath();
      ctx.moveTo(sx, sy);
      ctx.lineTo(sx + (Math.random() - 0.5) * 5, sy + len);
      ctx.stroke();
    }

    // horizontal wear lines - old paint
    for (let i = 0; i < 40; i++) {
      const y = Math.random() * size;
      const len = 10 + Math.random() * 50;
      const x = Math.random() * (size - len);
      ctx.strokeStyle = `rgba(50, 45, 35, ${0.15 + Math.random() * 0.2})`;
      ctx.lineWidth = 0.5;
      ctx.beginPath();
      ctx.moveTo(x, y);
      ctx.lineTo(x + len, y);
      ctx.stroke();
    }

    // scratches
    for (let i = 0; i < 8; i++) {
      const x1 = Math.random() * size;
      const y1 = Math.random() * size;
      const x2 = x1 + (Math.random() - 0.5) * 35;
      const y2 = y1 + (Math.random() - 0.5) * 35;
      ctx.strokeStyle = `rgba(55, 48, 38, ${0.15 + Math.random() * 0.15})`;
      ctx.lineWidth = 0.5;
      ctx.beginPath();
      ctx.moveTo(x1, y1);
      ctx.lineTo(x2, y2);
      ctx.stroke();
    }

    // panel edge highlight
    ctx.strokeStyle = "rgba(55, 48, 38, 0.6)";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(0, size - 1);
    ctx.lineTo(size - 1, size - 1);
    ctx.lineTo(size - 1, 0);
    ctx.stroke();
    ctx.strokeStyle = "rgba(15, 12, 8, 0.6)";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.lineTo(size - 1, 0);
    ctx.moveTo(0, 0);
    ctx.lineTo(0, size - 1);
    ctx.stroke();

    this.textures.addCanvas("hullMetal", canvas);
  }

  private createMetalFloorTexture(): void {
    const size = 128;
    const canvas = document.createElement("canvas");
    canvas.width = size;
    canvas.height = size;
    const ctx = canvas.getContext("2d")!;

    // base - worn concrete
    ctx.fillStyle = "#6d6f62";
    ctx.fillRect(0, 0, size, size);

    // concrete grain - uneven surface
    for (let i = 0; i < 5000; i++) {
      const x = Math.random() * size;
      const y = Math.random() * size;
      const brightness = 85 + Math.random() * 45;
      const g = brightness - 3 + Math.random() * 6;
      ctx.fillStyle = `rgba(${brightness}, ${g}, ${brightness - 8}, ${0.12 + Math.random() * 0.15})`;
      ctx.fillRect(x, y, 1, 1);
    }

    // darker patches - wear from foot traffic
    for (let i = 0; i < 6; i++) {
      const sx = Math.random() * size;
      const sy = Math.random() * size;
      const w = 10 + Math.random() * 25;
      const h = 3 + Math.random() * 8;
      ctx.fillStyle = `rgba(55, 53, 48, ${0.06 + Math.random() * 0.06})`;
      ctx.fillRect(sx, sy, w, h);
    }

    // thin cracks
    for (let i = 0; i < 3; i++) {
      const x1 = Math.random() * size;
      const y1 = Math.random() * size;
      ctx.strokeStyle = `rgba(50, 48, 42, ${0.12 + Math.random() * 0.12})`;
      ctx.lineWidth = 0.5;
      ctx.beginPath();
      ctx.moveTo(x1, y1);
      for (let j = 0; j < 3; j++) {
        const nx = x1 + (Math.random() - 0.5) * 50;
        const ny = y1 + (Math.random() - 0.3) * 50;
        ctx.lineTo(nx, ny);
      }
      ctx.stroke();
    }

    // panel edge highlight (bottom & right)
    ctx.strokeStyle = "rgba(95, 92, 80, 0.5)";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(0, size - 1);
    ctx.lineTo(size - 1, size - 1);
    ctx.lineTo(size - 1, 0);
    ctx.stroke();

    // panel edge shadow (top & left)
    ctx.strokeStyle = "rgba(45, 43, 38, 0.5)";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.lineTo(size - 1, 0);
    ctx.moveTo(0, 0);
    ctx.lineTo(0, size - 1);
    ctx.stroke();

    this.textures.addCanvas("metalFloor", canvas);
  }

  private createSoldierTextures(
    prefix: string,
    bodyColor: number,
    visorColor: number,
    highlightColor: number,
    _darkColor: number,
  ): void {
    const facings: Array<"down" | "up" | "left" | "right"> = [
      "down",
      "up",
      "left",
      "right",
    ];
    for (const facing of facings) {
      const g = this.make.graphics({});
      this.drawSoldierBody(
        g,
        30,
        30,
        facing,
        bodyColor,
        visorColor,
        highlightColor,
      );
      g.generateTexture(this.facingTextureKey(prefix, facing), 60, 60);
      g.destroy();
    }
  }

  private drawSoldierBody(
    g: Phaser.GameObjects.Graphics,
    cx: number,
    cy: number,
    facing: "up" | "down" | "left" | "right",
    bodyColor: number,
    visorColor: number,
    highlightColor: number,
  ): void {
    if (facing === "down" || facing === "up") {
      // front/back view — wider silhouette
      // shoulders
      g.fillStyle(bodyColor, 1);
      g.fillRoundedRect(cx - 15, cy - 6, 30, 21, 4);
      // helmet
      g.fillStyle(bodyColor, 1);
      g.fillRoundedRect(cx - 9, cy - 21, 18, 18, 6);
      // highlight on helmet
      g.fillStyle(highlightColor, 0.4);
      g.fillRoundedRect(cx - 6, cy - 18, 12, 9, 4);

      if (facing === "down") {
        // visor — two glowing dots
        g.fillStyle(visorColor, 1);
        g.fillCircle(cx - 5, cy - 10, 2.2);
        g.fillCircle(cx + 5, cy - 10, 2.2);
        // visor glow
        g.fillStyle(visorColor, 0.3);
        g.fillCircle(cx - 5, cy - 10, 4.5);
        g.fillCircle(cx + 5, cy - 10, 4.5);
      } else {
        // back-plate detail
        g.fillStyle(highlightColor, 0.2);
        g.fillRoundedRect(cx - 6, cy - 15, 12, 6, 3);
      }
      // torso
      g.fillStyle(bodyColor, 0.9);
      g.fillRoundedRect(cx - 12, cy + 3, 24, 15, 3);
    } else {
      // side view — narrower profile
      const dir = facing === "right" ? 1 : -1;
      // body
      g.fillStyle(bodyColor, 1);
      g.fillRoundedRect(cx - 6, cy - 6, 12, 24, 4);
      // helmet
      g.fillStyle(bodyColor, 1);
      g.fillRoundedRect(cx - 6, cy - 21, 15, 18, 6);
      // highlight
      g.fillStyle(highlightColor, 0.4);
      g.fillRoundedRect(cx - 3, cy - 18, 9, 9, 4);
      // shoulder bump
      g.fillStyle(bodyColor, 1);
      g.fillRoundedRect(cx - 7 + dir * 3, cy - 3, 15, 9, 3);
      // single visor dot
      g.fillStyle(visorColor, 1);
      g.fillCircle(cx + dir * 5, cy - 12, 2.2);
      g.fillStyle(visorColor, 0.3);
      g.fillCircle(cx + dir * 5, cy - 12, 4.5);
    }
  }

  private facingTextureKey(prefix: string, facing: string): string {
    return `${prefix}${facing.charAt(0).toUpperCase()}${facing.slice(1)}`;
  }

  private playAttackEffect(enemySprite: Phaser.Physics.Arcade.Sprite): void {
    if (!this.player) return;

    // --- 揮擊弧線 ---
    const slash = this.add.graphics();
    slash.setDepth(150);

    const px = this.player.x;
    const py = this.player.y;
    const angle = Phaser.Math.Angle.Between(
      px,
      py,
      enemySprite.x,
      enemySprite.y,
    );
    const radius = 35;

    slash.lineStyle(3, 0xffffff, 1);
    slash.beginPath();
    slash.arc(px, py, radius, angle - 0.8, angle + 0.8, false);
    slash.strokePath();

    // 弧線淡出
    this.tweens.add({
      targets: slash,
      alpha: 0,
      duration: 300,
      ease: "Power2",
      onComplete: () => slash.destroy(),
    });

    // --- 敵人閃紅 ---
    enemySprite.setTint(0xff0000);
    this.time.delayedCall(200, () => {
      enemySprite.clearTint();
    });
  }

  private drawLegs(
    graphics: Phaser.GameObjects.Graphics,
    x: number,
    y: number,
    facing: "up" | "down" | "left" | "right",
    walkPhase: number,
    isMoving: boolean,
    darkColor: number,
  ): void {
    graphics.clear();

    const legWidth = 6;
    const legHeight = 9;
    const swing = isMoving ? Math.sin(walkPhase) * 4.5 : 0;

    graphics.fillStyle(darkColor, 1);

    if (facing === "down" || facing === "up") {
      // two legs side by side, offset vertically when walking
      graphics.fillRect(x - 7, y + 18 + swing, legWidth, legHeight);
      graphics.fillRect(x + 1, y + 18 - swing, legWidth, legHeight);
    } else {
      // side view — legs overlap, offset horizontally when walking
      graphics.fillRect(x - 3 + swing, y + 18, legWidth, legHeight);
      graphics.fillRect(x - 3 - swing, y + 18, legWidth, legHeight);
    }
  }

  private createChestTextures(): void {
    const width = 40;
    const height = 32;

    // 關閉的寶箱
    const closed = this.make.graphics({});
    closed.fillStyle(0x8b4513, 1);
    closed.fillRect(0, 10, width, height - 10);
    closed.fillStyle(0x654321, 1);
    closed.fillRect(0, 0, width, 12);
    closed.fillStyle(0xffd700, 1);
    closed.fillRect(0, 10, width, 3);
    closed.fillRect(16, 6, 8, 10);
    closed.lineStyle(2, 0x3d2314, 1);
    closed.strokeRect(0, 0, width, height);
    closed.generateTexture("chest_closed", width, height);
    closed.destroy();

    // 打開的寶箱
    const open = this.make.graphics({});
    open.fillStyle(0x8b4513, 1);
    open.fillRect(0, 16, width, height - 16);
    open.fillStyle(0x654321, 1);
    open.fillRect(0, 0, width, 10);
    open.fillStyle(0xffec8b, 1);
    open.fillRect(4, 18, width - 8, height - 22);
    open.fillStyle(0xffd700, 1);
    open.fillRect(0, 16, width, 3);
    open.lineStyle(2, 0x3d2314, 1);
    open.strokeRect(0, 0, width, height);
    open.generateTexture("chest_open", width, height);
    open.destroy();
  }

  private createEscapeDoorTextures(): void {
    const size = 80;
    const centerX = size / 2;
    const centerY = size / 2;

    // ⚫ 鎖定的逃脫門 - 灰色魔法陣 (未啟動)
    const locked = this.make.graphics({});

    // 外圈 - 灰色
    locked.lineStyle(3, 0x666666, 0.8);
    locked.strokeCircle(centerX, centerY, 35);
    locked.strokeCircle(centerX, centerY, 30);

    // 內圈 - 灰色
    locked.lineStyle(2, 0x888888, 0.7);
    locked.strokeCircle(centerX, centerY, 20);

    // 魔法陣符文 (6個點)
    for (let i = 0; i < 6; i++) {
      const angle = (i / 6) * Math.PI * 2 - Math.PI / 2;
      const radius = 28;
      const x = centerX + Math.cos(angle) * radius;
      const y = centerY + Math.sin(angle) * radius;
      locked.fillStyle(0x666666, 0.8);
      locked.fillCircle(x, y, 3);
    }

    // 六芒星 (灰色)
    locked.lineStyle(2, 0x777777, 0.6);
    for (let i = 0; i < 6; i++) {
      const angle1 = (i / 6) * Math.PI * 2 - Math.PI / 2;
      const angle2 = ((i + 2) / 6) * Math.PI * 2 - Math.PI / 2;
      const radius = 25;
      const x1 = centerX + Math.cos(angle1) * radius;
      const y1 = centerY + Math.sin(angle1) * radius;
      const x2 = centerX + Math.cos(angle2) * radius;
      const y2 = centerY + Math.sin(angle2) * radius;
      locked.beginPath();
      locked.moveTo(x1, y1);
      locked.lineTo(x2, y2);
      locked.strokePath();
    }

    // 中心鎖圖示 (灰色)
    locked.fillStyle(0x555555, 1);
    locked.fillCircle(centerX, centerY, 8);
    locked.fillStyle(0x333333, 1);
    locked.fillCircle(centerX, centerY, 5);
    locked.fillCircle(centerX, centerY + 2, 2);

    locked.generateTexture("escape_door_locked", size, size);
    locked.destroy();

    // 🟢 解鎖的逃脫門 - 綠色魔法陣 (已解鎖但未啟動)
    const unlocked = this.make.graphics({});

    // 外圈 - 綠色發光
    unlocked.lineStyle(3, 0x44ff88, 0.9);
    unlocked.strokeCircle(centerX, centerY, 35);
    unlocked.lineStyle(2, 0x66ffaa, 0.7);
    unlocked.strokeCircle(centerX, centerY, 30);

    // 內圈 - 亮綠色
    unlocked.lineStyle(2, 0x88ffcc, 0.8);
    unlocked.strokeCircle(centerX, centerY, 20);

    // 發光光暈
    unlocked.fillStyle(0x44ff88, 0.15);
    unlocked.fillCircle(centerX, centerY, 35);

    // 魔法陣符文 (6個發光點)
    for (let i = 0; i < 6; i++) {
      const angle = (i / 6) * Math.PI * 2 - Math.PI / 2;
      const radius = 28;
      const x = centerX + Math.cos(angle) * radius;
      const y = centerY + Math.sin(angle) * radius;
      // 發光效果
      unlocked.fillStyle(0x44ff88, 0.3);
      unlocked.fillCircle(x, y, 5);
      unlocked.fillStyle(0x88ffcc, 1);
      unlocked.fillCircle(x, y, 3);
    }

    // 六芒星 (綠色發光)
    unlocked.lineStyle(2, 0x66ffaa, 0.7);
    for (let i = 0; i < 6; i++) {
      const angle1 = (i / 6) * Math.PI * 2 - Math.PI / 2;
      const angle2 = ((i + 2) / 6) * Math.PI * 2 - Math.PI / 2;
      const radius = 25;
      const x1 = centerX + Math.cos(angle1) * radius;
      const y1 = centerY + Math.sin(angle1) * radius;
      const x2 = centerX + Math.cos(angle2) * radius;
      const y2 = centerY + Math.sin(angle2) * radius;
      unlocked.beginPath();
      unlocked.moveTo(x1, y1);
      unlocked.lineTo(x2, y2);
      unlocked.strokePath();
    }

    // 中心圖示 - 解鎖符號 (亮綠色)
    unlocked.fillStyle(0xaaffdd, 1);
    unlocked.fillCircle(centerX, centerY, 8);
    unlocked.fillStyle(0x44ff88, 1);
    unlocked.fillCircle(centerX, centerY, 6);
    // 向上箭頭
    unlocked.fillStyle(0xffffff, 1);
    unlocked.fillTriangle(
      centerX,
      centerY - 4,
      centerX - 3,
      centerY + 2,
      centerX + 3,
      centerY + 2,
    );

    unlocked.generateTexture("escape_door_unlocked", size, size);
    unlocked.destroy();

    // ✨ 打開的逃脫門 - 激活的綠色魔法陣 (透明發光)
    const open = this.make.graphics({});

    // 最外層發光
    for (let i = 0; i < 4; i++) {
      const alpha = 0.2 - i * 0.04;
      const radius = 38 + i * 3;
      open.fillStyle(0x44ff88, alpha);
      open.fillCircle(centerX, centerY, radius);
    }

    // 外圈 - 強烈綠光
    open.lineStyle(4, 0x44ff88, 1);
    open.strokeCircle(centerX, centerY, 35);
    open.lineStyle(3, 0xaaffdd, 0.8);
    open.strokeCircle(centerX, centerY, 30);

    // 內圈 - 亮綠色
    open.lineStyle(3, 0xccffee, 0.9);
    open.strokeCircle(centerX, centerY, 20);

    // 傳送門中心 - 綠色帶透明
    open.fillStyle(0x66ffaa, 0.4);
    open.fillCircle(centerX, centerY, 30);
    open.fillStyle(0xaaffdd, 0.3);
    open.fillCircle(centerX, centerY, 20);

    // 魔法陣符文 (6個強烈發光點)
    for (let i = 0; i < 6; i++) {
      const angle = (i / 6) * Math.PI * 2 - Math.PI / 2;
      const radius = 28;
      const x = centerX + Math.cos(angle) * radius;
      const y = centerY + Math.sin(angle) * radius;
      // 強烈發光
      open.fillStyle(0x44ff88, 0.5);
      open.fillCircle(x, y, 6);
      open.fillStyle(0xffffff, 1);
      open.fillCircle(x, y, 3);
    }

    // 旋轉的六芒星 (強烈綠光)
    open.lineStyle(3, 0xaaffdd, 0.9);
    for (let i = 0; i < 6; i++) {
      const angle1 = (i / 6) * Math.PI * 2 - Math.PI / 2;
      const angle2 = ((i + 2) / 6) * Math.PI * 2 - Math.PI / 2;
      const radius = 25;
      const x1 = centerX + Math.cos(angle1) * radius;
      const y1 = centerY + Math.sin(angle1) * radius;
      const x2 = centerX + Math.cos(angle2) * radius;
      const y2 = centerY + Math.sin(angle2) * radius;
      open.beginPath();
      open.moveTo(x1, y1);
      open.lineTo(x2, y2);
      open.strokePath();
    }

    // 中心強烈發光
    open.fillStyle(0xffffff, 0.9);
    open.fillCircle(centerX, centerY, 10);
    open.fillStyle(0xaaffdd, 0.7);
    open.fillCircle(centerX, centerY, 15);
    open.fillStyle(0x44ff88, 0.4);
    open.fillCircle(centerX, centerY, 20);

    // 粒子效果 (8個旋轉的光點)
    for (let i = 0; i < 8; i++) {
      const angle = (i / 8) * Math.PI * 2;
      const radius = 18;
      const x = centerX + Math.cos(angle) * radius;
      const y = centerY + Math.sin(angle) * radius;
      open.fillStyle(0xffffff, 0.9);
      open.fillCircle(x, y, 2);
    }

    open.generateTexture("escape_door_open", size, size);
    open.destroy();
  }

  private createSwitchTextures(): void {
    const size = 30;

    // 未激活的開關 (灰色按鈕)
    const inactive = this.make.graphics({});
    // 底座
    inactive.fillStyle(0x404040, 1);
    inactive.fillRect(0, 0, size, size);
    // 按鈕
    inactive.fillStyle(0x808080, 1);
    inactive.fillCircle(size / 2, size / 2, size / 3);
    // 邊框
    inactive.lineStyle(2, 0x202020, 1);
    inactive.strokeRect(0, 0, size, size);
    inactive.generateTexture("switch_inactive", size, size);
    inactive.destroy();

    // 激活的開關 (綠色發光)
    const active = this.make.graphics({});
    // 底座
    active.fillStyle(0x404040, 1);
    active.fillRect(0, 0, size, size);
    // 按鈕發光效果
    active.fillStyle(0x00ff00, 0.4);
    active.fillCircle(size / 2, size / 2, size / 2.5);
    // 按鈕
    active.fillStyle(0x00ff00, 1);
    active.fillCircle(size / 2, size / 2, size / 3);
    // 邊框
    active.lineStyle(2, 0x00ff00, 1);
    active.strokeRect(0, 0, size, size);
    active.generateTexture("switch_active", size, size);
    active.destroy();
  }

  private createEscapeParticleTexture(): void {
    const g = this.add.graphics();
    g.fillStyle(0x00f0ff, 1);
    g.fillCircle(4, 4, 4);
    g.generateTexture("escape_particle", 8, 8);
    g.destroy();
  }

  private playEscapeParticles(x: number, y: number): void {
    const emitter = this.add.particles(x, y, "escape_particle", {
      speed: { min: 60, max: 180 },
      scale: { start: 1, end: 0 },
      alpha: { start: 1, end: 0 },
      lifespan: 800,
      quantity: 30,
      emitting: false,
      tint: [0x00f0ff, 0x4ecca3, 0xffd700],
    });
    emitter.setDepth(1000);
    emitter.explode(30);

    // clean up after animation
    this.time.delayedCall(1000, () => emitter.destroy());
  }

  private updateContainers(containers: ContainerState[]): void {
    const activeEntityIds = new Set(containers.map((c) => c.entity_id));

    // 移除不存在的寶箱
    this.chests.forEach((chest, entityId) => {
      if (!activeEntityIds.has(entityId)) {
        chest.sprite.destroy();
        this.chests.delete(entityId);
      }
    });

    // 新增或更新寶箱
    containers.forEach((container) => {
      let chest = this.chests.get(container.entity_id);

      if (!chest) {
        // 新增寶箱
        const sprite = this.add.sprite(
          container.position.x,
          container.position.y,
          container.is_open ? "chest_open" : "chest_closed",
        );
        sprite.setDepth(50);
        chest = { sprite, entityId: container.entity_id };
        this.chests.set(container.entity_id, chest);
      } else {
        // 更新寶箱狀態
        chest.sprite.setTexture(
          container.is_open ? "chest_open" : "chest_closed",
        );
        chest.sprite.setPosition(container.position.x, container.position.y);
      }

      // 如果是打開的寶箱，更新跳窗內容
      if (
        container.is_open &&
        this.openedChestEntityId === container.entity_id
      ) {
        this.updatePopupItems(container.items, container.entity_id);
      }
    });
  }

  private updateEscapeDoors(escapeDoors: EscapeDoorState[]): void {
    const activeEntityIds = new Set(escapeDoors.map((d) => d.entity_id));

    // 移除不存在的逃脫門
    this.escapeDoors.forEach((door, entityId) => {
      if (!activeEntityIds.has(entityId)) {
        door.sprite.destroy();
        this.escapeDoors.delete(entityId);
      }
    });

    // 新增或更新逃脫門
    escapeDoors.forEach((door) => {
      let escapeDoor = this.escapeDoors.get(door.entity_id);

      if (!escapeDoor) {
        // 根據狀態選擇 texture
        let texture = "escape_door_locked";
        if (door.is_open) {
          texture = "escape_door_open";
        } else if (!door.is_locked) {
          texture = "escape_door_unlocked";
        }

        // 新增逃脫門
        const sprite = this.add.sprite(
          door.position.x,
          door.position.y,
          texture,
        );
        sprite.setDepth(55); // 比寶箱稍高一點
        escapeDoor = { sprite, entityId: door.entity_id };
        this.escapeDoors.set(door.entity_id, escapeDoor);
      } else {
        // 更新逃脫門狀態
        let texture = "escape_door_locked";
        if (door.is_open) {
          texture = "escape_door_open";
        } else if (!door.is_locked) {
          texture = "escape_door_unlocked";
        }
        escapeDoor.sprite.setTexture(texture);
        escapeDoor.sprite.setPosition(door.position.x, door.position.y);
      }
    });
  }

  private updateSwitches(switches: SwitchState[]): void {
    const activeEntityIds = new Set(switches.map((s) => s.entity_id));

    // 移除不存在的開關
    this.switches.forEach((switchObj, entityId) => {
      if (!activeEntityIds.has(entityId)) {
        switchObj.sprite.destroy();
        this.switches.delete(entityId);
      }
    });

    // 新增或更新開關
    switches.forEach((switchState) => {
      let switchObj = this.switches.get(switchState.entity_id);

      if (!switchObj) {
        // 新增開關
        const sprite = this.add.sprite(
          switchState.position.x,
          switchState.position.y,
          switchState.is_activated ? "switch_active" : "switch_inactive",
        );
        sprite.setDepth(50);
        switchObj = { sprite, entityId: switchState.entity_id };
        this.switches.set(switchState.entity_id, switchObj);
      } else {
        // 更新開關狀態
        switchObj.sprite.setTexture(
          switchState.is_activated ? "switch_active" : "switch_inactive",
        );
        switchObj.sprite.setPosition(
          switchState.position.x,
          switchState.position.y,
        );
      }
    });
  }

  private updateWalls(walls: WallState[]): void {
    const activeEntityIds = new Set(walls.map((w) => w.entity_id));

    // 移除不存在的牆壁
    this.walls.forEach((wall, entityId) => {
      if (!activeEntityIds.has(entityId)) {
        wall.graphics.destroy();
        this.walls.delete(entityId);
      }
    });

    // 新增或更新牆壁
    walls.forEach((wallState) => {
      let wall = this.walls.get(wallState.entity_id);

      if (!wall) {
        const graphics = this.add.graphics();
        graphics.fillStyle(0x4a5568, 1);
        graphics.fillRect(
          wallState.position.x,
          wallState.position.y,
          wallState.width,
          wallState.height,
        );
        graphics.lineStyle(1, 0x6b7280, 0.6);
        graphics.strokeRect(
          wallState.position.x,
          wallState.position.y,
          wallState.width,
          wallState.height,
        );
        graphics.setDepth(50);
        wall = { graphics, entityId: wallState.entity_id };
        this.walls.set(wallState.entity_id, wall);
      }
    });

    // 從牆壁反推建築範圍，按 house_id 分組建立屋頂 + 地板（只做一次）
    if (!this.serverBuildingsCreated && walls.length > 0) {
      this.serverBuildingsCreated = true;

      // 按 house_id 分組
      const houseGroups = new Map<string, WallState[]>();
      walls.forEach((w) => {
        if (!w.house_id) return;
        const group = houseGroups.get(w.house_id) || [];
        group.push(w);
        houseGroups.set(w.house_id, group);
      });

      let buildingIndex = 0;
      houseGroups.forEach((houseWalls, houseId) => {
        // 算出這棟房子的 bounding box
        let minX = Infinity,
          minY = Infinity,
          maxX = -Infinity,
          maxY = -Infinity;
        houseWalls.forEach((w) => {
          minX = Math.min(minX, w.position.x);
          minY = Math.min(minY, w.position.y);
          maxX = Math.max(maxX, w.position.x + w.width);
          maxY = Math.max(maxY, w.position.y + w.height);
        });

        const bw = maxX - minX;
        const bh = maxY - minY;

        // 地板
        const floor = this.add.graphics();
        floor.fillStyle(0x2a3040, 1);
        floor.fillRect(minX, minY, bw, bh);
        floor.lineStyle(1, 0x3d4556, 0.4);
        for (let tx = minX; tx < maxX; tx += 40) {
          floor.lineBetween(tx, minY, tx, maxY);
        }
        for (let ty = minY; ty < maxY; ty += 40) {
          floor.lineBetween(minX, ty, maxX, ty);
        }
        floor.setDepth(1);

        // 屋頂
        const roof = this.add.graphics();
        roof.fillStyle(0x2d3748, 0.97);
        roof.fillRect(minX - 5, minY - 5, bw + 10, bh + 10);
        roof.lineStyle(2, 0x4a5568, 1);
        roof.strokeRect(minX - 5, minY - 5, bw + 10, bh + 10);
        roof.setDepth(200);

        // 入口標示（門在下方）
        const doorMarker = this.add.graphics();
        doorMarker.setDepth(250);
        const doorX = minX + bw / 2;
        const doorY = maxY + 5;
        const arrowSize = 10;
        doorMarker.fillStyle(0xffaa44, 1);
        doorMarker.fillTriangle(
          doorX,
          doorY - arrowSize,
          doorX - arrowSize,
          doorY + arrowSize,
          doorX + arrowSize,
          doorY + arrowSize,
        );
        doorMarker.lineStyle(3, 0xffaa44, 0.8);
        doorMarker.strokeCircle(doorX, doorY, 18);
        this.tweens.add({
          targets: doorMarker,
          alpha: 0.4,
          duration: 800,
          yoyo: true,
          repeat: -1,
          ease: "Sine.easeInOut",
        });

        this.outsideObjects.push(roof);

        const wallGroup = this.physics.add.staticGroup();
        const door = this.add.graphics();
        door.setDepth(51);
        const doorCollider = this.add.rectangle(0, 0, 0, 0);
        doorCollider.setVisible(false);

        const building: Building = {
          id: `server_building_${buildingIndex}`,
          x: minX,
          y: minY,
          width: bw,
          height: bh,
          doorSide: "bottom",
          wallGroup,
          roof,
          floor,
          doorMarker,
          door,
          doorCollider,
          isOpen: true,
        };
        this.buildings.push(building);
        buildingIndex++;
      });
    }
  }

  private updateDoors(doors: DoorState[]): void {
    const activeEntityIds = new Set(doors.map((d) => d.entity_id));

    // 移除不存在的門
    this.serverDoors.forEach((door, entityId) => {
      if (!activeEntityIds.has(entityId)) {
        door.rect.destroy();
        this.serverDoors.delete(entityId);
      }
    });

    // 新增或更新門
    doors.forEach((doorState) => {
      let door = this.serverDoors.get(doorState.entity_id);

      if (!door) {
        const rect = this.add.rectangle(
          doorState.position.x,
          doorState.position.y + doorState.height / 2,
          doorState.width,
          doorState.height,
          0x5a6577,
        );
        rect.setOrigin(0, 0.5);
        rect.setStrokeStyle(2, 0x6b7280);
        rect.setDepth(51);

        door = { rect, entityId: doorState.entity_id, isOpen: false };
        this.serverDoors.set(doorState.entity_id, door);

        if (doorState.is_open) {
          door.isOpen = true;
          rect.setRotation(Math.PI / 2);
        }
      } else if (door.isOpen !== doorState.is_open) {
        door.isOpen = doorState.is_open;
        const targetRotation = doorState.is_open ? Math.PI / 2 : 0;

        this.tweens.add({
          targets: door.rect,
          rotation: targetRotation,
          duration: 300,
          ease: "Power2",
        });
      }
    });
  }

  private getNearbyDoor(): { entityId: string } | null {
    if (!this.player) return null;
    const interactDistance = 60;

    for (const [entityId, door] of this.serverDoors) {
      const distance = Phaser.Math.Distance.Between(
        this.player.x,
        this.player.y,
        door.rect.x + door.rect.width / 2,
        door.rect.y,
      );
      if (distance < interactDistance) {
        return { entityId };
      }
    }
    return null;
  }

  private toggleChest(entityId: string): void {
    // 發送互動請求到後端
    socketManager.sendMessage(ActionType.Interact, {
      entity_id: entityId,
    });

    // 如果是關閉跳窗
    if (this.isPopupOpen && this.openedChestEntityId === entityId) {
      this.hideChestPopup();
      this.openedChestEntityId = undefined;
    } else {
      // 開啟跳窗
      this.openedChestEntityId = entityId;
      this.showChestPopup();

      // If container already has items from server state, populate immediately
      const gameState = this.lastGameState;
      if (gameState) {
        const container = gameState.containers?.find(
          (c) => c.entity_id === entityId,
        );
        if (container && container.is_open && container.items?.length > 0) {
          this.updatePopupItems(container.items, entityId);
        }
      }
    }
  }

  private interactWithSwitch(entityId: string): void {
    console.log("Interacting with switch:", entityId);
    // 發送互動請求到後端
    socketManager.sendMessage(ActionType.Interact, {
      entity_id: entityId,
    });
  }

  private interactWithEscapeDoor(entityId: string): void {
    console.log("Interacting with escape door:", entityId);
    // 發送互動請求到後端
    socketManager.sendMessage(ActionType.Interact, {
      entity_id: entityId,
    });
  }

  private checkChestDistance(): void {
    if (!this.player || !this.openedChestEntityId || !this.isPopupOpen) return;

    const chest = this.chests.get(this.openedChestEntityId);
    if (!chest) return;

    const distance = Phaser.Math.Distance.Between(
      this.player.x,
      this.player.y,
      chest.sprite.x,
      chest.sprite.y,
    );

    const interactDistance = 60;
    if (distance > interactDistance) {
      // Just close popup locally, let backend state control chest visual
      this.hideChestPopup();
      this.openedChestEntityId = undefined;
    }
  }

  private showChestPopup(): void {
    if (this.isPopupOpen) return;

    const centerX = this.cameras.main.width / 2;
    const centerY = this.cameras.main.height / 2;
    const popupWidth = 320;
    const popupHeight = 280;

    const bg = this.add.graphics();
    bg.fillStyle(0x0a0a12, 0.9);
    bg.fillRoundedRect(
      -popupWidth / 2,
      -popupHeight / 2,
      popupWidth,
      popupHeight,
      8,
    );
    bg.lineStyle(1, 0x00f0ff, 1);
    bg.strokeRoundedRect(
      -popupWidth / 2,
      -popupHeight / 2,
      popupWidth,
      popupHeight,
      8,
    );

    const title = this.add.text(0, -popupHeight / 2 + 20, "CONTAINER", {
      fontSize: "18px",
      color: "#00f0ff",
      letterSpacing: 6,
    });
    title.setOrigin(0.5);

    // Placeholder for empty/loading state
    this.popupItemsText = this.add.text(0, 0, "Loading...", {
      fontSize: "14px",
      color: "#556677",
      align: "center",
    });
    this.popupItemsText.setOrigin(0.5);

    const hint = this.add.text(
      0,
      popupHeight / 2 - 25,
      "Q Close  //  F Take Item",
      {
        fontSize: "12px",
        color: "#334455",
      },
    );
    hint.setOrigin(0.5);

    this.chestPopup = this.add.container(centerX, centerY, [
      bg,
      title,
      this.popupItemsText,
      hint,
    ]);
    this.chestPopup.setDepth(1000);
    this.chestPopup.setScrollFactor(0);

    this.isPopupOpen = true;
  }

  private updatePopupItems(items: ItemState[], entityId?: string): void {
    if (!this.chestPopup) return;

    const chestId = entityId || this.openedChestEntityId;
    if (!chestId) return;

    const now = Date.now();

    // Filter out items that are still pending pickup (sent interact, awaiting server confirmation)
    const displayItems = items.filter((item) => {
      const lootedAt = this.chestLootedAtMap.get(item.entity_id);
      if (lootedAt && now - lootedAt < this.PENDING_DURATION) {
        return false;
      }
      if (lootedAt) {
        this.chestLootedAtMap.delete(item.entity_id);
      }
      return true;
    });

    this.currentChestItems = displayItems.map((item) => ({ ...item }));

    // Skip rebuild if items haven't changed (prevents hover flicker from game loop)
    const fingerprint = displayItems.map((i) => i.entity_id).join(",");
    if (fingerprint === this.chestItemFingerprint) return;
    this.chestItemFingerprint = fingerprint;

    this.clearItemRows("chest");

    if (displayItems.length === 0) {
      if (this.popupItemsText) {
        this.popupItemsText.setText("(Empty)");
        this.popupItemsText.setVisible(true);
      }
    } else {
      if (this.popupItemsText) this.popupItemsText.setVisible(false);
      this.createItemRows(displayItems, this.chestPopup, -50, "chest");
    }
  }

  private hideChestPopup(): void {
    this.clearItemRows("chest");
    this.chestItemFingerprint = "";
    if (this.chestPopup) {
      this.chestPopup.destroy();
      this.chestPopup = undefined;
    }
    this.popupItemsText = undefined;
    this.isPopupOpen = false;
    this.currentChestItems = [];
    // 不清除 chestLootedAtMap，讓後端確認時自動清除
  }

  // === 道具欄功能 ===

  private toggleInventory(): void {
    if (!this.equipmentPanel) return;
    this.equipmentPanel.toggle();
    if (this.equipmentPanel.isVisible()) {
      this.equipmentPanel.updateInventory(this.inventoryItems);
      this.equipmentPanel.updateEquipment(this.equippedItems);
    }
  }

  // === Item row grid system with manual hit testing ===

  private clearItemRows(source: "chest" | "all"): void {
    this.hideItemTooltip();
    this.hoveredRowIndex = -1;
    const remaining: typeof this.itemRows = [];
    for (const row of this.itemRows) {
      if (source === "all" || row.source === source) {
        row.label.destroy();
        row.rowBg.destroy();
      } else {
        remaining.push(row);
      }
    }
    this.itemRows = remaining;
  }

  private formatItemLine(item: ItemState): string {
    const tag = this.getItemStatTag(item);
    return tag ? `${item.name}  ${tag}` : `${item.name} x${item.quantity}`;
  }

  private getItemStatTag(item: ItemState): string {
    if (item.attack_power) return `ATK ${item.attack_power}`;
    if (item.defense_rating) return `DEF ${item.defense_rating}`;
    if (item.healing_amount) return `+${item.healing_amount} HP`;
    if (item.mana_amount) return `+${item.mana_amount} MP`;
    return "";
  }

  private createItemRows(
    items: ItemState[],
    container: Phaser.GameObjects.Container,
    startY: number,
    source: "chest",
  ): void {
    const rowHeight = 28;
    const popupWidth = 320;
    const rowWidth = popupWidth - 16;
    const containerX = container.x;
    const containerY = container.y;

    items.forEach((item, i) => {
      // localY = top edge of row in container-local coords
      const rowTop = startY + i * rowHeight;
      const rowCenterY = rowTop + rowHeight / 2;

      // Row background inside container
      const rowBg = this.add.graphics();
      this.drawRowBg(rowBg, i, rowWidth, rowHeight, rowTop, false);
      container.add(rowBg);

      // Text label centered in row
      const label = this.add.text(0, rowCenterY, this.formatItemLine(item), {
        fontSize: "13px",
        color: "#ccdde8",
      });
      label.setOrigin(0.5);
      container.add(label);

      // Screen-space rect for manual hit testing
      const screenRect = {
        x: containerX - rowWidth / 2,
        y: containerY + rowTop,
        w: rowWidth,
        h: rowHeight,
      };

      this.itemRows.push({ screenRect, item, label, rowBg, source });
    });

    // If we had a hovered item before rebuild, restore hover state
    if (this.hoveredItemEntityId) {
      this.restoreHoverState();
    }
  }

  /** Restore hover after rows are rebuilt (e.g. item was looted from chest, triggering rebuild) */
  private restoreHoverState(): void {
    for (let i = 0; i < this.itemRows.length; i++) {
      if (this.itemRows[i].item.entity_id === this.hoveredItemEntityId) {
        this.hoveredRowIndex = i;
        this.applyRowHover(i);
        this.showItemTooltip(
          this.itemRows[i].item,
          this.lastPointerX,
          this.lastPointerY,
        );
        return;
      }
    }
    // Item no longer exists (was looted etc.)
    this.hoveredItemEntityId = undefined;
    this.hoveredRowIndex = -1;
  }

  private drawRowBg(
    g: Phaser.GameObjects.Graphics,
    index: number,
    rowWidth: number,
    rowHeight: number,
    rowTop: number,
    hovered: boolean,
  ): void {
    g.clear();
    if (hovered) {
      g.fillStyle(0x00f0ff, 0.08);
      g.fillRoundedRect(-rowWidth / 2, rowTop, rowWidth, rowHeight, 4);
      g.lineStyle(1, 0x00f0ff, 0.2);
      g.strokeRoundedRect(-rowWidth / 2, rowTop, rowWidth, rowHeight, 4);
    } else {
      const bgAlpha = index % 2 === 0 ? 0.25 : 0.15;
      g.fillStyle(0x112233, bgAlpha);
      g.fillRoundedRect(-rowWidth / 2, rowTop, rowWidth, rowHeight, 4);
      g.lineStyle(1, 0x00f0ff, 0.06);
      g.lineBetween(
        -rowWidth / 2 + 8,
        rowTop + rowHeight,
        rowWidth / 2 - 8,
        rowTop + rowHeight,
      );
    }
  }

  private getRowLocalTop(row: (typeof this.itemRows)[0]): number {
    return row.screenRect.y - (this.chestPopup?.y ?? 0);
  }

  private getRowWidth(_row: (typeof this.itemRows)[0]): number {
    return 320 - 16;
  }

  private applyRowHover(index: number): void {
    const row = this.itemRows[index];
    row.label.setColor("#00f0ff");
    this.drawRowBg(
      row.rowBg,
      index,
      this.getRowWidth(row),
      row.screenRect.h,
      this.getRowLocalTop(row),
      true,
    );
  }

  private applyRowUnhover(index: number): void {
    const row = this.itemRows[index];
    row.label.setColor("#ccdde8");
    this.drawRowBg(
      row.rowBg,
      index,
      this.getRowWidth(row),
      row.screenRect.h,
      this.getRowLocalTop(row),
      false,
    );
  }

  private handleItemRowHover(pointerX: number, pointerY: number): void {
    this.lastPointerX = pointerX;
    this.lastPointerY = pointerY;

    let foundIndex = -1;
    for (let i = 0; i < this.itemRows.length; i++) {
      const { screenRect } = this.itemRows[i];
      if (
        pointerX >= screenRect.x &&
        pointerX <= screenRect.x + screenRect.w &&
        pointerY >= screenRect.y &&
        pointerY <= screenRect.y + screenRect.h
      ) {
        foundIndex = i;
        break;
      }
    }

    if (foundIndex === this.hoveredRowIndex) {
      // Same row — just move tooltip
      if (foundIndex !== -1) {
        this.moveItemTooltip(pointerX, pointerY);
      }
      return;
    }

    // Unhover previous
    if (
      this.hoveredRowIndex !== -1 &&
      this.hoveredRowIndex < this.itemRows.length
    ) {
      this.applyRowUnhover(this.hoveredRowIndex);
      this.hideItemTooltip();
    }

    this.hoveredRowIndex = foundIndex;
    this.hoveredItemEntityId =
      foundIndex !== -1 ? this.itemRows[foundIndex].item.entity_id : undefined;

    // Hover new
    if (foundIndex !== -1) {
      this.applyRowHover(foundIndex);
      this.showItemTooltip(this.itemRows[foundIndex].item, pointerX, pointerY);
    }
  }

  private getItemType(
    item: ItemState,
  ): "weapon" | "armor" | "consumable" | "unknown" {
    if (item.attack_power || item.weapon_type) return "weapon";
    if (item.defense_rating || item.armor_slot) return "armor";
    if (item.healing_amount || item.mana_amount) return "consumable";
    return "unknown";
  }

  private buildTooltipContent(item: ItemState): {
    lines: { label: string; value: string; color: string }[];
    typeLabel: string;
    typeColor: string;
  } {
    const type = this.getItemType(item);
    const lines: { label: string; value: string; color: string }[] = [];

    switch (type) {
      case "weapon": {
        const typeColor = "#ff4466";
        if (item.weapon_type)
          lines.push({
            label: "TYPE",
            value: item.weapon_type.toUpperCase(),
            color: "#99aabb",
          });
        if (item.attack_power)
          lines.push({
            label: "ATK",
            value: `${item.attack_power}`,
            color: typeColor,
          });
        if (item.critical_rate)
          lines.push({
            label: "CRIT",
            value: `${Math.round(item.critical_rate)}%`,
            color: "#ffaa33",
          });
        return { lines, typeLabel: "WEAPON", typeColor };
      }
      case "armor": {
        const typeColor = "#44aaff";
        if (item.armor_slot)
          lines.push({
            label: "SLOT",
            value: item.armor_slot.toUpperCase(),
            color: "#99aabb",
          });
        if (item.defense_rating)
          lines.push({
            label: "DEF",
            value: `${item.defense_rating}`,
            color: typeColor,
          });
        return { lines, typeLabel: "ARMOR", typeColor };
      }
      case "consumable": {
        const typeColor = "#44ff88";
        if (item.healing_amount)
          lines.push({
            label: "HEAL",
            value: `+${item.healing_amount} HP`,
            color: typeColor,
          });
        if (item.mana_amount)
          lines.push({
            label: "MANA",
            value: `+${item.mana_amount} MP`,
            color: "#aa88ff",
          });
        return { lines, typeLabel: "CONSUMABLE", typeColor };
      }
      default:
        return { lines, typeLabel: "ITEM", typeColor: "#556677" };
    }
  }

  private showItemTooltip(
    item: ItemState,
    screenX: number,
    screenY: number,
  ): void {
    this.hideItemTooltip();

    const { lines, typeLabel, typeColor } = this.buildTooltipContent(item);
    const padding = 14;
    const tooltipWidth = 220;

    const children: Phaser.GameObjects.GameObject[] = [];
    let curY = padding;

    // Item name
    const nameText = this.add.text(padding, curY, item.name, {
      fontSize: "15px",
      color: "#00f0ff",
      fontStyle: "bold",
    });
    children.push(nameText);
    curY += 22;

    // Type badge
    const typeText = this.add.text(padding, curY, typeLabel, {
      fontSize: "10px",
      color: typeColor,
      letterSpacing: 3,
    });
    children.push(typeText);
    curY += 20;

    // Separator line
    const sep = this.add.graphics();
    sep.lineStyle(1, 0x00f0ff, 0.15);
    sep.lineBetween(padding, curY, tooltipWidth - padding, curY);
    children.push(sep);
    curY += 10;

    // Stat rows
    for (const line of lines) {
      const labelText = this.add.text(padding, curY, line.label, {
        fontSize: "12px",
        color: "#556677",
        letterSpacing: 2,
      });
      const valueText = this.add.text(
        tooltipWidth - padding,
        curY,
        line.value,
        {
          fontSize: "13px",
          color: line.color,
        },
      );
      valueText.setOrigin(1, 0);
      children.push(labelText, valueText);
      curY += 20;
    }

    // Description
    if (item.description) {
      curY += 6;
      const descSep = this.add.graphics();
      descSep.lineStyle(1, 0x00f0ff, 0.1);
      descSep.lineBetween(padding, curY, tooltipWidth - padding, curY);
      children.push(descSep);
      curY += 8;
      const desc = this.add.text(padding, curY, item.description, {
        fontSize: "11px",
        color: "#667788",
        wordWrap: { width: tooltipWidth - padding * 2 },
        lineSpacing: 4,
      });
      children.push(desc);
      curY += desc.height;
    }

    // Quantity (if >1)
    if (item.quantity > 1) {
      curY += 6;
      const qtyText = this.add.text(
        tooltipWidth - padding,
        curY,
        `x${item.quantity}`,
        {
          fontSize: "11px",
          color: "#556677",
        },
      );
      qtyText.setOrigin(1, 0);
      children.push(qtyText);
      curY += 16;
    }

    const tooltipHeight = curY + padding;

    // Background (drawn first, inserted at index 0)
    const bg = this.add.graphics();
    bg.fillStyle(0x080810, 0.95);
    bg.fillRoundedRect(0, 0, tooltipWidth, tooltipHeight, 6);
    bg.lineStyle(
      1,
      typeColor === "#556677" ? 0x00f0ff : parseInt(typeColor.slice(1), 16),
      0.4,
    );
    bg.strokeRoundedRect(0, 0, tooltipWidth, tooltipHeight, 6);
    children.unshift(bg);

    this.itemTooltip = this.add.container(screenX + 14, screenY - 10, children);
    this.itemTooltip.setDepth(2000);
    this.itemTooltip.setScrollFactor(0);

    // Keep tooltip on screen
    const cam = this.cameras.main;
    if (screenX + 14 + tooltipWidth > cam.width) {
      this.itemTooltip.setX(screenX - tooltipWidth - 8);
    }
    if (screenY - 10 + tooltipHeight > cam.height) {
      this.itemTooltip.setY(screenY - tooltipHeight - 8);
    }
  }

  private moveItemTooltip(screenX: number, screenY: number): void {
    if (!this.itemTooltip) return;
    this.itemTooltip.setPosition(screenX + 14, screenY - 10);
  }

  private hideItemTooltip(): void {
    if (this.itemTooltip) {
      this.itemTooltip.destroy();
      this.itemTooltip = undefined;
    }
  }

  private syncInventory(serverInventory: ItemState[]): void {
    const now = Date.now();

    // 建立後端物品 Map (by entity_id)
    const serverItemMap = new Map<string, ItemState>();
    for (const item of serverInventory) {
      serverItemMap.set(item.entity_id, item);
    }

    // 過濾本地物品：保留後端有的 + pending 中的
    const newInventory: ItemState[] = [];

    for (const localItem of this.inventoryItems) {
      const isPending =
        localItem.lootedAt && now - localItem.lootedAt < this.PENDING_DURATION;

      if (serverItemMap.has(localItem.entity_id)) {
        // 後端有，使用後端資料（清除 pending 狀態）
        const serverItem = serverItemMap.get(localItem.entity_id)!;
        newInventory.push({
          ...serverItem,
          lootedAt: undefined, // 後端確認後清除 pending
        });
        serverItemMap.delete(localItem.entity_id);
      } else if (isPending) {
        // 後端沒有，但還在 pending 中，保留本地的
        newInventory.push(localItem);
      }
      // 後端沒有且不是 pending → 不保留（被移除了）
    }

    // 加入後端有但本地沒有的（其他來源的物品）
    for (const item of serverItemMap.values()) {
      newInventory.push(item);
    }

    this.inventoryItems = newInventory;

    // 更新裝備面板
    if (this.equipmentPanel?.isVisible()) {
      this.equipmentPanel.updateInventory(this.inventoryItems);
    }
  }

  // Map backend EquipmentState (chest/gloves/legs) → local EquippedItems (body/hands/feet).
  private syncEquipment(serverEquipment: EquipmentState): void {
    this.equippedItems = {
      weapon: serverEquipment.weapon,
      head: serverEquipment.head,
      body: serverEquipment.chest,
      hands: serverEquipment.gloves,
      feet: serverEquipment.legs,
      ring_1: serverEquipment.ring_1,
      ring_2: serverEquipment.ring_2,
      consumable_1: serverEquipment.consumable_1,
      consumable_2: serverEquipment.consumable_2,
      consumable_3: serverEquipment.consumable_3,
    };

    if (this.equipmentPanel?.isVisible()) {
      this.equipmentPanel.updateEquipment(this.equippedItems);
    }
  }

  private pickupSingleItemFromChest(): void {
    if (
      !this.isPopupOpen ||
      this.currentChestItems.length === 0 ||
      !this.openedChestEntityId
    ) {
      return;
    }

    const item = this.currentChestItems[0];

    socketManager.sendMessage(ActionType.Interact, {
      entity_id: item.entity_id,
    });

    // Optimistic update: remove from chest, add to inventory
    this.currentChestItems = this.currentChestItems.filter(
      (i) => i.entity_id !== item.entity_id,
    );

    const now = Date.now();
    this.inventoryItems.push({
      ...item,
      lootedAt: now,
    });

    this.chestLootedAtMap.set(item.entity_id, now);

    this.updatePopupItems(this.currentChestItems);

    if (this.equipmentPanel?.isVisible()) {
      this.equipmentPanel.updateInventory(this.inventoryItems);
    }
  }

  private getNearbyChest(): { entityId: string } | null {
    if (!this.player) return null;
    const interactDistance = 60;

    for (const [entityId, chest] of this.chests) {
      const distance = Phaser.Math.Distance.Between(
        this.player.x,
        this.player.y,
        chest.sprite.x,
        chest.sprite.y,
      );
      if (distance < interactDistance) {
        return { entityId };
      }
    }
    return null;
  }

  private getNearbySwitch(): { entityId: string } | null {
    if (!this.player) return null;
    const interactDistance = 60;

    for (const [entityId, switchObj] of this.switches) {
      const distance = Phaser.Math.Distance.Between(
        this.player.x,
        this.player.y,
        switchObj.sprite.x,
        switchObj.sprite.y,
      );
      if (distance < interactDistance) {
        return { entityId };
      }
    }
    return null;
  }

  private getNearbyEscapeDoor(): { entityId: string } | null {
    if (!this.player) return null;
    const interactDistance = 60;

    for (const [entityId, escapeDoor] of this.escapeDoors) {
      const distance = Phaser.Math.Distance.Between(
        this.player.x,
        this.player.y,
        escapeDoor.sprite.x,
        escapeDoor.sprite.y,
      );
      if (distance < interactDistance) {
        return { entityId };
      }
    }
    return null;
  }

  private createPlayer(x: number, y: number, username?: string): void {
    this.player = this.physics.add.sprite(x, y, "playerDown");
    this.player.setCollideWorldBounds(true);
    this.player.setDepth(100);

    // set circular physics body to match backend collision (radius 20), offset for 60x60 texture
    this.player.body.setCircle(20, 10, 10);

    // create legs overlay that will follow player
    this.playerLegs = this.add.graphics();
    this.playerLegs.setDepth(101);
    this.playerFacing = "down";
    this.walkPhase = 0;
    this.drawLegs(this.playerLegs, x, y, "down", 0, false, 0x667788);

    // username label above player
    this.playerNameText = this.add.text(x, y - 35, username || "You", {
      fontSize: "11px",
      fontFamily: "Orbitron, monospace",
      color: "#00f0ff",
      stroke: "#0a0a12",
      strokeThickness: 3,
      align: "center",
    });
    this.playerNameText.setOrigin(0.5, 1);
    this.playerNameText.setDepth(102);

    // 玩家與所有建築牆壁/門碰撞
    this.buildings.forEach((building) => {
      this.physics.add.collider(this.player!, building.wallGroup);
      this.physics.add.collider(this.player!, building.doorCollider);
    });

    // 相機跟隨玩家
    this.cameras.main.startFollow(this.player, true, 0.1, 0.1);
  }

  private defaultCursorCSS = "";
  private crosshairCursorCSS = "";

  private setupCustomCursor(): void {
    // --- Default cursor: tech/Deus Ex style pointer ---
    const dSize = 32;
    const dc = document.createElement("canvas");
    dc.width = dSize;
    dc.height = dSize;
    const dCtx = dc.getContext("2d")!;

    // angled pointer — top-left origin
    dCtx.strokeStyle = "rgba(0, 240, 255, 0.9)";
    dCtx.lineWidth = 1.5;
    dCtx.fillStyle = "rgba(0, 240, 255, 0.12)";

    // pointer triangle
    dCtx.beginPath();
    dCtx.moveTo(6, 4);
    dCtx.lineTo(6, 22);
    dCtx.lineTo(12, 17);
    dCtx.lineTo(17, 24);
    dCtx.lineTo(20, 22);
    dCtx.lineTo(15, 15);
    dCtx.lineTo(21, 13);
    dCtx.closePath();
    dCtx.fill();
    dCtx.stroke();

    // small corner brackets — top-left
    dCtx.strokeStyle = "rgba(0, 240, 255, 0.4)";
    dCtx.lineWidth = 1;
    dCtx.beginPath();
    dCtx.moveTo(2, 10);
    dCtx.lineTo(2, 2);
    dCtx.lineTo(10, 2);
    dCtx.stroke();

    this.defaultCursorCSS = `url(${dc.toDataURL()}) 6 4, default`;
    this.input.setDefaultCursor(this.defaultCursorCSS);

    // --- Crosshair cursor: pink, for targeting other players ---
    const cSize = 32;
    const cMid = cSize / 2;
    const cc = document.createElement("canvas");
    cc.width = cSize;
    cc.height = cSize;
    const cCtx = cc.getContext("2d")!;

    // outer ring glow
    cCtx.strokeStyle = "rgba(255, 0, 170, 0.3)";
    cCtx.lineWidth = 2;
    cCtx.beginPath();
    cCtx.arc(cMid, cMid, 10, 0, Math.PI * 2);
    cCtx.stroke();

    // crosshair lines
    cCtx.strokeStyle = "rgba(255, 0, 170, 0.85)";
    cCtx.lineWidth = 1.5;
    const gap = 4;
    const len = 6;
    cCtx.beginPath();
    cCtx.moveTo(cMid, cMid - gap - len);
    cCtx.lineTo(cMid, cMid - gap);
    cCtx.stroke();
    cCtx.beginPath();
    cCtx.moveTo(cMid, cMid + gap);
    cCtx.lineTo(cMid, cMid + gap + len);
    cCtx.stroke();
    cCtx.beginPath();
    cCtx.moveTo(cMid - gap - len, cMid);
    cCtx.lineTo(cMid - gap, cMid);
    cCtx.stroke();
    cCtx.beginPath();
    cCtx.moveTo(cMid + gap, cMid);
    cCtx.lineTo(cMid + gap + len, cMid);
    cCtx.stroke();

    // center dot
    cCtx.fillStyle = "#ff00aa";
    cCtx.beginPath();
    cCtx.arc(cMid, cMid, 1.5, 0, Math.PI * 2);
    cCtx.fill();

    this.crosshairCursorCSS = `url(${cc.toDataURL()}) ${cMid} ${cMid}, crosshair`;
  }

  create(): void {
    // custom crosshair cursor
    this.setupCustomCursor();

    // Game ambient music — stop menu theme, start game ambient
    this.sound.stopByKey("menuTheme");
    if (!this.sound.get("gameAmbient")?.isPlaying) {
      this.sound.play("gameAmbient", { loop: true, volume: 0.25 });
    }

    // Connect via SocketManager
    this.connectToServer();

    // setup world boundaries
    this.physics.world.setBounds(0, 0, this.mapWidth, this.mapHeight);

    // create map background with cosmic theme
    this.createMapBackground();

    // buildings are now created from server wall data in updateWalls()

    // 寶箱由後端同步，不在這裡創建

    // 設置相機邊界（擴大讓玩家能看到船外太空）
    const outerMargin = 200;
    const spaceMargin = 150; // extra space beyond hull visible at edges
    this.cameras.main.setBounds(
      -outerMargin - spaceMargin,
      -outerMargin - spaceMargin,
      this.mapWidth + (outerMargin + spaceMargin) * 2,
      this.mapHeight + (outerMargin + spaceMargin) * 2,
    );

    // 輸入控制
    this.cursors = this.input.keyboard!.createCursorKeys();
    this.wasd = {
      up: this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.W),
      down: this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.S),
      left: this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.A),
      right: this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.D),
    };

    // ESC 返回主選單
    this.input.keyboard?.on("keydown-ESC", () => {
      this.sound.stopByKey("gameAmbient");
      this.scene.start("MainMenuScene");
    });

    // E 鍵互動（門、寶箱、開關、逃脫門）
    this.input.keyboard?.on("keydown-E", () => {
      // When the equipment panel is open, E equips/unequips the hovered item.
      if (this.equipmentPanel?.isVisible()) {
        this.equipmentPanel.handleEquipKey();
        return;
      }

      // 檢查後端門
      const nearbyDoor = this.getNearbyDoor();
      if (nearbyDoor) {
        socketManager.sendMessage(ActionType.Interact, {
          entity_id: nearbyDoor.entityId,
        });
        return;
      }
      // 檢查開關
      const nearbySwitch = this.getNearbySwitch();
      if (nearbySwitch) {
        this.interactWithSwitch(nearbySwitch.entityId);
        return;
      }
      // 檢查逃脫門
      const nearbyEscapeDoor = this.getNearbyEscapeDoor();
      if (nearbyEscapeDoor) {
        this.interactWithEscapeDoor(nearbyEscapeDoor.entityId);
        return;
      }
      // 檢查寶箱
      const nearbyChest = this.getNearbyChest();
      if (nearbyChest) {
        this.toggleChest(nearbyChest.entityId);
      }
    });

    // I 鍵開啟/關閉道具欄
    this.input.keyboard?.on("keydown-I", () => {
      this.toggleInventory();
    });

    // F 鍵從寶箱取得道具
    this.input.keyboard?.on("keydown-F", () => {
      this.pickupSingleItemFromChest();
    });

    // Q 鍵關閉任何打開的彈窗
    this.input.keyboard?.on("keydown-Q", () => {
      if (this.equipmentPanel?.isVisible()) {
        this.equipmentPanel.hide();
        return;
      }
      if (this.isPopupOpen) {
        this.hideChestPopup();
        this.openedChestEntityId = undefined;
      }
    });

    // H 鍵顯示/隱藏操作說明
    this.input.keyboard?.on("keydown-H", () => {
      this.toggleControlsPanel();
    });

    // Scene-level pointer tracking for item row hover (bypasses broken Phaser scrollFactor input)
    this.input.on("pointermove", (pointer: Phaser.Input.Pointer) => {
      this.handleItemRowHover(pointer.x, pointer.y);
    });

    // Disable browser right-click menu for equipment panel context menus
    this.input.mouse?.disableContextMenu();

    // Initialize equipment panel
    this.equipmentPanel = new EquipmentPanel(this);
    this.equipmentPanel.onEquip = (item, slot) => {
      // Send to backend
      socketManager.sendMessage(ActionType.Equip, {
        item_entity_id: item.entity_id,
      });
      // Optimistic update
      this.equippedItems[slot] = item;
      this.inventoryItems = this.inventoryItems.filter(
        (i) => i.entity_id !== item.entity_id,
      );
    };
    this.equipmentPanel.onUnequip = (item, slot) => {
      // Send to backend
      socketManager.sendMessage(ActionType.Unequip, {
        item_entity_id: item.entity_id,
      });
      // Optimistic update
      this.equippedItems[slot] = null;
      this.inventoryItems.push(item);
    };

    // 創建室內遮罩（用於遮住建築外面）
    this.indoorMask = this.add.graphics();
    this.indoorMask.setDepth(500);
    this.indoorMask.setVisible(false);

    // 顯示座標 UI
    this.createUI();

    // 放開移動鍵時停止
    this.input.keyboard?.on("keyup", (event: KeyboardEvent) => {
      const movementKeys = [
        "KeyW",
        "KeyA",
        "KeyS",
        "KeyD",
        "ArrowUp",
        "ArrowDown",
        "ArrowLeft",
        "ArrowRight",
      ];

      if (movementKeys.includes(event.code)) {
        const anyMovementKeyDown =
          this.wasd.up.isDown ||
          this.wasd.down.isDown ||
          this.wasd.left.isDown ||
          this.wasd.right.isDown;

        if (!anyMovementKeyDown) {
          socketManager.sendMessage("move", { vx: 0, vy: 0 });
        }
      }
    });
  }

  private connectToServer(): void {
    // Connect if not already connected
    if (!socketManager.isConnected()) {
      socketManager.connect("ws://localhost:5555/game/ws");
      GameStateLogger.logConnectionStatus(
        "Connecting to game server...",
        "#ffcc00",
      );
    } else {
      GameStateLogger.logConnectionStatus(
        "Already connected to server",
        "#4ecca3",
      );
    }

    // Subscribe to connection status changes
    socketManager.onConnectionStatusChange((status) => {
      switch (status) {
        case "connected":
          GameStateLogger.logConnectionStatus(
            "Connected successfully!",
            "#4ecca3",
          );
          break;
        case "connecting":
          break;
        case "disconnected":
          GameStateLogger.logConnectionStatus(
            "Disconnected from server",
            "#ff4444",
          );
          break;
        case "error":
          GameStateLogger.logError("WebSocket connection error");
          break;
      }
    });

    // Subscribe to game state updates
    this.gameStateUnsubscribe = socketManager.onGameStateUpdate(
      (state: ClientGameState) => {
        this.handleGameStateUpdate(state);
      },
    );

    // Listen for exit door unlocked message
    socketManager.on("exit_door_unlocked", (payload: { message: string }) => {
      console.log("Exit door unlocked!", payload);
      this.showNotification(payload.message, "#4ecca3");
    });

    // Listen for interact responses (success/error messages)
    socketManager.on(
      "interact",
      (payload: { success: boolean; message: string }) => {
        console.log("Interact response:", payload);
        if (payload.message) {
          const color = payload.success ? "#4ecca3" : "#ff4444";
          this.showNotification(payload.message, color);
        }
      },
    );

    // Listen for end_game — show final position overlay and lock interaction
    socketManager.on(
      "end_game",
      (payload: { player_id: string; position: number; result: string }) => {
        console.log("Game ended, final position:", payload);
        this.showGameEndOverlay(payload.position, payload.result);
      },
    );

    // Reset the logger for new session
    GameStateLogger.reset();
  }

  private handleGameStateUpdate(state: ClientGameState): void {
    this.lastGameState = state;

    // Update current player position from server
    if (state.current_player) {
      const pos = state.current_player.position;

      // 第一次收到位置，建立玩家
      if (!this.player) {
        this.createPlayer(pos.x, pos.y, state.current_player.username);
      }

      // 設定目標位置，在 update() 中平滑移動
      this.targetPosition = { x: pos.x, y: pos.y };

      // 同步玩家背包
      if (state.current_player.inventory) {
        this.syncInventory(state.current_player.inventory);
      }
      // 同步玩家裝備
      if (state.current_player.equipment) {
        this.syncEquipment(state.current_player.equipment);
      }
    } else {
      // current_player is null — player has escaped
      if (this.player && this.player.visible) {
        this.playEscapeParticles(this.player.x, this.player.y);
        this.player.setVisible(false);
        this.playerLegs?.setVisible(false);
        this.playerNameText?.setVisible(false);
      }
    }

    // Update other players on screen
    this.updateOtherPlayers(state.other_players || []);

    // Update walls from server
    this.updateWalls(state.walls || []);

    // Update doors from server
    this.updateDoors(state.doors || []);

    // Update containers from server
    this.updateContainers(state.containers || []);

    // Update escape doors from server
    this.updateEscapeDoors(state.escape_doors || []);

    // Update switches from server
    this.updateSwitches(state.switches || []);

    // 檢測狀態變化並顯示通知（避免重複）
    this.checkEscapeDoorStateChanges(state);
    this.checkPlayerEscapedState(state);

    // Update escaped count HUD
    if (this.escapedCountText) {
      const count = state.escaped_count ?? 0;
      this.escapedCountText.setText(`Escaped: ${count}`);
    }
  }

  private updateOtherPlayers(
    otherPlayersData: Array<{
      id: string;
      username: string;
      position: { x: number; y: number };
    }>,
  ): void {
    // Track which players are still in the game
    const activePlayerIds = new Set(otherPlayersData.map((p) => p.id));

    // Remove players who left
    this.otherPlayers.forEach((sprite, playerId) => {
      if (!activePlayerIds.has(playerId)) {
        this.playEscapeParticles(sprite.x, sprite.y);
        sprite.destroy();
        this.otherPlayers.delete(playerId);
        this.otherPlayersTargets.delete(playerId);

        // remove legs too
        const legs = this.otherPlayersLegs.get(playerId);
        if (legs) {
          legs.destroy();
          this.otherPlayersLegs.delete(playerId);
        }
        this.otherPlayersFacing.delete(playerId);
        this.otherPlayersWalkPhase.delete(playerId);

        // remove name text
        const nameText = this.otherPlayersNameTexts.get(playerId);
        if (nameText) {
          nameText.destroy();
          this.otherPlayersNameTexts.delete(playerId);
        }
        if (this.hoveredPlayerId === playerId) {
          this.hoveredPlayerId = undefined;
        }
      }
    });

    // Update or create other players
    otherPlayersData.forEach((playerData) => {
      let sprite = this.otherPlayers.get(playerData.id);

      if (!sprite) {
        // Create new sprite for this player
        sprite = this.physics.add.sprite(
          playerData.position.x,
          playerData.position.y,
          "otherPlayerDown",
        );
        sprite.setDepth(99);

        // set circular physics body, offset for 60x60 texture
        sprite.body.setCircle(20, 10, 10);

        // 點擊攻擊
        sprite.setInteractive();
        sprite.on("pointerdown", () => {
          if (!this.canAttack || !this.player) return;
          const distance = Phaser.Math.Distance.Between(
            this.player.x,
            this.player.y,
            sprite!.x,
            sprite!.y,
          );
          if (distance > 60) return;
          const entityId = this.otherPlayersEntityIds.get(playerData.id);
          if (entityId) {
            socketManager.sendMessage(ActionType.Attack, {
              enemy_entity_id: entityId,
            });
            this.playAttackEffect(sprite!);
            this.canAttack = false;
            this.time.delayedCall(500, () => {
              this.canAttack = true;
            });
          }
        });

        this.otherPlayers.set(playerData.id, sprite);
        this.otherPlayersEntityIds.set(playerData.id, playerData.entity_id);

        // create legs for this other player
        const legs = this.add.graphics();
        legs.setDepth(100);
        this.otherPlayersLegs.set(playerData.id, legs);
        this.otherPlayersFacing.set(playerData.id, "down");
        this.otherPlayersWalkPhase.set(playerData.id, 0);
        this.drawLegs(
          legs,
          playerData.position.x,
          playerData.position.y,
          "down",
          0,
          false,
          0x667788,
        );

        // create name text (hidden until hover)
        const nameText = this.add.text(
          playerData.position.x,
          playerData.position.y - 35,
          playerData.username || "Unknown",
          {
            fontSize: "11px",
            fontFamily: "Orbitron, monospace",
            color: "#ff00aa",
            stroke: "#0a0a12",
            strokeThickness: 3,
            align: "center",
          },
        );
        nameText.setOrigin(0.5, 1);
        nameText.setDepth(102);
        nameText.setVisible(false);
        this.otherPlayersNameTexts.set(playerData.id, nameText);

        // hover to show name + crosshair cursor
        const pid = playerData.id;
        sprite.on("pointerover", () => {
          this.hoveredPlayerId = pid;
          this.input.setDefaultCursor(this.crosshairCursorCSS);
        });
        sprite.on("pointerout", () => {
          if (this.hoveredPlayerId === pid) {
            this.hoveredPlayerId = undefined;
          }
          this.input.setDefaultCursor(this.defaultCursorCSS);
        });
      }

      // 設定目標位置，在 update() 中平滑移動
      this.otherPlayersTargets.set(playerData.id, {
        x: playerData.position.x,
        y: playerData.position.y,
      });
    });
  }

  private createMapBackground(): void {
    const graphics = this.add.graphics();
    const outerMargin = 200;

    // === deep space beyond the ship hull ===
    const spaceMargin = 150;
    const spaceOuter = outerMargin + spaceMargin; // camera limit

    // dark space backdrop — only the ring beyond the hull
    const spaceBg = this.add.graphics();
    spaceBg.fillStyle(0x050510, 1);
    // fill the full camera area, then the hull area will be drawn on top at depth -1
    spaceBg.fillRect(
      -spaceOuter,
      -spaceOuter,
      this.mapWidth + spaceOuter * 2,
      this.mapHeight + spaceOuter * 2,
    );
    spaceBg.setDepth(-3);

    // scatter stars only in the space region beyond the hull
    for (let i = 0; i < 150; i++) {
      const star = this.add.graphics();
      const size = Phaser.Math.FloatBetween(0.4, 2);
      const color = i < 90 ? 0xffffff : i < 120 ? 0xaaddff : 0xffccaa;
      star.fillStyle(color, Phaser.Math.FloatBetween(0.4, 1));
      star.fillCircle(0, 0, size);
      star.setPosition(
        Phaser.Math.Between(-spaceOuter, this.mapWidth + spaceOuter),
        Phaser.Math.Between(-spaceOuter, this.mapHeight + spaceOuter),
      );
      star.setScrollFactor(Phaser.Math.FloatBetween(0.3, 0.5));
      star.setDepth(-2);

      if (i % 4 === 0) {
        this.tweens.add({
          targets: star,
          alpha: 0.1,
          duration: Phaser.Math.Between(1000, 3000),
          ease: "Sine.easeInOut",
          yoyo: true,
          repeat: -1,
          delay: Phaser.Math.Between(0, 2000),
        });
      }
    }

    // === outer hull structure (fills entire outer area) ===
    const hw2 = this.mapWidth;
    const hh2 = this.mapHeight;

    // hull plating with metal texture
    // top
    const hullTop = this.add.tileSprite(
      -outerMargin,
      -outerMargin,
      hw2 + outerMargin * 2,
      outerMargin,
      "hullMetal",
    );
    hullTop.setOrigin(0, 0);
    hullTop.setDepth(-1);
    // bottom
    const hullBottom = this.add.tileSprite(
      -outerMargin,
      hh2,
      hw2 + outerMargin * 2,
      outerMargin,
      "hullMetal",
    );
    hullBottom.setOrigin(0, 0);
    hullBottom.setDepth(-1);
    // left
    const hullLeft = this.add.tileSprite(
      -outerMargin,
      0,
      outerMargin,
      hh2,
      "hullMetal",
    );
    hullLeft.setOrigin(0, 0);
    hullLeft.setDepth(-1);
    // right
    const hullRight = this.add.tileSprite(
      hw2,
      0,
      outerMargin,
      hh2,
      "hullMetal",
    );
    hullRight.setOrigin(0, 0);
    hullRight.setDepth(-1);
    // corners
    const hullTopLeft = this.add.tileSprite(
      -outerMargin,
      -outerMargin,
      outerMargin,
      outerMargin,
      "hullMetal",
    );
    hullTopLeft.setOrigin(0, 0);
    hullTopLeft.setDepth(-1);
    const hullTopRight = this.add.tileSprite(
      hw2,
      -outerMargin,
      outerMargin,
      outerMargin,
      "hullMetal",
    );
    hullTopRight.setOrigin(0, 0);
    hullTopRight.setDepth(-1);
    const hullBottomLeft = this.add.tileSprite(
      -outerMargin,
      hh2,
      outerMargin,
      outerMargin,
      "hullMetal",
    );
    hullBottomLeft.setOrigin(0, 0);
    hullBottomLeft.setDepth(-1);
    const hullBottomRight = this.add.tileSprite(
      hw2,
      hh2,
      outerMargin,
      outerMargin,
      "hullMetal",
    );
    hullBottomRight.setOrigin(0, 0);
    hullBottomRight.setDepth(-1);

    // === viewports (windows to see space) ===
    const viewportGraphics = this.add.graphics();

    const viewports = [
      // top windows
      { x: 120, y: -outerMargin + 20, w: 140, h: 80 },
      { x: 450, y: -outerMargin + 15, w: 160, h: 90 },
      { x: 800, y: -outerMargin + 25, w: 130, h: 75 },
      // bottom windows
      { x: 170, y: hh2 + outerMargin - 100, w: 150, h: 80 },
      { x: 550, y: hh2 + outerMargin - 95, w: 140, h: 80 },
      { x: 900, y: hh2 + outerMargin - 105, w: 120, h: 75 },
      // left windows
      { x: -outerMargin + 20, y: 120, w: 80, h: 120 },
      { x: -outerMargin + 15, y: 420, w: 85, h: 130 },
      // right windows
      { x: hw2 + outerMargin - 100, y: 170, w: 80, h: 120 },
      { x: hw2 + outerMargin - 105, y: 500, w: 85, h: 125 },
    ];

    viewports.forEach((vp) => {
      // space visible through viewport
      viewportGraphics.fillStyle(0x050510, 1);
      viewportGraphics.fillRoundedRect(vp.x, vp.y, vp.w, vp.h, 6);
      // window frame
      viewportGraphics.lineStyle(3, 0x3a4556, 1);
      viewportGraphics.strokeRoundedRect(vp.x, vp.y, vp.w, vp.h, 6);
      viewportGraphics.lineStyle(1, 0x4a5568, 1);
      viewportGraphics.strokeRoundedRect(
        vp.x + 3,
        vp.y + 3,
        vp.w - 6,
        vp.h - 6,
        4,
      );
    });

    // parallax stars in viewports
    viewports.forEach((vp) => {
      for (let i = 0; i < 8; i++) {
        const star = this.add.graphics();
        const size = Phaser.Math.FloatBetween(0.5, 2);
        const color = i < 5 ? 0xffffff : 0xaaddff;
        star.fillStyle(color, Phaser.Math.FloatBetween(0.6, 1));
        star.fillCircle(0, 0, size);
        const sx = Phaser.Math.Between(vp.x + 10, vp.x + vp.w - 10);
        const sy = Phaser.Math.Between(vp.y + 10, vp.y + vp.h - 10);
        star.setPosition(sx, sy);
        star.setScrollFactor(Phaser.Math.FloatBetween(0.85, 0.95));
        star.setDepth(0);

        if (i < 3) {
          this.tweens.add({
            targets: star,
            alpha: 0.1,
            duration: Phaser.Math.Between(800, 2000),
            ease: "Sine.easeInOut",
            yoyo: true,
            repeat: -1,
            delay: Phaser.Math.Between(0, 1500),
          });
        }
      }
    });

    viewportGraphics.setDepth(0);

    // === spaceship hull exterior ===
    const hullGraphics = this.add.graphics();
    const hw = this.mapWidth;
    const hh = this.mapHeight;
    const hullPad = 8;

    // outer hull shell - thick border around the ship
    hullGraphics.lineStyle(10, 0x2a3040, 1);
    hullGraphics.strokeRoundedRect(
      -hullPad,
      -hullPad,
      hw + hullPad * 2,
      hh + hullPad * 2,
      12,
    );
    hullGraphics.lineStyle(3, 0x4a5568, 1);
    hullGraphics.strokeRoundedRect(
      -hullPad - 5,
      -hullPad - 5,
      hw + hullPad * 2 + 10,
      hh + hullPad * 2 + 10,
      16,
    );
    hullGraphics.lineStyle(1, 0x6b7280, 1);
    hullGraphics.strokeRoundedRect(
      -hullPad - 8,
      -hullPad - 8,
      hw + hullPad * 2 + 16,
      hh + hullPad * 2 + 16,
      18,
    );

    // ventilation grilles (top)
    const ventGraphics = this.add.graphics();
    const ventPositions = [
      { x: 150, y: -60, w: 80, h: 35, horizontal: true },
      { x: 450, y: -55, w: 60, h: 30, horizontal: true },
      { x: 800, y: -65, w: 70, h: 35, horizontal: true },
      // bottom
      { x: 250, y: hh + 25, w: 80, h: 35, horizontal: true },
      { x: 650, y: hh + 30, w: 60, h: 30, horizontal: true },
      // left
      { x: -70, y: 200, w: 35, h: 60, horizontal: false },
      { x: -60, y: 500, w: 30, h: 70, horizontal: false },
      // right (away from engines)
      { x: hw + 25, y: 100, w: 35, h: 50, horizontal: false },
    ];
    ventPositions.forEach((v) => {
      // vent frame
      ventGraphics.fillStyle(0x1a2030, 1);
      ventGraphics.fillRect(v.x, v.y, v.w, v.h);
      ventGraphics.lineStyle(1, 0x3a4556, 1);
      ventGraphics.strokeRect(v.x, v.y, v.w, v.h);
      // grille slats
      ventGraphics.lineStyle(1, 0x2a3545, 1);
      if (v.horizontal) {
        for (let ly = v.y + 5; ly < v.y + v.h - 2; ly += 5) {
          ventGraphics.lineBetween(v.x + 3, ly, v.x + v.w - 3, ly);
        }
      } else {
        for (let lx = v.x + 5; lx < v.x + v.w - 2; lx += 5) {
          ventGraphics.lineBetween(lx, v.y + 3, lx, v.y + v.h - 3);
        }
      }
    });
    ventGraphics.setDepth(0);

    // pipes / conduits along hull
    const pipeGraphics = this.add.graphics();
    const pipeMax = outerMargin - 40;
    // top pipes
    pipeGraphics.lineStyle(4, 0x3a4556, 1);
    pipeGraphics.lineBetween(40, -25, hw - 40, -25);
    pipeGraphics.lineStyle(2, 0x4a5568, 1);
    pipeGraphics.lineBetween(40, -30, hw - 40, -30);
    // bottom pipes
    pipeGraphics.lineStyle(4, 0x3a4556, 1);
    pipeGraphics.lineBetween(40, hh + 25, hw - 40, hh + 25);
    pipeGraphics.lineStyle(2, 0x4a5568, 1);
    pipeGraphics.lineBetween(40, hh + 30, hw - 40, hh + 30);
    // left pipes
    pipeGraphics.lineStyle(4, 0x3a4556, 1);
    pipeGraphics.lineBetween(-25, 40, -25, hh - 40);
    pipeGraphics.lineStyle(2, 0x4a5568, 1);
    pipeGraphics.lineBetween(-30, 40, -30, hh - 40);
    // right pipes
    pipeGraphics.lineStyle(4, 0x3a4556, 1);
    pipeGraphics.lineBetween(hw + 25, 40, hw + 25, hh - 40);
    pipeGraphics.lineStyle(2, 0x4a5568, 1);
    pipeGraphics.lineBetween(hw + 30, 40, hw + 30, hh - 40);
    pipeGraphics.setDepth(0);

    // engines (right side - 3 engines)
    const engineGraphics = this.add.graphics();
    const engineX = hw + outerMargin - 20;
    const enginePositions = [hh * 0.2, hh * 0.5, hh * 0.8];

    enginePositions.forEach((ey) => {
      // engine housing
      engineGraphics.fillStyle(0x1e2530, 1);
      engineGraphics.fillRoundedRect(hw + 10, ey - 30, outerMargin - 25, 60, 6);
      engineGraphics.lineStyle(2, 0x4a5568, 1);
      engineGraphics.strokeRoundedRect(
        hw + 10,
        ey - 30,
        outerMargin - 25,
        60,
        6,
      );
      // inner detail
      engineGraphics.fillStyle(0x2a3040, 1);
      engineGraphics.fillRoundedRect(hw + 20, ey - 20, outerMargin - 45, 40, 4);
      engineGraphics.lineStyle(1, 0x5a6577, 1);
      engineGraphics.strokeRoundedRect(
        hw + 20,
        ey - 20,
        outerMargin - 45,
        40,
        4,
      );
      // exhaust glow layers
      engineGraphics.fillStyle(0x0066cc, 1);
      engineGraphics.fillCircle(engineX, ey, 45);
      engineGraphics.fillStyle(0x00aaff, 1);
      engineGraphics.fillCircle(engineX, ey, 28);
      engineGraphics.fillStyle(0xffaa44, 1);
      engineGraphics.fillCircle(engineX, ey, 15);
      engineGraphics.fillStyle(0xccffff, 1);
      engineGraphics.fillCircle(engineX, ey, 6);
    });
    engineGraphics.setDepth(0);

    // engine glow pulse
    this.tweens.add({
      targets: engineGraphics,
      alpha: 0.5,
      duration: 1500,
      ease: "Sine.easeInOut",
      yoyo: true,
      repeat: -1,
    });

    // corner structural beams
    const beamGraphics = this.add.graphics();
    // top-left
    beamGraphics.lineStyle(5, 0x3a4556, 1);
    beamGraphics.lineBetween(-outerMargin + 10, -outerMargin + 10, -5, -5);
    beamGraphics.lineStyle(3, 0x4a5568, 1);
    beamGraphics.lineBetween(-outerMargin + 15, -outerMargin + 5, 0, -10);
    // top-right
    beamGraphics.lineStyle(5, 0x3a4556, 1);
    beamGraphics.lineBetween(
      hw + outerMargin - 10,
      -outerMargin + 10,
      hw + 5,
      -5,
    );
    beamGraphics.lineStyle(3, 0x4a5568, 1);
    beamGraphics.lineBetween(hw + outerMargin - 15, -outerMargin + 5, hw, -10);
    // bottom-left
    beamGraphics.lineStyle(5, 0x3a4556, 1);
    beamGraphics.lineBetween(
      -outerMargin + 10,
      hh + outerMargin - 10,
      -5,
      hh + 5,
    );
    beamGraphics.lineStyle(3, 0x4a5568, 1);
    beamGraphics.lineBetween(
      -outerMargin + 15,
      hh + outerMargin - 5,
      0,
      hh + 10,
    );
    // bottom-right
    beamGraphics.lineStyle(5, 0x3a4556, 1);
    beamGraphics.lineBetween(
      hw + outerMargin - 10,
      hh + outerMargin - 10,
      hw + 5,
      hh + 5,
    );
    beamGraphics.lineStyle(3, 0x4a5568, 1);
    beamGraphics.lineBetween(
      hw + outerMargin - 15,
      hh + outerMargin - 5,
      hw,
      hh + 10,
    );
    beamGraphics.setDepth(0);

    // hull warning stripes at corners
    const stripeGraphics = this.add.graphics();
    const corners = [
      { x: -outerMargin + 15, y: -outerMargin + 15 },
      { x: hw + outerMargin - 45, y: -outerMargin + 15 },
      { x: -outerMargin + 15, y: hh + outerMargin - 45 },
      { x: hw + outerMargin - 45, y: hh + outerMargin - 45 },
    ];
    corners.forEach((c) => {
      for (let i = 0; i < 3; i++) {
        stripeGraphics.fillStyle(0xddaa00, 1);
        stripeGraphics.fillRect(c.x + i * 10, c.y, 5, 30);
      }
    });
    stripeGraphics.setDepth(0);

    hullGraphics.setDepth(0);

    // spaceship floor - tiled metal texture
    const floorTile = this.add.tileSprite(
      0,
      0,
      this.mapWidth,
      this.mapHeight,
      "metalFloor",
    );
    floorTile.setOrigin(0, 0);
    floorTile.setDepth(-1);

    // viewport windows - see space outside
    const windowPositions = [
      { x: 100, y: 0, w: 120, h: 8 },
      { x: 350, y: 0, w: 120, h: 8 },
      { x: 600, y: 0, w: 120, h: 8 },
      { x: 850, y: 0, w: 120, h: 8 },
      { x: 100, y: this.mapHeight - 8, w: 120, h: 8 },
      { x: 350, y: this.mapHeight - 8, w: 120, h: 8 },
      { x: 600, y: this.mapHeight - 8, w: 120, h: 8 },
      { x: 850, y: this.mapHeight - 8, w: 120, h: 8 },
    ];

    const windowGraphics = this.add.graphics();
    windowPositions.forEach((win) => {
      // space visible through window
      windowGraphics.fillStyle(0x0a0a1a, 1);
      windowGraphics.fillRect(win.x, win.y, win.w, win.h);
      // window frame
      windowGraphics.lineStyle(2, 0x5a6577, 0.8);
      windowGraphics.strokeRect(win.x, win.y, win.w, win.h);
      // stars through window
      for (let i = 0; i < 5; i++) {
        const sx = Phaser.Math.Between(win.x + 5, win.x + win.w - 5);
        const sy = Phaser.Math.Between(win.y + 2, win.y + win.h - 2);
        windowGraphics.fillStyle(0xffffff, Phaser.Math.FloatBetween(0.4, 1));
        windowGraphics.fillCircle(sx, sy, 1);
      }
    });
    windowGraphics.setDepth(-1);

    // ambient hull lights along edges
    const lightGraphics = this.add.graphics();
    for (let x = 40; x < this.mapWidth; x += 200) {
      // top edge lights
      lightGraphics.fillStyle(0xffaa44, 0.15);
      lightGraphics.fillCircle(x, 15, 30);
      lightGraphics.fillStyle(0xffaa44, 0.4);
      lightGraphics.fillCircle(x, 15, 3);
      // bottom edge lights
      lightGraphics.fillStyle(0xffaa44, 0.15);
      lightGraphics.fillCircle(x, this.mapHeight - 15, 30);
      lightGraphics.fillStyle(0xffaa44, 0.4);
      lightGraphics.fillCircle(x, this.mapHeight - 15, 3);
    }
    lightGraphics.setDepth(-1);

    // pulsing light animation
    this.tweens.add({
      targets: lightGraphics,
      alpha: 0.5,
      duration: 2000,
      ease: "Sine.easeInOut",
      yoyo: true,
      repeat: -1,
    });

    // hull boundary - industrial metal frame
    graphics.lineStyle(6, 0x3a3428, 1);
    graphics.strokeRect(0, 0, this.mapWidth, this.mapHeight);
    graphics.lineStyle(2, 0x554a38, 1);
    graphics.strokeRect(3, 3, this.mapWidth - 6, this.mapHeight - 6);
    // inner warn trim
    graphics.lineStyle(1, 0xffaa44, 0.15);
    graphics.strokeRect(6, 6, this.mapWidth - 12, this.mapHeight - 12);

    graphics.setDepth(-1);

    // save as outdoor objects
    this.outsideObjects.push(graphics);
    this.outsideObjects.push(floorTile);
    this.outsideObjects.push(windowGraphics);
    this.outsideObjects.push(lightGraphics);
  }

  private createBuildings(): void {
    // 建築配置
    const buildingConfigs = [
      { x: 200, y: 200, width: 200, height: 150, doorSide: "bottom" as const },
    ];

    buildingConfigs.forEach((config, index) => {
      const building = this.createBuilding(
        `building_${index}`,
        config.x,
        config.y,
        config.width,
        config.height,
        config.doorSide,
      );
      this.buildings.push(building);
    });
  }

  private createBuilding(
    id: string,
    x: number,
    y: number,
    width: number,
    height: number,
    doorSide: "top" | "bottom" | "left" | "right",
  ): Building {
    const wallThickness = 12;
    const doorWidth = 50;

    // 艙室地板 - 金屬格紋
    const floor = this.add.graphics();
    floor.fillStyle(0x2a3040, 1);
    floor.fillRect(x, y, width, height);
    floor.lineStyle(1, 0x3d4556, 0.4);
    for (let tx = x; tx < x + width; tx += 40) {
      floor.lineBetween(tx, y, tx, y + height);
    }
    for (let ty = y; ty < y + height; ty += 40) {
      floor.lineBetween(x, ty, x + width, ty);
    }
    floor.setDepth(1);

    // 牆壁群組
    const wallGroup = this.physics.add.staticGroup();
    const wallGraphics = this.add.graphics();
    wallGraphics.setDepth(50);

    const createWall = (wx: number, wy: number, ww: number, wh: number) => {
      // 艙壁 - 金屬質感
      wallGraphics.fillStyle(0x4a5568, 1);
      wallGraphics.fillRect(wx, wy, ww, wh);
      wallGraphics.lineStyle(1, 0x6b7280, 0.6);
      wallGraphics.strokeRect(wx, wy, ww, wh);

      // 碰撞牆壁
      const wallSprite = this.physics.add.staticSprite(
        wx + ww / 2,
        wy + wh / 2,
        undefined as unknown as string,
      );
      wallSprite.body?.setSize(ww, wh);
      wallSprite.setVisible(false);
      wallGroup.add(wallSprite);
    };

    // 上牆
    if (doorSide !== "top") {
      createWall(x, y, width, wallThickness);
    } else {
      const sideWidth = (width - doorWidth) / 2;
      createWall(x, y, sideWidth, wallThickness);
      createWall(x + sideWidth + doorWidth, y, sideWidth, wallThickness);
    }

    // 下牆
    if (doorSide !== "bottom") {
      createWall(x, y + height - wallThickness, width, wallThickness);
    } else {
      const sideWidth = (width - doorWidth) / 2;
      createWall(x, y + height - wallThickness, sideWidth, wallThickness);
      createWall(
        x + sideWidth + doorWidth,
        y + height - wallThickness,
        sideWidth,
        wallThickness,
      );
    }

    // 左牆
    if (doorSide !== "left") {
      createWall(x, y, wallThickness, height);
    } else {
      const sideHeight = (height - doorWidth) / 2;
      createWall(x, y, wallThickness, sideHeight);
      createWall(x, y + sideHeight + doorWidth, wallThickness, sideHeight);
    }

    // 右牆
    if (doorSide !== "right") {
      createWall(x + width - wallThickness, y, wallThickness, height);
    } else {
      const sideHeight = (height - doorWidth) / 2;
      createWall(x + width - wallThickness, y, wallThickness, sideHeight);
      createWall(
        x + width - wallThickness,
        y + sideHeight + doorWidth,
        wallThickness,
        sideHeight,
      );
    }

    // 艙頂（遮蓋建築內部）
    const roof = this.add.graphics();
    roof.fillStyle(0x2d3748, 0.97);
    roof.fillRect(x - 5, y - 5, width + 10, height + 10);
    roof.lineStyle(2, 0x4a5568, 1);
    roof.strokeRect(x - 5, y - 5, width + 10, height + 10);
    roof.setDepth(200);

    // 入口標示（在屋頂上方，標示門的位置）
    const doorMarker = this.add.graphics();
    doorMarker.setDepth(250); // 高於屋頂(200)

    let doorX = 0;
    let doorY = 0;
    const arrowSize = 10;

    // 計算門在屋頂上的位置
    if (doorSide === "top") {
      doorX = x + width / 2;
      doorY = y - 5; // 屋頂邊緣
    } else if (doorSide === "bottom") {
      doorX = x + width / 2;
      doorY = y + height + 5;
    } else if (doorSide === "left") {
      doorX = x - 5;
      doorY = y + height / 2;
    } else {
      doorX = x + width + 5;
      doorY = y + height / 2;
    }

    // 畫入口標示（青色箭頭指向門口）
    doorMarker.fillStyle(0xffaa44, 1);

    // 根據門的方向畫箭頭（從外面指向建築內部）
    if (doorSide === "top") {
      // 門在上方，箭頭指向下（進入建築）
      doorMarker.fillTriangle(
        doorX,
        doorY + arrowSize,
        doorX - arrowSize,
        doorY - arrowSize,
        doorX + arrowSize,
        doorY - arrowSize,
      );
    } else if (doorSide === "bottom") {
      // 門在下方，箭頭指向上（進入建築）
      doorMarker.fillTriangle(
        doorX,
        doorY - arrowSize,
        doorX - arrowSize,
        doorY + arrowSize,
        doorX + arrowSize,
        doorY + arrowSize,
      );
    } else if (doorSide === "left") {
      // 門在左方，箭頭指向右（進入建築）
      doorMarker.fillTriangle(
        doorX + arrowSize,
        doorY,
        doorX - arrowSize,
        doorY - arrowSize,
        doorX - arrowSize,
        doorY + arrowSize,
      );
    } else {
      // 門在右方，箭頭指向左（進入建築）
      doorMarker.fillTriangle(
        doorX - arrowSize,
        doorY,
        doorX + arrowSize,
        doorY - arrowSize,
        doorX + arrowSize,
        doorY + arrowSize,
      );
    }

    // 入口圓圈
    doorMarker.lineStyle(3, 0xffaa44, 0.8);
    doorMarker.strokeCircle(doorX, doorY, 18);

    // 閃爍動畫
    this.tweens.add({
      targets: doorMarker,
      alpha: 0.4,
      duration: 800,
      yoyo: true,
      repeat: -1,
      ease: "Sine.easeInOut",
    });

    // 儲存牆壁圖形為室外物件
    this.outsideObjects.push(wallGraphics);
    this.outsideObjects.push(roof);

    // 創建門 (可開關的)
    const door = this.add.graphics();
    door.setDepth(51);

    // 計算門的位置和大小
    let doorRectX = 0;
    let doorRectY = 0;
    let doorRectW = 0;
    let doorRectH = 0;

    if (doorSide === "top") {
      doorRectX = x + (width - doorWidth) / 2;
      doorRectY = y;
      doorRectW = doorWidth;
      doorRectH = wallThickness;
    } else if (doorSide === "bottom") {
      doorRectX = x + (width - doorWidth) / 2;
      doorRectY = y + height - wallThickness;
      doorRectW = doorWidth;
      doorRectH = wallThickness;
    } else if (doorSide === "left") {
      doorRectX = x;
      doorRectY = y + (height - doorWidth) / 2;
      doorRectW = wallThickness;
      doorRectH = doorWidth;
    } else {
      doorRectX = x + width - wallThickness;
      doorRectY = y + (height - doorWidth) / 2;
      doorRectW = wallThickness;
      doorRectH = doorWidth;
    }

    // 畫門 (金屬艙門)
    door.fillStyle(0x5a6577, 1);
    door.fillRect(doorRectX, doorRectY, doorRectW, doorRectH);
    door.lineStyle(2, 0x6b7280, 1);
    door.strokeRect(doorRectX, doorRectY, doorRectW, doorRectH);

    // 創建門的碰撞體 (Rectangle)
    const doorCollider = this.add.rectangle(
      doorRectX + doorRectW / 2,
      doorRectY + doorRectH / 2,
      doorRectW,
      doorRectH,
    );
    this.physics.add.existing(doorCollider, true); // true = static body
    doorCollider.setVisible(false);

    return {
      id,
      x,
      y,
      width,
      height,
      doorSide,
      wallGroup,
      roof,
      floor,
      doorMarker,
      door,
      doorCollider,
      isOpen: false,
    };
  }

  private isPlayerInsideBuilding(building: Building): boolean {
    if (!this.player) return false;
    return (
      this.player.x >= building.x &&
      this.player.x <= building.x + building.width &&
      this.player.y >= building.y &&
      this.player.y <= building.y + building.height
    );
  }

  private checkBuildingStatus(): void {
    let insideBuilding: Building | null = null;

    for (const building of this.buildings) {
      if (this.isPlayerInsideBuilding(building)) {
        insideBuilding = building;
        break;
      }
    }

    // 狀態改變時更新視覺
    if (insideBuilding !== this.currentBuilding) {
      if (insideBuilding) {
        // 進入建築：隱藏室外物件，顯示當前建築內部
        this.enterBuilding(insideBuilding);
      } else {
        // 離開建築：顯示室外物件
        this.exitBuilding();
      }
      this.currentBuilding = insideBuilding;
    }
  }

  private enterBuilding(building: Building): void {
    // 隱藏當前建築屋頂和入口標示
    building.roof.setVisible(false);
    building.doorMarker.setVisible(false);

    // 隱藏所有入口標示
    this.buildings.forEach((b) => {
      b.doorMarker.setVisible(false);
    });

    // 顯示室內遮罩，遮住建築外面的一切
    this.indoorMask.setVisible(true);
    this.updateIndoorMask(building);
  }

  private exitBuilding(): void {
    // 顯示所有屋頂和入口標示
    this.buildings.forEach((b) => {
      b.roof.setVisible(true);
      b.doorMarker.setVisible(true);
    });

    // 隱藏室內遮罩
    this.indoorMask.setVisible(false);
  }

  private updateIndoorMask(building: Building): void {
    this.indoorMask.clear();

    // 用黑色填充整個地圖，但挖空建築內部區域
    const padding = 5;
    const bx = building.x - padding;
    const by = building.y - padding;
    const bw = building.width + padding * 2;
    const bh = building.height + padding * 2;

    this.indoorMask.fillStyle(0x000000, 1);

    // 上方區域
    this.indoorMask.fillRect(-1000, -1000, this.mapWidth + 2000, by + 1000);
    // 下方區域
    this.indoorMask.fillRect(
      -1000,
      by + bh,
      this.mapWidth + 2000,
      this.mapHeight + 1000,
    );
    // 左側區域
    this.indoorMask.fillRect(-1000, by, bx + 1000, bh);
    // 右側區域
    this.indoorMask.fillRect(bx + bw, by, this.mapWidth + 1000, bh);
  }

  private toggleDoor(building: Building): void {
    building.isOpen = !building.isOpen;

    if (building.isOpen) {
      // 開門：隱藏門並禁用碰撞
      building.door.setVisible(false);
      const body = building.doorCollider
        .body as Phaser.Physics.Arcade.StaticBody;
      if (body) {
        body.enable = false;
      }
    } else {
      // 關門：顯示門並啟用碰撞
      building.door.setVisible(true);
      const body = building.doorCollider
        .body as Phaser.Physics.Arcade.StaticBody;
      if (body) {
        body.enable = true;
      }
    }
  }

  private getNearbyBuilding(): Building | null {
    if (!this.player) return null;
    const interactDistance = 60;

    for (const building of this.buildings) {
      // 計算門的中心位置
      // const doorWidth = 50;
      let doorCenterX = 0;
      let doorCenterY = 0;

      if (building.doorSide === "top") {
        doorCenterX = building.x + building.width / 2;
        doorCenterY = building.y;
      } else if (building.doorSide === "bottom") {
        doorCenterX = building.x + building.width / 2;
        doorCenterY = building.y + building.height;
      } else if (building.doorSide === "left") {
        doorCenterX = building.x;
        doorCenterY = building.y + building.height / 2;
      } else {
        doorCenterX = building.x + building.width;
        doorCenterY = building.y + building.height / 2;
      }

      const distance = Phaser.Math.Distance.Between(
        this.player.x,
        this.player.y,
        doorCenterX,
        doorCenterY,
      );

      if (distance < interactDistance) {
        return building;
      }
    }

    return null;
  }

  private createUI(): void {
    const posText = this.add.text(10, 10, "", {
      fontSize: "14px",
      color: "#4ecca3",
      backgroundColor: "#16213e",
      padding: { x: 10, y: 5 },
    });
    posText.setScrollFactor(0);
    posText.setDepth(1000);

    // Escaped players count (top-right)
    this.escapedCountText = this.add.text(
      this.cameras.main.width - 10,
      10,
      "Escaped: 0",
      {
        fontSize: "14px",
        color: "#FFD700",
        backgroundColor: "#16213e",
        padding: { x: 10, y: 5 },
      },
    );
    this.escapedCountText.setOrigin(1, 0); // right-aligned
    this.escapedCountText.setScrollFactor(0);
    this.escapedCountText.setDepth(1000);

    // 每幀更新座標
    this.events.on("update", () => {
      if (!this.player) {
        posText.setText("Waiting for server...");
        return;
      }
      const status = this.currentBuilding ? `Indoor` : `Outdoor`;
      posText.setText(
        `X: ${Math.round(this.player.x)} Y: ${Math.round(this.player.y)} | ${status}`,
      );
    });
  }

  private showNotification(message: string, color: string): void {
    // Create notification text at top center of screen
    const notification = this.add.text(
      this.cameras.main.centerX,
      100,
      message,
      {
        fontSize: "20px",
        color: "#ffffff",
        backgroundColor: color,
        padding: { x: 20, y: 10 },
      },
    );
    notification.setOrigin(0.5);
    notification.setScrollFactor(0);
    notification.setDepth(2000);

    // Fade out and destroy after 3 seconds
    this.tweens.add({
      targets: notification,
      alpha: 0,
      duration: 2000,
      delay: 1000,
      onComplete: () => {
        notification.destroy();
      },
    });
  }

  /**
   * 檢測逃生門和開關的狀態變化，並顯示對應通知
   * 避免重複通知：只在狀態真正改變時才顯示
   */
  private checkEscapeDoorStateChanges(state: ClientGameState): void {
    // 檢查是否有逃生門資料
    const escapeDoor = state.escape_doors?.[0];
    if (!escapeDoor) return;

    // 檢查開關是否被激活 → 逃生門解鎖
    const switchData = state.switches?.[0];
    if (switchData) {
      // 開關剛被激活（從 false/null 變成 true）
      if (
        switchData.is_activated === true &&
        this.previousSwitchActivated !== true
      ) {
        this.showNotification("Exit door unlocked! Run to escape!", "#4ecca3");
      }
      this.previousSwitchActivated = switchData.is_activated;
    }

    // 檢查逃生門是否被打開
    if (escapeDoor.is_open === true && this.previousEscapeDoorOpened !== true) {
      this.showNotification("Escape door opened!", "#4ecca3");
    }
    this.previousEscapeDoorOpened = escapeDoor.is_open;

    // 儲存逃生門鎖定狀態（用於未來可能的需求）
    this.previousEscapeDoorLocked = escapeDoor.is_locked;
  }

  /**
   * 檢測玩家是否逃脫成功
   * 後端會設置 player.escape = true
   */
  private checkPlayerEscapedState(state: ClientGameState): void {
    // 檢查當前玩家
    if (state.current_player?.escape === true && state.current_player.id) {
      if (!this.escapedPlayers.has(state.current_player.id)) {
        this.showNotification(
          `${state.current_player.username} escaped successfully!`,
          "#FFD700", // 金色
        );
        this.escapedPlayers.add(state.current_player.id);
      }
    }

    // 檢查其他玩家
    state.other_players?.forEach((player) => {
      if (player.escape === true && player.id) {
        if (!this.escapedPlayers.has(player.id)) {
          this.showNotification(
            `${player.username} escaped successfully!`,
            "#FFD700",
          );
          this.escapedPlayers.add(player.id);
        }
      }
    });
  }

  destroy(): void {
    // Clean up subscriptions when scene is destroyed
    if (this.gameStateUnsubscribe) {
      this.gameStateUnsubscribe();
      GameStateLogger.logConnectionStatus("Scene shutting down", "#808080");
    }

    // 重置狀態追蹤
    this.previousEscapeDoorLocked = null;
    this.previousEscapeDoorOpened = null;
    this.previousSwitchActivated = null;
    this.escapedPlayers.clear();
  }

  update(): void {
    // skip all input/movement if player has escaped
    if (this.player && !this.player.visible) {
      // still update other players smoothly
      this.updateOtherPlayersSmooth();
      return;
    }

    // handle movement
    let vx = 0;
    let vy = 0;

    // calculate horizontal direction
    if (this.cursors.left.isDown || this.wasd.left.isDown) {
      vx = -1;
    } else if (this.cursors.right.isDown || this.wasd.right.isDown) {
      vx = 1;
    }

    // calculate vertical direction
    if (this.cursors.up.isDown || this.wasd.up.isDown) {
      vy = -1;
    } else if (this.cursors.down.isDown || this.wasd.down.isDown) {
      vy = 1;
    }

    // update player facing and legs
    if (this.player && this.playerLegs) {
      const isMoving = vx !== 0 || vy !== 0;
      if (isMoving) {
        // determine facing from dominant axis
        let newFacing: "up" | "down" | "left" | "right";
        if (Math.abs(vy) >= Math.abs(vx)) {
          newFacing = vy < 0 ? "up" : "down";
        } else {
          newFacing = vx < 0 ? "left" : "right";
        }
        if (newFacing !== this.playerFacing) {
          this.playerFacing = newFacing;
          this.player.setTexture(this.facingTextureKey("player", newFacing));
        }
        this.walkPhase += 0.3;
      }
      this.drawLegs(
        this.playerLegs,
        this.player.x,
        this.player.y,
        this.playerFacing,
        this.walkPhase,
        isMoving,
        0x667788,
      );
    }

    // update player name position
    if (this.player && this.playerNameText) {
      this.playerNameText.setPosition(this.player.x, this.player.y - 35);
    }

    // send websocket message for movement
    if (vx !== 0 || vy !== 0) {
      socketManager.sendMessage(ActionType.Move, {
        vx: vx,
        vy: vy,
      });
    }

    // 平滑移動到目標位置 (lerp)
    const lerpFactor = 0.3; // 0-1，越大越快到達目標

    if (this.player && this.targetPosition) {
      this.player.x = Phaser.Math.Linear(
        this.player.x,
        this.targetPosition.x,
        lerpFactor,
      );
      this.player.y = Phaser.Math.Linear(
        this.player.y,
        this.targetPosition.y,
        lerpFactor,
      );
    }

    this.updateOtherPlayersSmooth();

    // 檢查是否進入/離開建築
    this.checkBuildingStatus();

    // 檢查寶箱距離，太遠自動關閉（只有跳窗開啟時才檢查）
    if (this.isPopupOpen) {
      this.checkChestDistance();
    }
  }

  private updateOtherPlayersSmooth(): void {
    const lerpFactor = 0.3;
    this.otherPlayers.forEach((sprite, playerId) => {
      const target = this.otherPlayersTargets.get(playerId);
      if (target) {
        const prevX = sprite.x;
        const prevY = sprite.y;

        sprite.x = Phaser.Math.Linear(sprite.x, target.x, lerpFactor);
        sprite.y = Phaser.Math.Linear(sprite.y, target.y, lerpFactor);

        // update facing and legs for other players
        const legs = this.otherPlayersLegs.get(playerId);
        if (legs) {
          const deltaX = target.x - prevX;
          const deltaY = target.y - prevY;
          const length = Math.sqrt(deltaX * deltaX + deltaY * deltaY);
          const isMoving = length > 0.5;

          if (isMoving) {
            let newFacing: "up" | "down" | "left" | "right";
            if (Math.abs(deltaY) >= Math.abs(deltaX)) {
              newFacing = deltaY < 0 ? "up" : "down";
            } else {
              newFacing = deltaX < 0 ? "left" : "right";
            }
            const prevFacing = this.otherPlayersFacing.get(playerId) || "down";
            if (newFacing !== prevFacing) {
              this.otherPlayersFacing.set(playerId, newFacing);
              sprite.setTexture(
                this.facingTextureKey("otherPlayer", newFacing),
              );
            }
            const phase = (this.otherPlayersWalkPhase.get(playerId) || 0) + 0.3;
            this.otherPlayersWalkPhase.set(playerId, phase);
          }

          const facing = this.otherPlayersFacing.get(playerId) || "down";
          const phase = this.otherPlayersWalkPhase.get(playerId) || 0;
          this.drawLegs(
            legs,
            sprite.x,
            sprite.y,
            facing,
            phase,
            isMoving,
            0x667788,
          );
        }

        // update name text position and hover visibility
        const nameText = this.otherPlayersNameTexts.get(playerId);
        if (nameText) {
          nameText.setPosition(sprite.x, sprite.y - 35);
          nameText.setVisible(this.hoveredPlayerId === playerId);
        }
      }
    });
  }

  // private connectWebSocket(): void {
  //   this.socket = new WebSocket("ws://localhost:5555/game/ws");

  //   this.socket.onopen = () => {
  //     console.log("WebSocket connected");
  //     this.updateStatus("WebSocket Connected", "#4ecca3");
  //   };

  //   this.socket.onerror = (error) => {
  //     console.error("WebSocket error:", error);
  //     this.updateStatus("WebSocket Error", "#ff4444");
  //   };

  //   this.socket.onclose = () => {
  //     console.log("WebSocket disconnected");
  //     this.updateStatus("WebSocket Disconnected", "#ffcc00");
  //   };

  //   this.socket.onmessage = (event) => {
  //     try {
  //       const data = JSON.parse(event.data);
  //       console.log("Received server message:", data);
  //     } catch (e) {
  //       console.error("Failed to parse message:", e);
  //     }
  //   };
  // }
  // websocket send message
  // sendMessage<T extends keyof ActionMap>(
  //   action: T,
  //   payload: ActionMap[T],
  // ): void {
  //   if (this.socket && this.socket.readyState === WebSocket.OPEN) {
  //     const message: ClientMessage<T> = {
  //       action,
  //       payload,
  //       seq: ++this.seq,
  //     };
  //     this.socket.send(JSON.stringify(message));
  //   }
  // }
}
