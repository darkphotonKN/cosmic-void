/**
 * TreasureHuntScene - 簡化版遊戲場景
 * 移動邏輯 + WebSocket + 建築（進入後看不到外面）
 */

import Phaser from "phaser";
import { ActionType } from "@/assets/types/client";
import { socketManager } from "@/utils/class/SocketManager";
import { ClientGameState, ContainerState, ItemState } from "@/types/gameState";
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

export class TreasureHuntScene extends Phaser.Scene {
  private player?: Phaser.Physics.Arcade.Sprite;
  private otherPlayers: Map<string, Phaser.Physics.Arcade.Sprite> = new Map();
  private otherPlayersTargets: Map<string, { x: number; y: number }> =
    new Map();

  // eye graphics for directional looking
  private playerEyes?: Phaser.GameObjects.Graphics;
  private otherPlayersEyes: Map<string, Phaser.GameObjects.Graphics> = new Map();

  // Controls
  private cursors!: Phaser.Types.Input.Keyboard.CursorKeys;
  private wasd!: {
    up: Phaser.Input.Keyboard.Key;
    down: Phaser.Input.Keyboard.Key;
    left: Phaser.Input.Keyboard.Key;
    right: Phaser.Input.Keyboard.Key;
  };

  // Status callback
  private onStatusChange?: (status: string, color: string) => void;

  // Game state
  private gameStateUnsubscribe?: () => void;
  private targetPosition: { x: number; y: number } | null = null;

  // 地圖大小
  private mapWidth = 1200;
  private mapHeight = 800;

  // 建築
  private buildings: Building[] = [];
  private currentBuilding: Building | null = null;
  private outsideObjects: Phaser.GameObjects.GameObject[] = [];
  private indoorMask!: Phaser.GameObjects.Graphics;

  // 寶箱 (從後端同步)
  private chests: Map<string, { sprite: Phaser.GameObjects.Sprite; entityId: string }> = new Map();

  // 寶箱跳窗
  private chestPopup?: Phaser.GameObjects.Container;
  private isPopupOpen = false;
  private openedChestEntityId?: string;
  private popupItemsText?: Phaser.GameObjects.Text;

  // 道具欄
  private inventoryPopup?: Phaser.GameObjects.Container;
  private isInventoryOpen = false;
  private inventoryItems: ItemState[] = [];
  private inventoryItemsText?: Phaser.GameObjects.Text;

  // 當前寶箱的物品（用於 F 鍵取得）
  private currentChestItems: ItemState[] = [];
  private chestLootedAtMap = new Map<string, number>(); // entityId → loot 時間戳
  private readonly PENDING_DURATION = 1000; // 1 秒內不比對剛拿的物品

  constructor() {
    super({ key: "TreasureHuntScene" });
  }

  setStatusCallback(callback: (status: string, color: string) => void): void {
    this.onStatusChange = callback;
  }

  preload(): void {
    // create styled circle textures
    this.createPlayerTexture();
    this.createOtherPlayerTexture();
    this.createChestTextures();
  }

  private createPlayerTexture(): void {
    const graphics = this.make.graphics({});
    // outer glow effect
    graphics.fillStyle(0x4ecca3, 0.3);
    graphics.fillCircle(20, 20, 20);
    // main body with gradient effect
    graphics.fillStyle(0x4ecca3, 1);
    graphics.fillCircle(20, 20, 18);
    // inner lighter circle for depth
    graphics.fillStyle(0x6effc8, 0.8);
    graphics.fillCircle(20, 18, 12);
    // no eyes here, they'll be drawn separately
    graphics.generateTexture("player", 40, 40);
    graphics.destroy();
  }

