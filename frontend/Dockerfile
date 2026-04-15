# syntax=docker/dockerfile:1
FROM oven/bun:1 AS base
# Set your preferred working directory
WORKDIR /frontend

# --- DEPENDENCIES STAGE ---
FROM base AS deps
COPY package.json bun.lock* ./
RUN bun install --frozen-lockfile

# --- DEV STAGE ---
FROM base AS dev
# Notice we pull from /frontend/node_modules now
COPY --from=deps /frontend/node_modules ./node_modules
# Copy local files into the current WORKDIR (/frontend)
COPY . . 
EXPOSE 3000 
CMD ["bun", "run", "dev", "--host", "0.0.0.0"]

# --- BUILDER STAGE ---
FROM base AS builder
COPY --from=deps /frontend/node_modules ./node_modules
COPY . .
RUN bun run build

# --- PRODUCTION STAGE ---
FROM base AS production
COPY --from=builder /frontend/.output ./.output
COPY --from=builder /frontend/node_modules ./node_modules
COPY package.json ./
EXPOSE 3000
CMD ["bun", "run", "preview", "--host", "0.0.0.0"]
