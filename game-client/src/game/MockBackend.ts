/**
 * MockBackend - 模擬後端邏輯
 * 實際專案中，這部分應該是真正的 WebSocket Server
 */

export interface WorldObjects {
  treasures: Treasure[];
  enemies: Enemy[];
  buildings: Building[];
  walls: Wall[];
  interiors: Interior[];
  players: Map<string, PlayerData>;
}

export interface Treasure {
  id: string;
  x: number;
  y: number;
  type: 'gold' | 'silver';
  value: number;
  collected: boolean;
  isIndoor: boolean;
  buildingId?: string;
}

export interface Enemy {
  id: string;
  x: number;
  y: number;
  type: 'skeleton' | 'goblin';
  hp: number;
  alive: boolean;
  isIndoor: boolean;
  buildingId?: string;
}

export interface Wall {
  x: number;
  y: number;
  width: number;
  height: number;
  buildingId?: string;
}

export interface Interior {
  id: string;
  x: number;
  y: number;
}

export interface Door {
  x: number;
  y: number;
  width: number;
  rotation: number;
  side: 'top' | 'bottom' | 'left' | 'right' | 'none';
}

export interface Building {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
  type: 'house' | 'tower' | 'ruins' | 'shrine';
  doorSide: 'top' | 'bottom' | 'left' | 'right' | 'none';
  hasRoof: boolean;
  hasPartition: boolean;
  walls: Wall[];
  door: Door | null;
  playerInside?: boolean;
}

export interface PlayerData {
  id: string;
  x: number;
  y: number;
  hp: number;
  inventory: Treasure[];
  score: number;
}

export interface VisibleObjects {
  treasures: Treasure[];
  enemies: Enemy[];
  buildings: Building[];
  walls: Wall[];
  players: PlayerData[];
  playerInside: string | null;
}

export interface CollectResult {
  success: boolean;
  treasure?: Treasure;
  newScore?: number;
}

export interface AttackResult {
  success: boolean;
  killed?: boolean;
  newScore?: number;
  enemyHp?: number;
  targetHp?: number;
}

interface BuildingConfig {
  type: 'house' | 'tower' | 'ruins' | 'shrine';
  width: number;
  height: number;
  doorSide: 'top' | 'bottom' | 'left' | 'right' | 'none';
  hasPartition: boolean;
  hasRoof?: boolean;
}

export class MockBackend {
  private worldObjects: WorldObjects;
  private mapWidth: number;
  private mapHeight: number;
  private viewRadius: number;

  constructor() {
    this.worldObjects = {
      treasures: [],
      enemies: [],
      buildings: [],
      walls: [],
      interiors: [],
      players: new Map(),
    };

    this.mapWidth = 3000;
    this.mapHeight = 3000;
    this.viewRadius = 250;

    this.initializeWorld();
  }

  private initializeWorld(): void {
    this.generateBuildings();

    // 生成戶外寶箱
    for (let i = 0; i < 20; i++) {
      let x: number,
        y: number,
        attempts = 0;
      do {
        x = 100 + Math.random() * (this.mapWidth - 200);
        y = 100 + Math.random() * (this.mapHeight - 200);
        attempts++;
      } while (this.isInsideAnyBuilding(x, y) && attempts < 50);

      this.worldObjects.treasures.push({
        id: `treasure_${i}`,
        x,
        y,
        type: Math.random() > 0.7 ? 'gold' : 'silver',
        value: Math.random() > 0.7 ? 100 : 50,
        collected: false,
        isIndoor: false,
      });
    }

    // 生成戶外敵人
    for (let i = 0; i < 10; i++) {
      let x: number,
        y: number,
        attempts = 0;
      do {
        x = 100 + Math.random() * (this.mapWidth - 200);
        y = 100 + Math.random() * (this.mapHeight - 200);
        attempts++;
      } while (this.isInsideAnyBuilding(x, y) && attempts < 50);

      this.worldObjects.enemies.push({
        id: `enemy_${i}`,
        x,
        y,
        type: Math.random() > 0.5 ? 'skeleton' : 'goblin',
        hp: 100,
        alive: true,
        isIndoor: false,
      });
    }
  }

