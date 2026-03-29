# syntax=docker/dockerfile:1
FROM oven/bun:1 AS base
WORKDIR /app

FROM base AS deps
COPY app/package.json app/bun.lock ./
RUN bun install --frozen-lockfile

# --- DEV STAGE ---
FROM base AS dev
COPY --from=deps /app/node_modules ./node_modules
COPY app/ .
EXPOSE 3000 
CMD ["bun", "run", "dev", "--host", "0.0.0.0"]

# --- PRODUCTION STAGE ---
FROM base AS builder
COPY --from=deps /app/node_modules ./node_modules
COPY app/ .
RUN bun run build

FROM base AS production
COPY --from=builder /app/.output ./.output
COPY --from=builder /app/node_modules ./node_modules
COPY app/package.json ./
EXPOSE 3000
CMD ["bun", "run", "preview", "--host", "0.0.0.0"]
