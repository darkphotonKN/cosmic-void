import { ClientGameState, formatPosition, formatVelocity } from '@/types/gameState';

/**
 * Log levels — same concept as slog on the backend.
 *
 *   0 = silent   (nothing)
 *   1 = error     (errors only)
 *   2 = info      (connections, equip actions, one-off events)
 *   3 = debug     (game state ticks — the 30/s firehose)
 *
 * Toggle at runtime in the browser console:
 *   GameStateLogger.setLevel(2)   // see actions, hide ticks
 *   GameStateLogger.setLevel(3)   // full firehose
 *   GameStateLogger.setLevel(0)   // silence everything
 */
export class GameStateLogger {
  private static updateCount = 0;
  private static lastUpdateTime = Date.now();
  private static updateInterval = 0;
  private static level = 2; // default: info (no tick spam)

  static setLevel(lvl: number): void {
    this.level = lvl;
    console.log(`%c[Logger] level set to ${lvl}`, 'color: #4ecca3; font-weight: bold');
  }

  static getLevel(): number { return this.level; }

  static logGameState(state: ClientGameState): void {
    if (this.level < 3) return; // skip unless debug level

    const timestamp = new Date().toLocaleTimeString();
    const now = Date.now();

    // Calculate time since last update
    if (this.lastUpdateTime) {
      this.updateInterval = now - this.lastUpdateTime;
    }
    this.lastUpdateTime = now;
    this.updateCount++;

    // Main header with update count and timing
    console.group(
      `%c[Game State Update #${this.updateCount}] - ${timestamp} (${this.updateInterval}ms)`,
      'color: #4ecca3; font-weight: bold; font-size: 14px'
    );

    // Session info
    console.log('%c📍 Session ID:', 'color: #ffd700; font-weight: bold', state.session_id);

    // Current player info (highlighted)
    if (state.current_player) {
      console.group('%c👤 Current Player (YOU)', 'color: #00ff00; font-weight: bold; font-size: 12px');
      console.table({
        'Username': state.current_player.username,
        'Player ID': state.current_player.id,
        'Entity ID': state.current_player.entity_id,
        'Position': formatPosition(state.current_player.position),
        'Velocity': formatVelocity(state.current_player.direction),
        'Speed': state.current_player.direction.speed.toFixed(2)
      });
      console.groupEnd();
    } else {
      console.warn('%c⚠️ No current player data received!', 'color: #ff6b6b; font-weight: bold');
    }

    // Other players
    if (state.other_players && state.other_players.length > 0) {
      console.group(`%c👥 Other Players (${state.other_players.length})`, 'color: #00bfff; font-weight: bold');

      const otherPlayersData = state.other_players.map(player => ({
        'Username': player.username,
        'Position': formatPosition(player.position),
        'Velocity': formatVelocity(player.direction),
        'Speed': player.direction.speed.toFixed(2)
      }));

      console.table(otherPlayersData);
      console.groupEnd();
    } else {
      console.log('%c👥 No other players in session', 'color: #808080; font-style: italic');
    }

    // Game objects summary
    console.group('%c🎮 Game Objects', 'color: #dda0dd; font-weight: bold');
    console.log(`📦 Items: ${state.items?.length || 0}`);
    if (state.items && state.items.length > 0) {
      console.log('  Items:', state.items);
    }

    console.log(`🚪 Doors: ${state.doors?.length || 0}`);
    if (state.doors && state.doors.length > 0) {
      const doorsData = state.doors.map(door => ({
        'Entity ID': door.entity_id,
        'Position': formatPosition(door.position),
        'Status': door.is_open ? '🟢 Open' : '🔴 Closed'
      }));
      console.table(doorsData);
    }
    console.groupEnd();

    // Full state for debugging (collapsed by default)
    console.group('%c📊 Full State Object (Debug)', 'color: #808080; font-size: 10px');
    console.log(JSON.stringify(state, null, 2));
    console.groupEnd();

    // Close main group
    console.groupEnd();
  }

  // Helper method to log connection status
  static logConnectionStatus(status: string, color: string = '#4ecca3'): void {
    console.log(
      `%c[WebSocket] ${status}`,
      `color: ${color}; font-weight: bold; padding: 2px 6px; background: #1a1a2e; border-left: 3px solid ${color}`
    );
  }

  // Helper method to log errors
  static logError(message: string, error?: any): void {
    console.group('%c❌ Error', 'color: #ff4444; font-weight: bold');
    console.error(message);
    if (error) {
      console.error('Details:', error);
    }
    console.groupEnd();
  }

  // Reset counter (useful for debugging)
  static reset(): void {
    this.updateCount = 0;
    this.lastUpdateTime = Date.now();
    console.log('%c🔄 Game state logger reset', 'color: #4ecca3; font-style: italic');
  }
}

// Expose to browser console for runtime toggling:
//   GameStateLogger.setLevel(3)  → full firehose
//   GameStateLogger.setLevel(2)  → actions only (default)
//   GameStateLogger.setLevel(0)  → silent
if (typeof window !== 'undefined') {
  (window as any).GameStateLogger = GameStateLogger;
}