  private generateBuildings(): void {
    const buildingConfigs: BuildingConfig[] = [
      { type: 'house', width: 280, height: 220, doorSide: 'bottom', hasPartition: true },
      { type: 'house', width: 320, height: 240, doorSide: 'left', hasPartition: true },
      { type: 'house', width: 260, height: 200, doorSide: 'right', hasPartition: true },
      { type: 'house', width: 300, height: 260, doorSide: 'top', hasPartition: true },
      { type: 'house', width: 200, height: 160, doorSide: 'bottom', hasPartition: false },
      { type: 'house', width: 180, height: 180, doorSide: 'left', hasPartition: false },
      { type: 'tower', width: 140, height: 140, doorSide: 'bottom', hasPartition: false },
      { type: 'tower', width: 160, height: 160, doorSide: 'right', hasPartition: true },
      { type: 'ruins', width: 200, height: 160, doorSide: 'none', hasPartition: false, hasRoof: false },
      { type: 'ruins', width: 180, height: 140, doorSide: 'none', hasPartition: false, hasRoof: false },
      { type: 'shrine', width: 220, height: 180, doorSide: 'bottom', hasPartition: true },
      { type: 'shrine', width: 200, height: 200, doorSide: 'top', hasPartition: false },
    ];

    const gridSize = 450;
    const positions: { x: number; y: number }[] = [];

    for (let gx = 1; gx < this.mapWidth / gridSize - 1; gx++) {
      for (let gy = 1; gy < this.mapHeight / gridSize - 1; gy++) {
        positions.push({
          x: gx * gridSize + (Math.random() - 0.5) * 80,
          y: gy * gridSize + (Math.random() - 0.5) * 80,
        });
      }
    }

    positions.sort(() => Math.random() - 0.5);

    const numBuildings = Math.min(buildingConfigs.length, positions.length);

    for (let i = 0; i < numBuildings; i++) {
      const config = buildingConfigs[i];
      const pos = positions[i];

      const building: Building = {
        id: `building_${i}`,
        x: pos.x,
        y: pos.y,
        width: config.width,
        height: config.height,
        type: config.type,
        doorSide: config.doorSide,
        hasRoof: config.hasRoof !== false,
        hasPartition: config.hasPartition || false,
        walls: [],
        door: null,
      };

      this.generateWallsForBuilding(building);
      this.worldObjects.buildings.push(building);

      // 在建築內部放置物品
      if (config.hasRoof !== false && Math.random() > 0.3) {
        const numTreasures = config.hasPartition ? 2 : 1;
        for (let t = 0; t < numTreasures; t++) {
          const treasurePos = this.getRandomInteriorPosition(building);
          this.worldObjects.treasures.push({
            id: `indoor_treasure_${i}_${t}`,
            x: treasurePos.x,
            y: treasurePos.y,
            type: 'gold',
            value: 150 + Math.floor(Math.random() * 100),
            collected: false,
            isIndoor: true,
            buildingId: building.id,
          });
        }
      }

      // 室內敵人
      if (config.hasRoof !== false && Math.random() > 0.5) {
        const enemyPos = this.getRandomInteriorPosition(building);
        this.worldObjects.enemies.push({
          id: `indoor_enemy_${i}`,
          x: enemyPos.x,
          y: enemyPos.y,
          type: Math.random() > 0.5 ? 'skeleton' : 'goblin',
          hp: 100,
          alive: true,
          isIndoor: true,
          buildingId: building.id,
        });
      }
    }
  }

  private getRandomInteriorPosition(building: Building): { x: number; y: number } {
    const margin = 40;
    return {
      x: building.x + (Math.random() - 0.5) * (building.width - margin * 2),
      y: building.y + (Math.random() - 0.5) * (building.height - margin * 2),
    };
  }

