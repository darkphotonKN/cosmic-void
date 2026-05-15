# =============================================================================
# game-client (Next.js 15 + Phaser 3) Dockerfile
# Build context = game-client/
#   cd game-client
#   docker build --platform linux/amd64 -t game-client:dev .
# =============================================================================

# ----- Stage 1: install dependencies (cached layer) -----
FROM node:24.15.0-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --legacy-peer-deps

# ----- Stage 2: build the standalone output -----
FROM node:24.15.0-alpine AS builder
WORKDIR /app

# NEXT_PUBLIC_* env vars are inlined into the JS bundle at build time, so they
# must be set BEFORE `npm run build`. Override at build time:
#   docker buildx build --build-arg NEXT_PUBLIC_API_URL=https://api.cosmicvoid.uk ...
ARG NEXT_PUBLIC_API_URL=https://api.cosmicvoid.uk
ARG NEXT_PUBLIC_WS_URL=wss://cosmicvoid.uk
ARG NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=""
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_WS_URL=$NEXT_PUBLIC_WS_URL
ENV NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=$NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
ENV NEXT_TELEMETRY_DISABLED=1

COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN npm run build

# ----- Stage 3: runtime (minimal) -----
FROM node:24.15.0-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

# Non-root user (security best practice)
RUN addgroup -S nodejs && adduser -S nextjs -G nodejs

# Copy the slim standalone build output
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

CMD ["node", "server.js"]