  private createOtherPlayerTexture(): void {
    const graphics = this.make.graphics({});
    // outer glow effect
    graphics.fillStyle(0xff6b6b, 0.3);
    graphics.fillCircle(20, 20, 20);
    // main body
    graphics.fillStyle(0xff6b6b, 1);
    graphics.fillCircle(20, 20, 18);
    // inner lighter circle
    graphics.fillStyle(0xff9999, 0.8);
    graphics.fillCircle(20, 18, 12);
    // no eyes here, they'll be drawn separately
    graphics.generateTexture("otherPlayer", 40, 40);
    graphics.destroy();
  }

  private drawEyes(graphics: Phaser.GameObjects.Graphics, x: number, y: number, vx: number, vy: number, isPlayer: boolean): void {
    graphics.clear();

    // calculate eye offset based on movement direction
    const maxOffset = 2;
    let eyeOffsetX = vx * maxOffset;
    let eyeOffsetY = vy * maxOffset;

    // base eye positions (relative to sprite center)
    const leftEyeX = x - 6 + eyeOffsetX;
    const leftEyeY = y - 4 + eyeOffsetY;
    const rightEyeX = x + 6 + eyeOffsetX;
    const rightEyeY = y - 4 + eyeOffsetY;

    // eye sockets (dark background)
    graphics.fillStyle(isPlayer ? 0x001122 : 0x220011, 1);
    graphics.fillCircle(x - 6, y - 4, 4);
    graphics.fillCircle(x + 6, y - 4, 4);

    // pupils (move with direction)
    graphics.fillStyle(isPlayer ? 0x00ffff : 0xff6666, 1);
    graphics.fillCircle(leftEyeX, leftEyeY, 2);
    graphics.fillCircle(rightEyeX, rightEyeY, 2);

    // eye shine
    graphics.fillStyle(0xffffff, 0.7);
    graphics.fillCircle(leftEyeX + 0.5, leftEyeY - 0.5, 0.8);
    graphics.fillCircle(rightEyeX + 0.5, rightEyeY - 0.5, 0.8);
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

  private updateContainers(containers: ContainerState[]): void {
    const activeEntityIds = new Set(containers.map(c => c.entity_id));

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
          container.is_open ? "chest_open" : "chest_closed"
        );
        sprite.setDepth(50);
        chest = { sprite, entityId: container.entity_id };
        this.chests.set(container.entity_id, chest);
      } else {
        // 更新寶箱狀態
        chest.sprite.setTexture(container.is_open ? "chest_open" : "chest_closed");
        chest.sprite.setPosition(container.position.x, container.position.y);
      }