  private generateWallsForBuilding(building: Building): void {
    const { x, y, width, height, doorSide, type, hasPartition } = building;
    const wallThickness = 14;
    const doorWidth = 60;

    const halfW = width / 2;
    const halfH = height / 2;

    // 廢墟只有部分牆壁
    if (type === 'ruins') {
      building.walls.push({
        x: x - halfW,
        y: y - halfH,
        width: wallThickness,
        height: height * 0.6,
      });
      building.walls.push({
        x: x + halfW - wallThickness,
        y: y,
        width: wallThickness,
        height: height * 0.5,
      });
      building.walls.push({
        x: x - halfW,
        y: y + halfH - wallThickness,
        width: width * 0.4,
        height: wallThickness,
      });
      return;
    }

    let doorX = 0,
      doorY = 0,
      doorRotation = 0;

    // 上牆
    if (doorSide !== 'top') {
      building.walls.push({
        x: x - halfW,
        y: y - halfH,
        width: width,
        height: wallThickness,
      });
    } else {
      const sideWidth = (width - doorWidth) / 2;
      building.walls.push({
        x: x - halfW,
        y: y - halfH,
        width: sideWidth,
        height: wallThickness,
      });
      building.walls.push({
        x: x + doorWidth / 2,
        y: y - halfH,
        width: sideWidth,
        height: wallThickness,
      });
      doorX = x;
      doorY = y - halfH;
      doorRotation = 0;
    }

    // 下牆
    if (doorSide !== 'bottom') {
      building.walls.push({
        x: x - halfW,
        y: y + halfH - wallThickness,
        width: width,
        height: wallThickness,
      });
    } else {
      const sideWidth = (width - doorWidth) / 2;
      building.walls.push({
        x: x - halfW,
        y: y + halfH - wallThickness,
        width: sideWidth,
        height: wallThickness,
      });
      building.walls.push({
        x: x + doorWidth / 2,
        y: y + halfH - wallThickness,
        width: sideWidth,
        height: wallThickness,
      });
      doorX = x;
      doorY = y + halfH;
      doorRotation = 180;
    }

    // 左牆
    if (doorSide !== 'left') {
      building.walls.push({
        x: x - halfW,
        y: y - halfH,
        width: wallThickness,
        height: height,
      });
    } else {
      const sideHeight = (height - doorWidth) / 2;
      building.walls.push({
        x: x - halfW,
        y: y - halfH,
        width: wallThickness,
        height: sideHeight,
      });
      building.walls.push({
        x: x - halfW,
        y: y + doorWidth / 2,
        width: wallThickness,
        height: sideHeight,
      });
      doorX = x - halfW;
      doorY = y;
      doorRotation = 270;
    }

    // 右牆
    if (doorSide !== 'right') {
      building.walls.push({
        x: x + halfW - wallThickness,
        y: y - halfH,
        width: wallThickness,
        height: height,
      });
    } else {
      const sideHeight = (height - doorWidth) / 2;
      building.walls.push({
        x: x + halfW - wallThickness,
        y: y - halfH,
        width: wallThickness,
        height: sideHeight,
      });
      building.walls.push({
        x: x + halfW - wallThickness,
        y: y + doorWidth / 2,
        width: wallThickness,
        height: sideHeight,
      });
      doorX = x + halfW;
      doorY = y;
      doorRotation = 90;
    }

    building.door = {
      x: doorX,
      y: doorY,
      width: doorWidth,
      rotation: doorRotation,
      side: doorSide,
    };

    if (hasPartition) {
      this.addPartitions(building);
    }
  }

  private addPartitions(building: Building): void {
    const { x, y, width, height } = building;
    const wallThickness = 10;
    const partitionDoorWidth = 45;

    const halfW = width / 2;
    const halfH = height / 2;

    if (width > height) {
      const partitionX = x + (Math.random() - 0.5) * (width * 0.3);

      building.walls.push({
        x: partitionX - wallThickness / 2,
        y: y - halfH + 14,
        width: wallThickness,
        height: (height - partitionDoorWidth) / 2 - 14,
      });

      building.walls.push({
        x: partitionX - wallThickness / 2,
        y: y + partitionDoorWidth / 2,
        width: wallThickness,
        height: (height - partitionDoorWidth) / 2 - 14,
      });
    } else {
      const partitionY = y + (Math.random() - 0.5) * (height * 0.3);

      building.walls.push({
        x: x - halfW + 14,
        y: partitionY - wallThickness / 2,
        width: (width - partitionDoorWidth) / 2 - 14,
        height: wallThickness,
      });

      building.walls.push({
        x: x + partitionDoorWidth / 2,
        y: partitionY - wallThickness / 2,
        width: (width - partitionDoorWidth) / 2 - 14,
        height: wallThickness,
      });
    }
  }

