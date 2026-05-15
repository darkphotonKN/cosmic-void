/**
 * Returns the WebSocket base URL ("<scheme>://<host>"), without a trailing
 * slash or path. Caller appends the path (e.g. "/game/ws").
 *
 * Production: build-time env var NEXT_PUBLIC_WS_URL (e.g. "wss://cosmicvoid.uk")
 *             gets inlined by Next.js at `npm run build`.
 * Dev (no env set): falls back to ws://localhost:5555 — game-service local.
 *
 * This is the only place in the codebase that knows about the WS host.
 * Never hardcode ws://localhost:5555 elsewhere.
 */
export function getWsBaseUrl(): string {
  return process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:5555";
}
