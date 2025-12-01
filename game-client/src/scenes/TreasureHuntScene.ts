/**
 * TreasureHuntScene - 多人尋寶遊戲場景
 * 展示視野系統 (Fog of War) + 建築碰撞 + 室內視線遮擋
 */

import Phaser from "phaser";
import { GameClient } from "@/game/GameClient";
import type { Building, Enemy, VisibleObjects } from "@/game/MockBackend";

interface BuildingGraphicsData {
  id: string;
  wallSprites: Phaser.Physics.Arcade.Sprite[];
  graphics?: Phaser.GameObjects.Graphics;
  roof?: Phaser.GameObjects.Graphics;
  floor?: Phaser.GameObjects.Graphics;
  doorGraphics?: Phaser.GameObjects.Graphics;
  arrowGraphics?: Phaser.GameObjects.Graphics;
  typeLabel?: Phaser.GameObjects.Text;
}

export class TreasureHuntScene extends Phaser.Scene {
  private client!: GameClient;
  private player!: Phaser.Physics.Arcade.Sprite;
  private visibleObjects!: {
    treasures: Map<string, Phaser.GameObjects.Sprite>;
    enemies: Map<string, Phaser.GameObjects.Sprite>;
    buildings: Map<string, BuildingGraphicsData>;
    walls: Map<string, Phaser.Physics.Arcade.Sprite>;
    players: Map<string, Phaser.GameObjects.Sprite>;
  };
  private lastUpdatePos!: { x: number; y: number };
  private updateThreshold!: number;

  // Groups
  private treasureGroup!: Phaser.GameObjects.Group;
  private enemyGroup!: Phaser.GameObjects.Group;
  private playerGroup!: Phaser.GameObjects.Group;
  private wallGroup!: Phaser.Physics.Arcade.StaticGroup;

  // UI
  private uiContainer!: Phaser.GameObjects.Container;
  private hpText!: Phaser.GameObjects.Text;
  private scoreText!: Phaser.GameObjects.Text;
  private itemText!: Phaser.GameObjects.Text;
  private minimapContainer!: Phaser.GameObjects.Container;
  private minimapGraphics!: Phaser.GameObjects.Graphics;

  // Fog of War
  private fogLayer!: Phaser.GameObjects.Graphics;
  private isPlayerIndoor: boolean = false;
  private currentBuildingId: string | null = null;
  private indoorIndicator?: Phaser.GameObjects.Text;

  // Controls
  private cursors!: Phaser.Types.Input.Keyboard.CursorKeys;
  private wasd!: {
    up: Phaser.Input.Keyboard.Key;
    down: Phaser.Input.Keyboard.Key;
    left: Phaser.Input.Keyboard.Key;
    right: Phaser.Input.Keyboard.Key;
    attack: Phaser.Input.Keyboard.Key;
    collect: Phaser.Input.Keyboard.Key;
  };

  // Attack effect
  private attackEffect!: Phaser.GameObjects.Arc;

  // Status callback
  private onStatusChange?: (status: string, color: string) => void;

  constructor() {
    super({ key: "TreasureHuntScene" });
  }

  init(): void {
    // 在場景初始化時設置屬性
    this.client = new GameClient();
    this.visibleObjects = {
      treasures: new Map(),
      enemies: new Map(),
      buildings: new Map(),
      walls: new Map(),
      players: new Map(),
    };
    this.lastUpdatePos = { x: 0, y: 0 };
    this.updateThreshold = 10;
    this.isPlayerIndoor = false;
    this.currentBuildingId = null;
  }

  setStatusCallback(callback: (status: string, color: string) => void): void {
    this.onStatusChange = callback;
  }

  preload(): void {
    this.createGraphics();
  }

