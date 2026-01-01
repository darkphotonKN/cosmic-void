import { ActionMap, PlayerSessionPayload } from '@/assets/types/client';

/**
 * GameSessionManager - Manages session and player IDs for game messages
 *
 * This class stores the current session ID and player ID that are required
 * for all game action messages sent to the backend.
 */
class GameSessionManager {
  private sessionId: string | null = null;
  private playerId: string | null = null;

  /**
   * Set the session ID (received from server when joining a game)
   */
  setSessionId(sessionId: string): void {
    this.sessionId = sessionId;
    console.log(`[GameSession] Session ID set: ${sessionId}`);
  }

  /**
   * Set the player ID (should be the user's ID from authentication)
   */
  setPlayerId(playerId: string): void {
    this.playerId = playerId;
    console.log(`[GameSession] Player ID set: ${playerId}`);
  }

  /**
   * Get current session ID
   */
  getSessionId(): string | null {
    return this.sessionId;
  }

  /**
   * Get current player ID
   */
  getPlayerId(): string | null {
    return this.playerId;
  }

  /**
   * Check if session is ready (both IDs are set)
   */
  isReady(): boolean {
    return this.sessionId !== null && this.playerId !== null;
  }

  /**
   * Clear session data (for logout/disconnect)
   */
  clear(): void {
    this.sessionId = null;
    this.playerId = null;
    console.log('[GameSession] Session data cleared');
  }

  /**
   * Create a base payload with session and player IDs
   * Throws error if session is not ready
   */
  createBasePayload(): PlayerSessionPayload {
    if (!this.isReady()) {
      throw new Error('GameSession not ready: sessionId or playerId is missing');
    }

    return {
      session_id: this.sessionId!,
      player_id: this.playerId!
    };
  }

  /**
   * Helper to create a complete game action payload
   * Merges the base session payload with action-specific data
   */
  createGamePayload<T extends keyof ActionMap>(
    actionData: Omit<ActionMap[T], keyof PlayerSessionPayload>
  ): ActionMap[T] {
    const basePayload = this.createBasePayload();
    return {
      ...basePayload,
      ...actionData
    } as ActionMap[T];
  }

  /**
   * Create a move payload with proper format
   */
  createMovePayload(vx: number, vy: number): ActionMap['move'] {
    return this.createGamePayload<'move'>({ vx, vy });
  }

  /**
   * Create an interact payload
   */
  createInteractPayload(entityId: string): ActionMap['interact'] {
    return this.createGamePayload<'interact'>({ entity_id: entityId });
  }

  /**
   * Create an attack payload
   */
  createAttackPayload(targetId: string): ActionMap['attack'] {
    return this.createGamePayload<'attack'>({ target_id: targetId });
  }

  /**
   * Create a pickup payload
   */
  createPickupPayload(itemId: string): ActionMap['pickup'] {
    return this.createGamePayload<'pickup'>({ item_id: itemId });
  }

  /**
   * Create a use item payload
   */
  createUsePayload(itemId: string, targetId?: string): ActionMap['use'] {
    return this.createGamePayload<'use'>({ item_id: itemId, target_id: targetId });
  }

  /**
   * Create a chat payload
   */
  createChatPayload(message: string): ActionMap['chat'] {
    return this.createGamePayload<'chat'>({ message });
  }
}

// Export singleton instance
export const gameSession = new GameSessionManager();

// Export type for testing or extending
export type { GameSessionManager };