      // 如果是打開的寶箱，更新跳窗內容
      if (container.is_open && this.openedChestEntityId === container.entity_id) {
        this.updatePopupItems(container.items, container.entity_id);
      }
    });
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
    }
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
    const popupWidth = 300;
    const popupHeight = 250;

    // 半透明背景
    const bg = this.add.graphics();
    bg.fillStyle(0x000000, 0.7);
    bg.fillRoundedRect(
      -popupWidth / 2,
      -popupHeight / 2,
      popupWidth,
      popupHeight,
      16,
    );
    bg.lineStyle(2, 0xffd700, 1);
    bg.strokeRoundedRect(
      -popupWidth / 2,
      -popupHeight / 2,
      popupWidth,
      popupHeight,
      16,
    );

    // 標題
    const title = this.add.text(0, -popupHeight / 2 + 20, "Chest Contents", {
      fontSize: "20px",
      color: "#ffd700",
    });
    title.setOrigin(0.5);

    // 物品區
    this.popupItemsText = this.add.text(0, 0, "Loading...", {
      fontSize: "14px",
      color: "#ffffff",
      align: "center",
      lineSpacing: 8,
    });
    this.popupItemsText.setOrigin(0.5);

    // 底部快捷鍵
    const hint = this.add.text(0, popupHeight / 2 - 25, "E - Close  |  F - Take", {
      fontSize: "12px",
      color: "#aaaaaa",
    });
    hint.setOrigin(0.5);

    // Container
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
    if (!this.popupItemsText) return;

    const chestId = entityId || this.openedChestEntityId;
    if (!chestId) return;

    const now = Date.now();
    const lootedAt = this.chestLootedAtMap.get(chestId);
    const isPending = lootedAt && (now - lootedAt) < this.PENDING_DURATION;

    // 如果剛 loot 過，等後端確認清空才更新
    if (isPending) {
      if (items.length === 0) {
        // 後端已確認清空
        this.chestLootedAtMap.delete(chestId);
      } else {
        // 後端還沒處理完，忽略這次更新
        return;
      }
    }

    // 儲存當前寶箱物品（用於 F 鍵取得）
    this.currentChestItems = items.map(item => ({ ...item }));

    if (items.length === 0) {
      this.popupItemsText.setText("(Empty)");
    } else {
      const itemLines = items.map(item => `${item.name} x${item.quantity}`);
      this.popupItemsText.setText(itemLines.join("\n"));
    }
  }

  private hideChestPopup(): void {
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
    if (this.isInventoryOpen) {
      this.hideInventory();
    } else {
      this.showInventory();
    }
  }

  private showInventory(): void {
    if (this.isInventoryOpen) return;

    const centerX = this.cameras.main.width / 2;
    const centerY = this.cameras.main.height / 2;
    const popupWidth = 300;
    const popupHeight = 300;

    // 半透明背景
    const bg = this.add.graphics();
    bg.fillStyle(0x000000, 0.8);
    bg.fillRoundedRect(
      -popupWidth / 2,
      -popupHeight / 2,
      popupWidth,
      popupHeight,
      16,
    );
    bg.lineStyle(2, 0x4ecca3, 1);
    bg.strokeRoundedRect(
      -popupWidth / 2,
      -popupHeight / 2,
      popupWidth,
      popupHeight,
      16,
    );

    // 標題
    const title = this.add.text(0, -popupHeight / 2 + 20, "Inventory", {
      fontSize: "22px",
      color: "#4ecca3",
    });
    title.setOrigin(0.5);

    // 物品列表
    this.inventoryItemsText = this.add.text(0, 0, "", {
      fontSize: "14px",
      color: "#ffffff",
      align: "center",
      lineSpacing: 8,
    });
    this.inventoryItemsText.setOrigin(0.5);
    this.updateInventoryDisplay();

    // 底部快捷鍵
    const hint = this.add.text(0, popupHeight / 2 - 25, "I - Close", {
      fontSize: "12px",
      color: "#aaaaaa",
    });
    hint.setOrigin(0.5);

    // Container
    this.inventoryPopup = this.add.container(centerX, centerY, [
      bg,
      title,
      this.inventoryItemsText,
      hint,
    ]);
    this.inventoryPopup.setDepth(1001);
    this.inventoryPopup.setScrollFactor(0);

    this.isInventoryOpen = true;
  }

  private hideInventory(): void {
    if (this.inventoryPopup) {
      this.inventoryPopup.destroy();
      this.inventoryPopup = undefined;
    }
    this.inventoryItemsText = undefined;
    this.isInventoryOpen = false;
  }

  private updateInventoryDisplay(): void {
    if (!this.inventoryItemsText) return;

    // 合併同名物品顯示
    const merged = new Map<string, number>();
    for (const item of this.inventoryItems) {
      merged.set(item.name, (merged.get(item.name) || 0) + item.quantity);
    }

    if (merged.size === 0) {
      this.inventoryItemsText.setText("(Empty)");
    } else {
      const itemLines = Array.from(merged.entries()).map(([name, qty]) => `${name} x${qty}`);
      this.inventoryItemsText.setText(itemLines.join("\n"));
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
      const isPending = localItem.lootedAt && (now - localItem.lootedAt) < this.PENDING_DURATION;

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

    // 更新顯示
    if (this.isInventoryOpen) {
      this.updateInventoryDisplay();
    }
  }

  private pickupItemFromChest(): void {
    // 只有在寶箱跳窗開啟時才能取得道具
    if (!this.isPopupOpen || this.currentChestItems.length === 0 || !this.openedChestEntityId) {
      return;
    }

    // 收集所有 item entity IDs
    const itemEntityIds = this.currentChestItems.map(item => item.entity_id);

    // 發送 loot action 到後端
    socketManager.sendMessage(ActionType.Loot, {
      container_entity_id: this.openedChestEntityId,
      item_entity_ids: itemEntityIds,
    });

    // 記錄 loot 時間
    const now = Date.now();
    this.chestLootedAtMap.set(this.openedChestEntityId, now);

    // 取得全部道具到本地道具欄
    for (const item of this.currentChestItems) {
      this.inventoryItems.push({
        ...item,
        lootedAt: now,
      });
    }

    // 清空寶箱（本地）
    this.currentChestItems = [];

    // 更新寶箱跳窗顯示
    this.updatePopupItems(this.currentChestItems);

    // 如果道具欄開啟，也更新顯示
    if (this.isInventoryOpen) {
      this.updateInventoryDisplay();
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

  private createPlayer(x: number, y: number): void {
    this.player = this.physics.add.sprite(x, y, 'player');
    this.player.setCollideWorldBounds(true);
    this.player.setDepth(100);

    // set circular physics body to match backend collision (radius 20)
    this.player.body.setCircle(20);

    // create eyes overlay that will follow player
    this.playerEyes = this.add.graphics();
    this.playerEyes.setDepth(101);
    this.drawEyes(this.playerEyes, 0, 0, 0, 0, true); // default looking forward

    // 玩家與所有建築牆壁/門碰撞
    this.buildings.forEach((building) => {
      this.physics.add.collider(this.player!, building.wallGroup);
      this.physics.add.collider(this.player!, building.doorCollider);
    });

    // 相機跟隨玩家
    this.cameras.main.startFollow(this.player, true, 0.1, 0.1);
  }

  create(): void {
    // Connect via SocketManager
    this.connectToServer();

    // setup world boundaries
    this.physics.world.setBounds(0, 0, this.mapWidth, this.mapHeight);

    // create map background with cosmic theme
    this.createMapBackground();

    // create buildings
    this.createBuildings();

    // 寶箱由後端同步，不在這裡創建

    // 設置相機邊界（玩家建立後再 follow）
    this.cameras.main.setBounds(0, 0, this.mapWidth, this.mapHeight);

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
      this.scene.start("MainMenuScene");
    });

    // E 鍵互動（門、寶箱）
    this.input.keyboard?.on("keydown-E", () => {
      // 檢查門
      const nearbyBuilding = this.getNearbyBuilding();
      if (nearbyBuilding) {
        this.toggleDoor(nearbyBuilding);
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
      this.pickupItemFromChest();
    });

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
          this.updateStatus("Connected", "#4ecca3");
          GameStateLogger.logConnectionStatus(
            "Connected successfully!",
            "#4ecca3",
          );
          break;
        case "connecting":
          this.updateStatus("Connecting...", "#ffcc00");
          break;
        case "disconnected":
          this.updateStatus("Disconnected", "#ff4444");
          GameStateLogger.logConnectionStatus(
            "Disconnected from server",
            "#ff4444",
          );
          break;
        case "error":
          this.updateStatus("Connection Error", "#ff4444");
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

    // Reset the logger for new session
    GameStateLogger.reset();
  }

  private handleGameStateUpdate(state: ClientGameState): void {
    // Update current player position from server
    if (state.current_player) {
      const pos = state.current_player.position;

      // 第一次收到位置，建立玩家
      if (!this.player) {
        this.createPlayer(pos.x, pos.y);
      }

      // 設定目標位置，在 update() 中平滑移動
      this.targetPosition = { x: pos.x, y: pos.y };

      // 同步玩家背包
      if (state.current_player.inventory) {
        this.syncInventory(state.current_player.inventory);
      }

      const playerCount = (state.other_players?.length || 0) + 1;
      this.updateStatus(
        `Players: ${playerCount} | You: (${pos.x.toFixed(0)}, ${pos.y.toFixed(0)})`,
        "#4ecca3",
      );
    } else {
      this.updateStatus("Waiting for player data...", "#ffcc00");
    }

    // Update other players on screen
    this.updateOtherPlayers(state.other_players || []);

    // Update containers from server
    this.updateContainers(state.containers || []);
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
        sprite.destroy();
        this.otherPlayers.delete(playerId);
        this.otherPlayersTargets.delete(playerId);

        // remove eyes too
        const eyes = this.otherPlayersEyes.get(playerId);
        if (eyes) {
          eyes.destroy();
          this.otherPlayersEyes.delete(playerId);
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
          'otherPlayer',
        );
        sprite.setDepth(99);

        // set circular physics body
        sprite.body.setCircle(20);

        this.otherPlayers.set(playerData.id, sprite);

        // create eyes for this other player
        const eyes = this.add.graphics();
        eyes.setDepth(100);
        this.otherPlayersEyes.set(playerData.id, eyes);
        this.drawEyes(eyes, playerData.position.x, playerData.position.y, 0, 0, false);
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

    // cosmic background gradient
    const gradient = this.add.graphics();
    for (let i = 0; i < this.mapHeight; i++) {
      const ratio = i / this.mapHeight;
      const color = Phaser.Display.Color.Interpolate.ColorWithColor(
        { r: 10, g: 10, b: 20 },
        { r: 30, g: 15, b: 60 },
        1,
        ratio
      );
      gradient.fillStyle(Phaser.Display.Color.GetColor(color.r, color.g, color.b), 1);
      gradient.fillRect(0, i, this.mapWidth, 1);
    }
    gradient.setDepth(-2);

    // add stars for cosmic effect
    const starGraphics = this.add.graphics();
    for (let i = 0; i < 100; i++) {
      const x = Phaser.Math.Between(0, this.mapWidth);
      const y = Phaser.Math.Between(0, this.mapHeight);
      const size = Phaser.Math.FloatBetween(0.5, 2);
      const alpha = Phaser.Math.FloatBetween(0.3, 0.9);

      starGraphics.fillStyle(0xffffff, alpha);
      starGraphics.fillCircle(x, y, size);
    }

    // add some twinkling stars
    for (let i = 0; i < 20; i++) {
      const x = Phaser.Math.Between(0, this.mapWidth);
      const y = Phaser.Math.Between(0, this.mapHeight);
      const star = this.add.graphics();
      star.fillStyle(0xffffcc, 1);
      star.fillCircle(0, 0, 1.5);
      star.setPosition(x, y);
      star.setDepth(-1);

      this.tweens.add({
        targets: star,
        alpha: 0.2,
        duration: Phaser.Math.Between(1000, 3000),
        ease: 'Sine.easeInOut',
        yoyo: true,
        repeat: -1,
        delay: Phaser.Math.Between(0, 2000)
      });
    }

    starGraphics.setDepth(-1);

    // grid lines with lower opacity and cosmic color
    graphics.lineStyle(1, 0x4a4a8e, 0.15);
    for (let x = 0; x <= this.mapWidth; x += 50) {
      graphics.lineBetween(x, 0, x, this.mapHeight);
    }
    for (let y = 0; y <= this.mapHeight; y += 50) {
      graphics.lineBetween(0, y, this.mapWidth, y);
    }

    // boundary with cosmic glow effect
    graphics.lineStyle(2, 0x00ffcc, 0.4);
    graphics.strokeRect(2, 2, this.mapWidth - 4, this.mapHeight - 4);
    graphics.lineStyle(3, 0x4ecca3, 0.6);
    graphics.strokeRect(0, 0, this.mapWidth, this.mapHeight);

    graphics.setDepth(-1);

    // save as outdoor objects
    this.outsideObjects.push(graphics);
    this.outsideObjects.push(gradient);
    this.outsideObjects.push(starGraphics);
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

    // 地板
    const floor = this.add.graphics();
    floor.fillStyle(0x8b7355, 1);
    floor.fillRect(x, y, width, height);
    floor.lineStyle(1, 0x000000, 0.2);
    for (let tx = x; tx < x + width; tx += 30) {
      floor.lineBetween(tx, y, tx, y + height);
    }
    for (let ty = y; ty < y + height; ty += 30) {
      floor.lineBetween(x, ty, x + width, ty);
    }
    floor.setDepth(1);

    // 牆壁群組
    const wallGroup = this.physics.add.staticGroup();
    const wallGraphics = this.add.graphics();
    wallGraphics.setDepth(50);

    const createWall = (wx: number, wy: number, ww: number, wh: number) => {
      // 視覺牆壁
      wallGraphics.fillStyle(0x654321, 1);
      wallGraphics.fillRect(wx, wy, ww, wh);
      wallGraphics.lineStyle(1, 0x000000, 0.5);
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

    // 屋頂（遮蓋建築內部）
    const roof = this.add.graphics();
    roof.fillStyle(0x8b4513, 0.97);
    roof.fillRect(x - 5, y - 5, width + 10, height + 10);
    roof.lineStyle(2, 0x5a2d0a, 1);
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

    // 畫入口標示（黃色箭頭指向門口）
    doorMarker.fillStyle(0xffcc00, 1);

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
    doorMarker.lineStyle(3, 0xffcc00, 0.8);
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

    // 畫門 (棕色)
    door.fillStyle(0x8b4513, 1);
    door.fillRect(doorRectX, doorRectY, doorRectW, doorRectH);
    door.lineStyle(2, 0x5a2d0a, 1);
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

    this.updateStatus(`Indoor - ${building.id}`, "#ffcc00");
  }

  private exitBuilding(): void {
    // 顯示所有屋頂和入口標示
    this.buildings.forEach((b) => {
      b.roof.setVisible(true);
      b.doorMarker.setVisible(true);
    });

    // 隱藏室內遮罩
    this.indoorMask.setVisible(false);

    this.updateStatus("Outdoor", "#4ecca3");
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

  private updateStatus(status: string, color: string): void {
    if (this.onStatusChange) {
      this.onStatusChange(status, color);
    }
  }

  destroy(): void {
    // Clean up subscriptions when scene is destroyed
    if (this.gameStateUnsubscribe) {
      this.gameStateUnsubscribe();
      GameStateLogger.logConnectionStatus("Scene shutting down", "#808080");
    }
  }

  update(): void {
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

    // update player eyes direction
    if (this.player && this.playerEyes) {
      this.drawEyes(this.playerEyes, this.player.x, this.player.y, vx, vy, true);
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

    // smooth movement for other players
    this.otherPlayers.forEach((sprite, playerId) => {
      const target = this.otherPlayersTargets.get(playerId);
      if (target) {
        const prevX = sprite.x;
        const prevY = sprite.y;

        sprite.x = Phaser.Math.Linear(sprite.x, target.x, lerpFactor);
        sprite.y = Phaser.Math.Linear(sprite.y, target.y, lerpFactor);

        // update eyes to look in movement direction
        const eyes = this.otherPlayersEyes.get(playerId);
        if (eyes) {
          const deltaX = target.x - prevX;
          const deltaY = target.y - prevY;
          const length = Math.sqrt(deltaX * deltaX + deltaY * deltaY);

          if (length > 0.5) {
            // normalize to get direction
            const vx = deltaX / length;
            const vy = deltaY / length;
            this.drawEyes(eyes, sprite.x, sprite.y, vx, vy, false);
          } else {
            // not moving, look forward
            this.drawEyes(eyes, sprite.x, sprite.y, 0, 0, false);
          }
        }
      }
    });

    // 檢查是否進入/離開建築
    this.checkBuildingStatus();

    // 檢查寶箱距離，太遠自動關閉（只有跳窗開啟時才檢查）
    if (this.isPopupOpen) {
      this.checkChestDistance();
    }
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