  private isInsideAnyBuilding(x: number, y: number): boolean {
    return this.worldObjects.buildings.some((b) => this.isInsideBuilding(x, y, b));
  }

  isInsideBuilding(x: number, y: number, building: Building): boolean {
    const halfW = building.width / 2;
    const halfH = building.height / 2;
    return (
      x >= building.x - halfW &&
      x <= building.x + halfW &&
      y >= building.y - halfH &&
      y <= building.y + halfH
    );
  }

  private hasLineOfSight(x1: number, y1: number, x2: number, y2: number): boolean {
    for (const building of this.worldObjects.buildings) {
      if (!building.hasRoof) continue;

      for (const wall of building.walls) {
        if (this.lineIntersectsRect(x1, y1, x2, y2, wall)) {
          return false;
        }
      }
    }
    return true;
  }

  private lineIntersectsRect(
    x1: number,
    y1: number,
    x2: number,
    y2: number,
    rect: Wall,
  ): boolean {
    const left = rect.x;
    const right = rect.x + rect.width;
    const top = rect.y;
    const bottom = rect.y + rect.height;

    return (
      this.lineIntersectsLine(x1, y1, x2, y2, left, top, right, top) ||
      this.lineIntersectsLine(x1, y1, x2, y2, right, top, right, bottom) ||
      this.lineIntersectsLine(x1, y1, x2, y2, left, bottom, right, bottom) ||
      this.lineIntersectsLine(x1, y1, x2, y2, left, top, left, bottom)
    );
  }

  private lineIntersectsLine(
    x1: number,
    y1: number,
    x2: number,
    y2: number,
    x3: number,
    y3: number,
    x4: number,
    y4: number,
  ): boolean {
    const denom = (y4 - y3) * (x2 - x1) - (x4 - x3) * (y2 - y1);
    if (Math.abs(denom) < 0.0001) return false;

    const ua = ((x4 - x3) * (y1 - y3) - (y4 - y3) * (x1 - x3)) / denom;
    const ub = ((x2 - x1) * (y1 - y3) - (y2 - y1) * (x1 - x3)) / denom;

    return ua >= 0 && ua <= 1 && ub >= 0 && ub <= 1;
  }

  playerJoin(playerId: string, x?: number, y?: number): VisibleObjects | null {
    this.worldObjects.players.set(playerId, {
      id: playerId,
      x: x || 1500,
      y: y || 1500,
      hp: 100,
      inventory: [],
      score: 0,
    });
    return this.getVisibleObjects(playerId);
  }

  updatePlayerPosition(playerId: string, x: number, y: number): VisibleObjects | null {
    const player = this.worldObjects.players.get(playerId);
    if (player) {
      player.x = x;
      player.y = y;
    }
    return this.getVisibleObjects(playerId);
  }

