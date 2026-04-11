# syntax=docker/dockerfile:1
FROM oven/bun:1 AS base
WORKDIR /frontend

FROM base AS deps
COPY frontend/package.json /bun.lock ./frontend
RUN bun install --frozen-lockfile

# --- DEV STAGE ---
FROM base AS dev
COPY --from=deps /frontend/node_modules ./node_modules
COPY frontend/ .
EXPOSE 3000 
CMD ["bun", "run", "dev", "--host", "0.0.0.0"]

# --- PRODUCTION STAGE ---
FROM base AS builder
COPY --from=deps /frontend/node_modules ./node_modules
COPY frontend/ .
RUN bun run build

FROM base AS production
COPY --from=builder /frontend/.output ./.output
COPY --from=builder /frontend/node_modules ./node_modules
COPY frontend/package.json ./
EXPOSE 3000
CMD ["bun", "run", "preview", "--host", "0.0.0.0"]
