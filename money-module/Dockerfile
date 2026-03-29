# syntax=docker/dockerfile:1
# language which we use
FROM oven/bun:1 AS base
WORKDIR /money-module

FROM base AS deps
COPY money-module/package.json ./
RUN bun install

FROM base AS dev
COPY --from=deps /money-module/node_modules ./node_modules
COPY money-module/ .
EXPOSE 6767 
CMD ["bun", "run", "index.ts"]