  getVisibleObjects(playerId: string): VisibleObjects | null {
    const player = this.worldObjects.players.get(playerId);
    if (!player) return null;

    const visible: VisibleObjects = {
      treasures: [],
      enemies: [],
      buildings: [],
      walls: [],
      players: [],
      playerInside: null,
    };

    const isInRange = (obj: { x: number; y: number }): boolean => {
      const dx = obj.x - player.x;
      const dy = obj.y - player.y;
      return Math.sqrt(dx * dx + dy * dy) <= this.viewRadius;
    };

    const playerBuilding = this.worldObjects.buildings.find(
      (b) => b.hasRoof && this.isInsideBuilding(player.x, player.y, b),
    );

    if (playerBuilding) {
      visible.playerInside = playerBuilding.id;
    }

    // 玩家在室內
    if (playerBuilding) {
      visible.treasures = this.worldObjects.treasures.filter((t) => {
        if (t.collected) return false;
        if (t.isIndoor && t.buildingId === playerBuilding.id) {
          return true;
        }
        return false;
      });

      visible.enemies = this.worldObjects.enemies.filter((e) => {
        if (!e.alive) return false;
        if (e.isIndoor && e.buildingId === playerBuilding.id) {
          return true;
        }
        return false;
      });

      this.worldObjects.players.forEach((p, id) => {
        if (id !== playerId) {
          if (this.isInsideBuilding(p.x, p.y, playerBuilding)) {
            visible.players.push({
              id: p.id,
              x: p.x,
              y: p.y,
              hp: p.hp,
              inventory: p.inventory,
              score: p.score,
            });
          }
        }
      });

      visible.buildings = [
        {
          ...playerBuilding,
          playerInside: true,
        },
      ];

      playerBuilding.walls.forEach((wall) => {
        visible.walls.push({
          ...wall,
          buildingId: playerBuilding.id,
        });
      });
    } else {
      // 玩家在室外
      visible.treasures = this.worldObjects.treasures.filter((t) => {
        if (t.collected) return false;
        if (!isInRange(t)) return false;
        if (t.isIndoor) return false;
        return this.hasLineOfSight(player.x, player.y, t.x, t.y);
      });

      visible.enemies = this.worldObjects.enemies.filter((e) => {
        if (!e.alive) return false;
        if (!isInRange(e)) return false;
        if (e.isIndoor) return false;
        return this.hasLineOfSight(player.x, player.y, e.x, e.y);
      });

      visible.buildings = this.worldObjects.buildings
        .filter((b) => {
          const dx = b.x - player.x;
          const dy = b.y - player.y;
          return Math.sqrt(dx * dx + dy * dy) <= this.viewRadius + 100;
        })
        .map((b) => ({
          ...b,
          playerInside: false,
        }));

      visible.buildings.forEach((b) => {
        b.walls.forEach((wall) => {
          visible.walls.push({
            ...wall,
            buildingId: b.id,
          });
        });
      });

      this.worldObjects.players.forEach((p, id) => {
        if (id !== playerId && isInRange(p)) {
          const otherPlayerBuilding = this.worldObjects.buildings.find(
            (b) => b.hasRoof && this.isInsideBuilding(p.x, p.y, b),
          );

          if (!otherPlayerBuilding && this.hasLineOfSight(player.x, player.y, p.x, p.y)) {
            visible.players.push({
              id: p.id,
              x: p.x,
              y: p.y,
              hp: p.hp,
              inventory: p.inventory,
              score: p.score,
            });
          }
        }
      });
    }

    return visible;
  }

  collectTreasure(playerId: string, treasureId: string): CollectResult {
    const player = this.worldObjects.players.get(playerId);
    const treasure = this.worldObjects.treasures.find((t) => t.id === treasureId);

    if (player && treasure && !treasure.collected) {
      const dx = treasure.x - player.x;
      const dy = treasure.y - player.y;
      const distance = Math.sqrt(dx * dx + dy * dy);

      if (distance <= 50) {
        treasure.collected = true;
        player.score += treasure.value;
        player.inventory.push(treasure);
        return { success: true, treasure, newScore: player.score };
      }
    }
    return { success: false };
  }

  attack(playerId: string, targetId: string, targetType: 'enemy' | 'player'): AttackResult {
    const player = this.worldObjects.players.get(playerId);
    if (!player) return { success: false };

    if (targetType === 'enemy') {
      const enemy = this.worldObjects.enemies.find((e) => e.id === targetId);
      if (enemy && enemy.alive) {
        const dx = enemy.x - player.x;
        const dy = enemy.y - player.y;
        const distance = Math.sqrt(dx * dx + dy * dy);

        if (distance <= 60) {
          enemy.hp -= 25;
          if (enemy.hp <= 0) {
            enemy.alive = false;
            player.score += 50;
            return { success: true, killed: true, newScore: player.score };
          }
          return { success: true, killed: false, enemyHp: enemy.hp };
        }
      }
    } else if (targetType === 'player') {
      const target = this.worldObjects.players.get(targetId);
      if (target) {
        const dx = target.x - player.x;
        const dy = target.y - player.y;
        const distance = Math.sqrt(dx * dx + dy * dy);

        if (distance <= 60) {
          target.hp -= 20;
          return { success: true, targetHp: target.hp };
        }
      }
    }
    return { success: false };
  }

  getPlayerInfo(playerId: string): PlayerData | undefined {
    return this.worldObjects.players.get(playerId);
  }

  get buildings(): Building[] {
    return this.worldObjects.buildings;
  }
}
