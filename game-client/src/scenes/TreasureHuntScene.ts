/**
 * TreasureHuntScene - 簡化版遊戲場景
 * 移動邏輯 + WebSocket + 建築（進入後看不到外面）
 */

import Phaser from "phaser";
import { ActionMap, ActionType, ClientMessage } from "@/assets/types/client";

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
}

export class TreasureHuntScene extends Phaser.Scene {
  private player!: Phaser.Physics.Arcade.Sprite;

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

  // WebSocket
  private socket!: WebSocket;
  private seq: number = 0;

  // 地圖大小
  private mapWidth = 1200;
  private mapHeight = 800;

  // 建築
  private buildings: Building[] = [];
  private currentBuilding: Building | null = null;
  private outsideObjects: Phaser.GameObjects.GameObject[] = [];
  private indoorMask!: Phaser.GameObjects.Graphics;

  constructor() {
    super({ key: "TreasureHuntScene" });
  }

  setStatusCallback(callback: (status: string, color: string) => void): void {
    this.onStatusChange = callback;
  }

  preload(): void {
    this.createPlayerTexture();
  }

  private createPlayerTexture(): void {
    const graphics = this.make.graphics({});
    graphics.fillStyle(0x4ecca3, 1);
    graphics.fillCircle(20, 20, 18);
    graphics.fillStyle(0xffffff, 1);
    graphics.fillCircle(14, 16, 4);
    graphics.fillCircle(26, 16, 4);
    graphics.generateTexture("player", 40, 40);
    graphics.destroy();
  }

  create(): void {
    // 連接 WebSocket
    this.connectWebSocket();

    // 設置世界邊界
    this.physics.world.setBounds(0, 0, this.mapWidth, this.mapHeight);

    // 創建地圖背景
    this.createMapBackground();

    // 創建建築
    this.createBuildings();

    // 創建玩家（置中）
    this.player = this.physics.add.sprite(
      this.mapWidth / 2,
      this.mapHeight / 2,
      "player",
    );
    this.player.setCollideWorldBounds(true);
    this.player.setDepth(100);

    // 玩家與所有建築牆壁碰撞
    this.buildings.forEach((building) => {
      this.physics.add.collider(this.player, building.wallGroup);
    });

    // 設置相機
    this.cameras.main.setBounds(0, 0, this.mapWidth, this.mapHeight);
    this.cameras.main.startFollow(this.player, true, 0.1, 0.1);

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

    // 創建室內遮罩（用於遮住建築外面）
    this.indoorMask = this.add.graphics();
    this.indoorMask.setDepth(500);
    this.indoorMask.setVisible(false);

    // 顯示座標 UI
    this.createUI();

    this.updateStatus("Connected", "#4ecca3");
  }

  private createMapBackground(): void {
    const graphics = this.add.graphics();

    // 背景色
    graphics.fillStyle(0x1a1a2e, 1);
    graphics.fillRect(0, 0, this.mapWidth, this.mapHeight);

    // 格線
    graphics.lineStyle(1, 0x333355, 0.3);
    for (let x = 0; x <= this.mapWidth; x += 50) {
      graphics.lineBetween(x, 0, x, this.mapHeight);
    }
    for (let y = 0; y <= this.mapHeight; y += 50) {
      graphics.lineBetween(0, y, this.mapWidth, y);
    }

    // 邊界
    graphics.lineStyle(3, 0x4ecca3, 0.8);
    graphics.strokeRect(0, 0, this.mapWidth, this.mapHeight);

    graphics.setDepth(-1);

    // 儲存為室外物件
    this.outsideObjects.push(graphics);
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
    };
  }

  private isPlayerInsideBuilding(building: Building): boolean {
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

  update(): void {
    // 處理移動
    const speed = 200;
    let vx = 0;
    let vy = 0;

    // 計算水平方向
    if (this.cursors.left.isDown || this.wasd.left.isDown) {
      vx = -speed;
    } else if (this.cursors.right.isDown || this.wasd.right.isDown) {
      vx = speed;
    }

    // 計算垂直方向
    if (this.cursors.up.isDown || this.wasd.up.isDown) {
      vy = -speed;
    } else if (this.cursors.down.isDown || this.wasd.down.isDown) {
      vy = speed;
    }

    // 設置速度
    this.player.setVelocity(vx, vy);

    // 發送 WebSocket 訊息（包含八個方位）
    if (vx !== 0 || vy !== 0) {
      this.sendMessage(ActionType.Move, { x: vx, y: vy });
    }

    // 檢查是否進入/離開建築
    this.checkBuildingStatus();
  }

  private connectWebSocket(): void {
    this.socket = new WebSocket("ws://localhost:5555/game/ws");

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
}