  private createGraphics(): void {
    // 玩家
    const playerGraphics = this.make.graphics({});
    playerGraphics.fillStyle(0x4ecca3, 1);
    playerGraphics.fillCircle(20, 20, 18);
    playerGraphics.fillStyle(0xffffff, 1);
    playerGraphics.fillCircle(14, 16, 4);
    playerGraphics.fillCircle(26, 16, 4);
    playerGraphics.generateTexture("treasurePlayer", 40, 40);
    playerGraphics.destroy();

    // 其他玩家
    const otherPlayerGraphics = this.make.graphics({});
    otherPlayerGraphics.fillStyle(0xe94560, 1);
    otherPlayerGraphics.fillCircle(20, 20, 18);
    otherPlayerGraphics.fillStyle(0xffffff, 1);
    otherPlayerGraphics.fillCircle(14, 16, 4);
    otherPlayerGraphics.fillCircle(26, 16, 4);
    otherPlayerGraphics.generateTexture("otherTreasurePlayer", 40, 40);
    otherPlayerGraphics.destroy();

    // 金寶箱
    const goldChestGraphics = this.make.graphics({});
    goldChestGraphics.fillStyle(0xffd700, 1);
    goldChestGraphics.fillRoundedRect(0, 8, 32, 22, 4);
    goldChestGraphics.fillStyle(0xb8860b, 1);
    goldChestGraphics.fillRoundedRect(0, 0, 32, 12, 4);
    goldChestGraphics.fillStyle(0xffffff, 1);
    goldChestGraphics.fillRect(14, 12, 4, 8);
    goldChestGraphics.generateTexture("goldChest", 32, 30);
    goldChestGraphics.destroy();

    // 銀寶箱
    const silverChestGraphics = this.make.graphics({});
    silverChestGraphics.fillStyle(0xc0c0c0, 1);
    silverChestGraphics.fillRoundedRect(0, 8, 28, 18, 4);
    silverChestGraphics.fillStyle(0x808080, 1);
    silverChestGraphics.fillRoundedRect(0, 0, 28, 10, 4);
    silverChestGraphics.fillStyle(0xffffff, 1);
    silverChestGraphics.fillRect(12, 10, 4, 6);
    silverChestGraphics.generateTexture("silverChest", 28, 26);
    silverChestGraphics.destroy();

    // 骷髏敵人
    const skeletonGraphics = this.make.graphics({});
    skeletonGraphics.fillStyle(0xeeeeee, 1);
    skeletonGraphics.fillCircle(16, 12, 12);
    skeletonGraphics.fillStyle(0x333333, 1);
    skeletonGraphics.fillCircle(12, 10, 3);
    skeletonGraphics.fillCircle(20, 10, 3);
    skeletonGraphics.fillRect(10, 16, 12, 2);
    skeletonGraphics.fillStyle(0xeeeeee, 1);
    skeletonGraphics.fillRect(12, 24, 8, 16);
    skeletonGraphics.generateTexture("skeleton", 32, 40);
    skeletonGraphics.destroy();

    // 哥布林敵人
    const goblinGraphics = this.make.graphics({});
    goblinGraphics.fillStyle(0x2d5a27, 1);
    goblinGraphics.fillCircle(16, 14, 14);
    goblinGraphics.fillStyle(0xff0000, 1);
    goblinGraphics.fillCircle(10, 12, 4);
    goblinGraphics.fillCircle(22, 12, 4);
    goblinGraphics.fillStyle(0x1a3518, 1);
    goblinGraphics.fillTriangle(8, 2, 0, 16, 10, 12);
    goblinGraphics.fillTriangle(24, 2, 32, 16, 22, 12);
    goblinGraphics.generateTexture("goblin", 32, 32);
    goblinGraphics.destroy();
  }

