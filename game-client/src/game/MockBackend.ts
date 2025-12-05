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

// TODO: 建築系統暫時移除
// interface BuildingConfig { ... }

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

    this.mapWidth = 900;
    this.mapHeight = 600;
    this.viewRadius = 300; // 大於地圖，簡化後可看到全部

    this.initializeWorld();
  }

  private initializeWorld(): void {
    // 簡化版本：不生成建築

    // 生成少量寶箱
    for (let i = 0; i < 5; i++) {
      const x = 50 + Math.random() * (this.mapWidth - 100);
      const y = 50 + Math.random() * (this.mapHeight - 100);

      this.worldObjects.treasures.push({
        id: `treasure_${i}`,
        x,
        y,
        type: Math.random() > 0.5 ? 'gold' : 'silver',
        value: Math.random() > 0.5 ? 100 : 50,
        collected: false,
        isIndoor: false,
      });
    }

    // 生成少量敵人
    for (let i = 0; i < 3; i++) {
      const x = 50 + Math.random() * (this.mapWidth - 100);
      const y = 50 + Math.random() * (this.mapHeight - 100);

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

  // TODO: 建築系統暫時移除，未來加入後端後再啟用
  // generateBuildings, getRandomInteriorPosition, generateWallsForBuilding, addPartitions, isInsideAnyBuilding
  // 已暫時移除，保留在 git 歷史記錄中

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
