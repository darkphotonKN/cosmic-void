/**
 * GameClient - 遊戲客戶端
 * 負責與後端通信（目前使用 MockBackend 模擬）
 */

import {
  MockBackend,
  VisibleObjects,
  CollectResult,
  AttackResult,
  PlayerData,
  Building,
} from './MockBackend';

export class GameClient {
  private backend: MockBackend;
  public playerId: string;
  public connected: boolean;

  constructor() {
    this.backend = new MockBackend();
    this.playerId = 'player_' + Math.random().toString(36).substring(2, 11);
    this.connected = false;
  }

  connect(): Promise<boolean> {
    return new Promise((resolve) => {
      setTimeout(() => {
        this.connected = true;
        resolve(true);
      }, 500);
    });
  }

  join(x?: number, y?: number): VisibleObjects | null {
    return this.backend.playerJoin(this.playerId, x || 1500, y || 1500);
  }

  updatePosition(x: number, y: number): VisibleObjects | null {
    return this.backend.updatePlayerPosition(this.playerId, x, y);
  }

  collect(treasureId: string): CollectResult {
    return this.backend.collectTreasure(this.playerId, treasureId);
  }

  attack(targetId: string, targetType: 'enemy' | 'player'): AttackResult {
    return this.backend.attack(this.playerId, targetId, targetType);
  }

  getPlayerInfo(): PlayerData | undefined {
    return this.backend.getPlayerInfo(this.playerId);
  }

  getBuildings(): Building[] {
    return this.backend.buildings;
  }

  isInsideBuilding(x: number, y: number, building: Building): boolean {
    return this.backend.isInsideBuilding(x, y, building);
  }
}