  async create(): Promise<void> {
    // 設置世界邊界
    this.physics.world.setBounds(0, 0, 3000, 3000);

    // 創建地圖背景
    this.createMapBackground();

    // 創建物件群組
    this.treasureGroup = this.add.group();
    this.enemyGroup = this.add.group();
    this.playerGroup = this.add.group();
    this.wallGroup = this.physics.add.staticGroup();

    // 創建玩家
    this.player = this.physics.add.sprite(1500, 1500, "treasurePlayer");
    this.player.setCollideWorldBounds(true);
    this.player.setDepth(100);
    this.player.body?.setSize(24, 24);

    // 玩家與牆壁碰撞
    this.physics.add.collider(this.player, this.wallGroup);

    // 攻擊效果
    this.attackEffect = this.add.circle(0, 0, 60, 0xff0000, 0.3);
    this.attackEffect.setVisible(false);
    this.attackEffect.setDepth(99);

    // 創建視野遮罩 (Fog of War)
    this.createFogOfWar();

    // 設置相機
    this.cameras.main.startFollow(this.player, true, 0.1, 0.1);
    this.cameras.main.setZoom(1);
    this.cameras.main.setBounds(0, 0, 3000, 3000);

    // 輸入控制
    this.cursors = this.input.keyboard!.createCursorKeys();
    this.wasd = {
      up: this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.W),
      down: this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.S),
      left: this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.A),
      right: this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.D),
      attack: this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.SPACE),
      collect: this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.E),
    };

    // ESC 返回主選單
    this.input.keyboard?.on("keydown-ESC", () => {
      this.scene.start("MainMenuScene");
    });

    // UI
    this.createUI();

    // 連線並加入遊戲
    await this.connectToServer();
  }

  private createMapBackground(): void {
    const graphics = this.add.graphics();
    graphics.lineStyle(1, 0x333355, 0.3);

    for (let x = 0; x <= 3000; x += 100) {
      graphics.lineBetween(x, 0, x, 3000);
    }
    for (let y = 0; y <= 3000; y += 100) {
      graphics.lineBetween(0, y, 3000, y);
    }

    // 添加草地裝飾
    for (let i = 0; i < 200; i++) {
      const x = Math.random() * 3000;
      const y = Math.random() * 3000;
      graphics.fillStyle(0x228b22, 0.3);
      graphics.fillCircle(x, y, 20 + Math.random() * 30);
    }

    graphics.setDepth(-1);
  }

  private createFogOfWar(): void {
    this.fogLayer = this.add.graphics();
    this.fogLayer.setDepth(1000);
    this.isPlayerIndoor = false;
    this.currentBuildingId = null;
  }

  private updateFogOfWar(): void {
    const cam = this.cameras.main;
    const fogColor = 0x0a0a1a;

    this.fogLayer.clear();

    const px = this.player.x;
    const py = this.player.y;

    const left = cam.scrollX - 50;
    const right = cam.scrollX + cam.width + 50;
    const top = cam.scrollY - 50;
    const bottom = cam.scrollY + cam.height + 50;

    // 室內模式
    if (this.isPlayerIndoor && this.currentBuildingId) {
      const building = this.client
        .getBuildings()
        .find((b) => b.id === this.currentBuildingId);

      if (building) {
        const halfW = building.width / 2;
        const halfH = building.height / 2;
        const bLeft = building.x - halfW;
        const bRight = building.x + halfW;
        const bTop = building.y - halfH;
        const bBottom = building.y + halfH;

        const fogAlpha = 0.95;
        this.fogLayer.fillStyle(fogColor, fogAlpha);

        // 上方區域
        this.fogLayer.fillRect(left, top, right - left, bTop - top);
        // 下方區域
        this.fogLayer.fillRect(left, bBottom, right - left, bottom - bBottom);
        // 左側區域
        this.fogLayer.fillRect(left, bTop, bLeft - left, bBottom - bTop);
        // 右側區域
        this.fogLayer.fillRect(bRight, bTop, right - bRight, bBottom - bTop);

        // 建築內部邊緣發光效果
        this.fogLayer.lineStyle(3, 0x4ecca3, 0.5);
        this.fogLayer.strokeRect(bLeft, bTop, building.width, building.height);

        return;
      }
    }

    // 室外模式
    const viewRadius = 250;
    const fogAlpha = 0.88;

    this.fogLayer.fillStyle(fogColor, fogAlpha);

    // 上方矩形
    if (py - viewRadius > top) {
      this.fogLayer.fillRect(left, top, right - left, py - viewRadius - top);
    }

    // 下方矩形
    if (py + viewRadius < bottom) {
      this.fogLayer.fillRect(
        left,
        py + viewRadius,
        right - left,
        bottom - (py + viewRadius),
      );
    }

    // 左右兩側用多個小矩形填充圓形外的區域
    const segments = 60;
    for (let i = 0; i < segments; i++) {
      const y1 = py - viewRadius + (i * viewRadius * 2) / segments;
      const y2 = py - viewRadius + ((i + 1) * viewRadius * 2) / segments;
      const yMid = (y1 + y2) / 2;

      const dy = Math.abs(yMid - py);
      if (dy < viewRadius) {
        const dx = Math.sqrt(viewRadius * viewRadius - dy * dy);

        if (px - dx > left) {
          this.fogLayer.fillRect(left, y1, px - dx - left, y2 - y1);
        }

        if (px + dx < right) {
          this.fogLayer.fillRect(px + dx, y1, right - (px + dx), y2 - y1);
        }
      }
    }

    // 視野邊緣漸層
    for (let r = viewRadius; r > viewRadius - 30; r -= 3) {
      const alpha = ((viewRadius - r) / 30) * fogAlpha;
      this.fogLayer.lineStyle(4, fogColor, alpha);
      this.fogLayer.strokeCircle(px, py, r);
    }

    // 視野邊界圈
    this.fogLayer.lineStyle(2, 0x4ecca3, 0.4);
    this.fogLayer.strokeCircle(px, py, viewRadius);
  }

  private createUI(): void {
    this.uiContainer = this.add.container(10, 10);
    this.uiContainer.setScrollFactor(0);
    this.uiContainer.setDepth(2000);

    const uiBg = this.add.rectangle(0, 0, 200, 100, 0x16213e, 0.9);
    uiBg.setOrigin(0, 0);
    uiBg.setStrokeStyle(2, 0x4ecca3);

    const uiTitle = this.add.text(10, 8, "📊 狀態", {
      fontSize: "14px",
      fontFamily: "Arial",
      color: "#4ecca3",
    });

    this.hpText = this.add.text(10, 30, "❤️ HP: 100", {
      fontSize: "14px",
      fontFamily: "Arial",
      color: "#ffffff",
    });

    this.scoreText = this.add.text(10, 50, "⭐ 分數: 0", {
      fontSize: "14px",
      fontFamily: "Arial",
      color: "#ffd700",
    });

    this.itemText = this.add.text(10, 70, "📦 物品: 0", {
      fontSize: "14px",
      fontFamily: "Arial",
      color: "#87ceeb",
    });

    this.uiContainer.add([
      uiBg,
      uiTitle,
      this.hpText,
      this.scoreText,
      this.itemText,
    ]);

    this.createMinimap();
  }

  private createMinimap(): void {
    const minimapSize = 150;
    const padding = 10;

    this.minimapContainer = this.add.container(
      this.cameras.main.width - minimapSize - padding,
      padding,
    );
    this.minimapContainer.setScrollFactor(0);
    this.minimapContainer.setDepth(2000);

    const minimapBg = this.add.rectangle(
      0,
      0,
      minimapSize,
      minimapSize,
      0x16213e,
      0.9,
    );
    minimapBg.setOrigin(0, 0);
    minimapBg.setStrokeStyle(2, 0x4ecca3);

    this.minimapGraphics = this.add.graphics();

    this.minimapContainer.add([minimapBg, this.minimapGraphics]);
  }

  private updateMinimap(): void {
    const minimapSize = 150;
    const scale = minimapSize / 3000;

    this.minimapGraphics.clear();

    if (this.isPlayerIndoor && this.currentBuildingId) {
      const building = this.client
        .getBuildings()
        .find((b) => b.id === this.currentBuildingId);

      if (building) {
        this.minimapGraphics.fillStyle(0x4ecca3, 0.5);
        this.minimapGraphics.fillRect(
          (building.x - building.width / 2) * scale,
          (building.y - building.height / 2) * scale,
          building.width * scale,
          building.height * scale,
        );

        this.minimapGraphics.lineStyle(2, 0x4ecca3, 1);
        this.minimapGraphics.strokeRect(
          (building.x - building.width / 2) * scale,
          (building.y - building.height / 2) * scale,
          building.width * scale,
          building.height * scale,
        );
      }
    } else {
      this.visibleObjects.buildings.forEach((_, id) => {
        const building = this.client.getBuildings().find((b) => b.id === id);
        if (building) {
          this.minimapGraphics.fillStyle(0x666666, 0.8);
          this.minimapGraphics.fillRect(
            (building.x - building.width / 2) * scale,
            (building.y - building.height / 2) * scale,
            building.width * scale,
            building.height * scale,
          );
        }
      });
    }

    // 玩家位置
    this.minimapGraphics.fillStyle(0x4ecca3, 1);
    this.minimapGraphics.fillCircle(
      this.player.x * scale,
      this.player.y * scale,
      4,
    );

    // 室外時畫視野範圍
    if (!this.isPlayerIndoor) {
      this.minimapGraphics.lineStyle(1, 0x4ecca3, 0.5);
      this.minimapGraphics.strokeCircle(
        this.player.x * scale,
        this.player.y * scale,
        250 * scale,
      );
    }

    // 其他玩家
    this.visibleObjects.players.forEach((sprite) => {
      this.minimapGraphics.fillStyle(0xe94560, 1);
      this.minimapGraphics.fillCircle(sprite.x * scale, sprite.y * scale, 3);
    });

    // 寶箱
    this.visibleObjects.treasures.forEach((sprite) => {
      this.minimapGraphics.fillStyle(0xffd700, 1);
      this.minimapGraphics.fillRect(
        sprite.x * scale - 2,
        sprite.y * scale - 2,
        4,
        4,
      );
    });

    // 敵人
    this.visibleObjects.enemies.forEach((sprite) => {
      this.minimapGraphics.fillStyle(0xff4444, 1);
      this.minimapGraphics.fillCircle(sprite.x * scale, sprite.y * scale, 3);
    });
  }

  private async connectToServer(): Promise<void> {
    this.updateStatus("正在連線...", "#4ecca3");

    await this.client.connect();

    this.updateStatus("已連線！加入遊戲中...", "#4ecca3");

    const initialData = this.client.join(this.player.x, this.player.y);
    if (initialData) {
      this.updateVisibleObjects(initialData);
    }

    this.updateStatus(
      `已加入遊戲 | 玩家ID: ${this.client.playerId.slice(-6)}`,
      "#4ecca3",
    );

    this.lastUpdatePos = { x: this.player.x, y: this.player.y };
  }

  private updateStatus(status: string, color: string): void {
    if (this.onStatusChange) {
      this.onStatusChange(status, color);
    }
  }

  private updateVisibleObjects(data: VisibleObjects): void {
    if (!data) return;

    // 確保場景還活著
    if (!this.scene || !this.scene.isActive()) return;

    const isIndoor = data.playerInside !== null;
    this.updateIndoorState(isIndoor, data.playerInside);

    // 處理寶箱
    const currentTreasureIds = new Set(data.treasures.map((t) => t.id));

    // 移除離開視野的寶箱
    const treasuresToRemove: string[] = [];
    for (const [id, sprite] of this.visibleObjects.treasures) {
      if (!currentTreasureIds.has(id)) {
        sprite.destroy();
        treasuresToRemove.push(id);
      }
    }
    treasuresToRemove.forEach((id) => this.visibleObjects.treasures.delete(id));

    for (const treasure of data.treasures) {
      if (!this.visibleObjects.treasures.has(treasure.id)) {
        const texture = treasure.type === "gold" ? "goldChest" : "silverChest";
        const sprite = this.add.sprite(treasure.x, treasure.y, texture);
        sprite.setDepth(10);
        sprite.setData("objectData", treasure);

        sprite.setAlpha(0);
        sprite.setScale(0.5);
        this.tweens.add({
          targets: sprite,
          alpha: 1,
          scale: 1,
          duration: 300,
          ease: "Back.easeOut",
        });

        this.visibleObjects.treasures.set(treasure.id, sprite);
        this.treasureGroup.add(sprite);
      }
    }

    // 處理敵人
    const currentEnemyIds = new Set(data.enemies.map((e) => e.id));

    // 移除離開視野的敵人
    const enemiesToRemove: string[] = [];
    for (const [id, sprite] of this.visibleObjects.enemies) {
      if (!currentEnemyIds.has(id)) {
        sprite.destroy();
        enemiesToRemove.push(id);
      }
    }
    enemiesToRemove.forEach((id) => this.visibleObjects.enemies.delete(id));

    for (const enemy of data.enemies) {
      if (!this.visibleObjects.enemies.has(enemy.id)) {
        const texture = enemy.type === "skeleton" ? "skeleton" : "goblin";
        const sprite = this.add.sprite(enemy.x, enemy.y, texture);
        sprite.setDepth(10);
        sprite.setData("objectData", enemy);

        sprite.setAlpha(0);
        this.tweens.add({
          targets: sprite,
          alpha: 1,
          duration: 200,
        });

        this.tweens.add({
          targets: sprite,
          x: sprite.x + (Math.random() - 0.5) * 50,
          y: sprite.y + (Math.random() - 0.5) * 50,
          duration: 2000,
          yoyo: true,
          repeat: -1,
          ease: "Sine.easeInOut",
        });

        this.visibleObjects.enemies.set(enemy.id, sprite);
        this.enemyGroup.add(sprite);
      } else {
        const sprite = this.visibleObjects.enemies.get(enemy.id)!;
        const objData = sprite.getData("objectData") as Enemy;
        objData.hp = enemy.hp;
      }
    }

    // 處理建築
    const currentBuildingIds = new Set(data.buildings.map((b) => b.id));

    // 移除離開視野的建築
    const buildingsToRemove: string[] = [];
    for (const [id, buildingData] of this.visibleObjects.buildings) {
      if (!currentBuildingIds.has(id)) {
        if (buildingData.graphics) buildingData.graphics.destroy();
        if (buildingData.roof) buildingData.roof.destroy();
        if (buildingData.floor) buildingData.floor.destroy();
        if (buildingData.doorGraphics) buildingData.doorGraphics.destroy();
        if (buildingData.arrowGraphics) buildingData.arrowGraphics.destroy();
        if (buildingData.typeLabel) buildingData.typeLabel.destroy();
        if (buildingData.wallSprites) {
          buildingData.wallSprites.forEach((w) => w.destroy());
        }
        buildingsToRemove.push(id);
      }
    }
    buildingsToRemove.forEach((id) => this.visibleObjects.buildings.delete(id));

    for (const building of data.buildings) {
      if (!this.visibleObjects.buildings.has(building.id)) {
        const buildingData = this.createBuildingGraphics(building);
        this.visibleObjects.buildings.set(building.id, buildingData);
      } else {
        const buildingData = this.visibleObjects.buildings.get(building.id)!;
        if (buildingData.roof) {
          buildingData.roof.setVisible(!building.playerInside);
        }
        if (buildingData.typeLabel) {
          buildingData.typeLabel.setVisible(!building.playerInside);
        }
      }
    }

    // 處理其他玩家
    const currentPlayerIds = new Set(data.players.map((p) => p.id));

    // 移除離開視野的玩家
    const playersToRemove: string[] = [];
    for (const [id, sprite] of this.visibleObjects.players) {
      if (!currentPlayerIds.has(id)) {
        sprite.destroy();
        playersToRemove.push(id);
      }
    }
    playersToRemove.forEach((id) => this.visibleObjects.players.delete(id));

    for (const playerData of data.players) {
      if (!this.visibleObjects.players.has(playerData.id)) {
        const sprite = this.add.sprite(
          playerData.x,
          playerData.y,
          "otherTreasurePlayer",
        );
        sprite.setDepth(99);
        sprite.setData("objectData", playerData);

        this.visibleObjects.players.set(playerData.id, sprite);
        this.playerGroup.add(sprite);
      } else {
        const sprite = this.visibleObjects.players.get(playerData.id)!;
        sprite.x = playerData.x;
        sprite.y = playerData.y;
        sprite.setData("objectData", playerData);
      }
    }
  }

  private updateIndoorState(
    isIndoor: boolean,
    buildingId: string | null,
  ): void {
    if (
      this.isPlayerIndoor === isIndoor &&
      this.currentBuildingId === buildingId
    ) {
      return;
    }

    this.isPlayerIndoor = isIndoor;
    this.currentBuildingId = buildingId;

    if (isIndoor) {
      if (!this.indoorIndicator) {
        this.indoorIndicator = this.add.text(
          this.cameras.main.width / 2,
          50,
          "🏠 室內",
          {
            fontSize: "18px",
            fontFamily: "Arial",
            color: "#ffcc00",
            stroke: "#000000",
            strokeThickness: 4,
            backgroundColor: "#00000088",
            padding: { x: 15, y: 8 },
          },
        );
        this.indoorIndicator.setOrigin(0.5);
        this.indoorIndicator.setScrollFactor(0);
        this.indoorIndicator.setDepth(2001);
      }
      this.indoorIndicator.setVisible(true);

      this.indoorIndicator.setAlpha(0);
      this.tweens.add({
        targets: this.indoorIndicator,
        alpha: 1,
        duration: 300,
      });
    } else {
      if (this.indoorIndicator) {
        this.indoorIndicator.setVisible(false);
      }
    }
  }

  private createBuildingGraphics(building: Building): BuildingGraphicsData {
    const { x, y, width, height, type, walls, hasRoof, door } = building;
    const halfW = width / 2;
    const halfH = height / 2;

    const buildingData: BuildingGraphicsData = {
      id: building.id,
      wallSprites: [],
    };

    // 地板
    const floor = this.add.graphics();
    floor.fillStyle(this.getBuildingFloorColor(type), 1);
    floor.fillRect(x - halfW, y - halfH, width, height);

    // 地板紋理
    floor.lineStyle(1, 0x000000, 0.1);
    const tileSize = 30;
    for (let tx = x - halfW; tx < x + halfW; tx += tileSize) {
      floor.lineBetween(tx, y - halfH, tx, y + halfH);
    }
    for (let ty = y - halfH; ty < y + halfH; ty += tileSize) {
      floor.lineBetween(x - halfW, ty, x + halfW, ty);
    }

    floor.setDepth(1);
    buildingData.floor = floor;

    // 門口標示
    if (door) {
      const doorGraphics = this.add.graphics();
      doorGraphics.setDepth(2);

      const doorDepth = 25;
      let dx = door.x,
        dy = door.y;
      let dw = door.width,
        dh = doorDepth;

      if (door.side === "top") {
        dx = door.x - door.width / 2;
        dy = door.y - doorDepth / 2;
      } else if (door.side === "bottom") {
        dx = door.x - door.width / 2;
        dy = door.y - doorDepth / 2;
      } else if (door.side === "left") {
        dx = door.x - doorDepth / 2;
        dy = door.y - door.width / 2;
        dw = doorDepth;
        dh = door.width;
      } else if (door.side === "right") {
        dx = door.x - doorDepth / 2;
        dy = door.y - door.width / 2;
        dw = doorDepth;
        dh = door.width;
      }

      doorGraphics.fillStyle(0xd4a574, 1);
      doorGraphics.fillRect(dx, dy, dw, dh);

      doorGraphics.lineStyle(3, 0xffcc00, 0.8);
      doorGraphics.strokeRect(dx, dy, dw, dh);

      // 門口箭頭指示
      const arrowGraphics = this.add.graphics();
      arrowGraphics.setDepth(3);
      arrowGraphics.fillStyle(0xffcc00, 0.9);

      const arrowSize = 12;
      let ax = door.x,
        ay = door.y;

      if (door.side === "top") {
        ay -= 35;
        arrowGraphics.fillTriangle(
          ax,
          ay + arrowSize,
          ax - arrowSize,
          ay - arrowSize,
          ax + arrowSize,
          ay - arrowSize,
        );
      } else if (door.side === "bottom") {
        ay += 35;
        arrowGraphics.fillTriangle(
          ax,
          ay - arrowSize,
          ax - arrowSize,
          ay + arrowSize,
          ax + arrowSize,
          ay + arrowSize,
        );
      } else if (door.side === "left") {
        ax -= 35;
        arrowGraphics.fillTriangle(
          ax + arrowSize,
          ay,
          ax - arrowSize,
          ay - arrowSize,
          ax - arrowSize,
          ay + arrowSize,
        );
      } else if (door.side === "right") {
        ax += 35;
        arrowGraphics.fillTriangle(
          ax - arrowSize,
          ay,
          ax + arrowSize,
          ay - arrowSize,
          ax + arrowSize,
          ay + arrowSize,
        );
      }

      this.tweens.add({
        targets: arrowGraphics,
        alpha: 0.3,
        duration: 800,
        yoyo: true,
        repeat: -1,
        ease: "Sine.easeInOut",
      });

      buildingData.doorGraphics = doorGraphics;
      buildingData.arrowGraphics = arrowGraphics;
    }

    // 牆壁
    const wallGraphics = this.add.graphics();
    wallGraphics.setDepth(50);

    walls.forEach((wall) => {
      wallGraphics.fillStyle(0x000000, 0.3);
      wallGraphics.fillRect(wall.x + 3, wall.y + 3, wall.width, wall.height);

      wallGraphics.fillStyle(this.getBuildingWallColor(type), 1);
      wallGraphics.fillRect(wall.x, wall.y, wall.width, wall.height);

      wallGraphics.fillStyle(0xffffff, 0.15);
      wallGraphics.fillRect(wall.x, wall.y, wall.width, 3);

      wallGraphics.lineStyle(1, 0x000000, 0.5);
      wallGraphics.strokeRect(wall.x, wall.y, wall.width, wall.height);

      // 碰撞體
      const wallSprite = this.physics.add.staticSprite(
        wall.x + wall.width / 2,
        wall.y + wall.height / 2,
        undefined as unknown as string,
      );
      wallSprite.body?.setSize(wall.width, wall.height);
      wallSprite.setVisible(false);
      this.wallGroup.add(wallSprite);
      buildingData.wallSprites.push(wallSprite);
    });

    buildingData.graphics = wallGraphics;

    // 屋頂
    if (hasRoof) {
      const roof = this.add.graphics();

      roof.fillStyle(this.getBuildingRoofColor(type), 0.97);
      roof.fillRect(x - halfW - 8, y - halfH - 8, width + 16, height + 16);

      roof.lineStyle(3, this.getBuildingRoofColor(type) - 0x333333, 1);
      roof.strokeRect(x - halfW - 8, y - halfH - 8, width + 16, height + 16);

      roof.fillStyle(this.getBuildingRoofColor(type) - 0x222222, 1);
      if (type === "house") {
        roof.fillRect(x + halfW - 40, y - halfH - 20, 25, 25);
        roof.fillStyle(0x555555, 1);
        roof.fillRect(x + halfW - 38, y - halfH - 18, 21, 21);
      } else if (type === "tower") {
        roof.fillCircle(x, y, Math.min(halfW, halfH) * 0.4);
      } else if (type === "shrine") {
        roof.fillStyle(0xffd700, 1);
        roof.fillRect(x - 15, y - 15, 30, 30);
        roof.fillStyle(0xffffff, 1);
        roof.fillCircle(x, y, 8);
      }

      const typeLabel = this.add.text(
        x,
        y - halfH - 25,
        this.getBuildingLabel(type),
        {
          fontSize: "14px",
          fontFamily: "Arial",
          color: "#ffffff",
          stroke: "#000000",
          strokeThickness: 3,
        },
      );
      typeLabel.setOrigin(0.5);
      typeLabel.setDepth(201);
      buildingData.typeLabel = typeLabel;

      roof.setDepth(200);
      buildingData.roof = roof;
    }

    return buildingData;
  }

  private getBuildingLabel(type: string): string {
    const labels: Record<string, string> = {
      house: "🏠 房屋",
      tower: "🗼 塔樓",
      ruins: "🏚️ 廢墟",
      shrine: "⛩️ 神殿",
    };
    return labels[type] || "建築";
  }

  private getBuildingFloorColor(type: string): number {
    const colors: Record<string, number> = {
      house: 0x8b7355,
      tower: 0x606060,
      ruins: 0x4a5240,
      shrine: 0xdaa520,
    };
    return colors[type] || 0x808080;
  }

  private getBuildingWallColor(type: string): number {
    const colors: Record<string, number> = {
      house: 0x654321,
      tower: 0x505050,
      ruins: 0x3d4235,
      shrine: 0xb8860b,
    };
    return colors[type] || 0x606060;
  }

  private getBuildingRoofColor(type: string): number {
    const colors: Record<string, number> = {
      house: 0x8b4513,
      tower: 0x404040,
      ruins: 0x2d3228,
      shrine: 0xffd700,
    };
    return colors[type] || 0x505050;
  }

  update(): void {
    if (!this.client.connected) return;

    // 處理移動
    const speed = 200;
    let vx = 0,
      vy = 0;

    if (this.cursors.left.isDown || this.wasd.left.isDown) vx = -speed;
    else if (this.cursors.right.isDown || this.wasd.right.isDown) vx = speed;

    if (this.cursors.up.isDown || this.wasd.up.isDown) vy = -speed;
    else if (this.cursors.down.isDown || this.wasd.down.isDown) vy = speed;

    this.player.setVelocity(vx, vy);

    // 檢查是否需要更新後端
    const dx = this.player.x - this.lastUpdatePos.x;
    const dy = this.player.y - this.lastUpdatePos.y;
    const distance = Math.sqrt(dx * dx + dy * dy);

    if (distance > this.updateThreshold) {
      const visibleData = this.client.updatePosition(
        this.player.x,
        this.player.y,
      );
      if (visibleData) {
        this.updateVisibleObjects(visibleData);
      }
      this.lastUpdatePos = { x: this.player.x, y: this.player.y };
    }

    // 處理攻擊
    if (Phaser.Input.Keyboard.JustDown(this.wasd.attack)) {
      this.handleAttack();
    }

    // 處理撿取
    if (Phaser.Input.Keyboard.JustDown(this.wasd.collect)) {
      this.handleCollect();
    }

    // 更新迷霧
    this.updateFogOfWar();

    // 更新小地圖
    this.updateMinimap();

    // 更新 UI
    this.updateUI();
  }

  private handleAttack(): void {
    this.attackEffect.setPosition(this.player.x, this.player.y);
    this.attackEffect.setVisible(true);
    this.attackEffect.setAlpha(0.5);
    this.tweens.add({
      targets: this.attackEffect,
      alpha: 0,
      scale: 1.5,
      duration: 200,
      onComplete: () => {
        this.attackEffect.setVisible(false);
        this.attackEffect.setScale(1);
      },
    });

    let attacked = false;

    // 攻擊敵人
    this.visibleObjects.enemies.forEach((sprite, id) => {
      if (attacked) return;
      const dx = sprite.x - this.player.x;
      const dy = sprite.y - this.player.y;
      const dist = Math.sqrt(dx * dx + dy * dy);

      if (dist <= 60) {
        const result = this.client.attack(id, "enemy");
        if (result.success) {
          attacked = true;

          this.tweens.add({
            targets: sprite,
            tint: 0xff0000,
            duration: 100,
            yoyo: true,
          });

          if (result.killed) {
            this.tweens.add({
              targets: sprite,
              alpha: 0,
              scale: 0,
              duration: 300,
              onComplete: () => {
                sprite.destroy();
                this.visibleObjects.enemies.delete(id);
              },
            });
          }
        }
      }
    });

    // 攻擊其他玩家
    if (!attacked) {
      this.visibleObjects.players.forEach((sprite, id) => {
        if (attacked) return;
        const dx = sprite.x - this.player.x;
        const dy = sprite.y - this.player.y;
        const dist = Math.sqrt(dx * dx + dy * dy);

        if (dist <= 60) {
          const result = this.client.attack(id, "player");
          if (result.success) {
            attacked = true;
            this.tweens.add({
              targets: sprite,
              tint: 0xff0000,
              duration: 100,
              yoyo: true,
            });
          }
        }
      });
    }
  }

  private handleCollect(): void {
    this.visibleObjects.treasures.forEach((sprite, id) => {
      const dx = sprite.x - this.player.x;
      const dy = sprite.y - this.player.y;
      const dist = Math.sqrt(dx * dx + dy * dy);

      if (dist <= 50) {
        const result = this.client.collect(id);
        if (result.success && result.treasure) {
          this.tweens.add({
            targets: sprite,
            y: sprite.y - 30,
            alpha: 0,
            scale: 0,
            duration: 300,
            onComplete: () => {
              sprite.destroy();
              this.visibleObjects.treasures.delete(id);
            },
          });

          const text = this.add.text(
            sprite.x,
            sprite.y - 20,
            `+${result.treasure.value}`,
            {
              fontSize: "20px",
              fontFamily: "Arial",
              color: "#ffd700",
              stroke: "#000",
              strokeThickness: 3,
            },
          );
          text.setOrigin(0.5);
          text.setDepth(1500);

          this.tweens.add({
            targets: text,
            y: text.y - 40,
            alpha: 0,
            duration: 1000,
            onComplete: () => text.destroy(),
          });
        }
      }
    });
  }

  private updateUI(): void {
    const playerInfo = this.client.getPlayerInfo();
    if (playerInfo) {
      this.hpText.setText(`❤️ HP: ${playerInfo.hp}`);
      this.scoreText.setText(`⭐ 分數: ${playerInfo.score}`);
      this.itemText.setText(`📦 物品: ${playerInfo.inventory.length}`);
    }

    if (this.isPlayerIndoor) {
      this.updateStatus(
        `🏠 室內探索中 | 玩家ID: ${this.client.playerId.slice(-6)}`,
        "#ffcc00",
      );
    } else {
      this.updateStatus(
        `🌲 室外探索中 | 玩家ID: ${this.client.playerId.slice(-6)}`,
        "#4ecca3",
      );
    }
  }
}
