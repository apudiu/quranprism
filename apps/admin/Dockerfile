# syntax=docker/dockerfile:1.7
# Production image for @qp/admin (SolidStart 2 / nitro SSR).
# Build context is the REPO ROOT: `docker build -f apps/admin/Dockerfile .`

# ---- build stage: Bun installs the workspace graph and runs the Vite build ----
FROM oven/bun:1.3.14-alpine AS build
WORKDIR /app

# Manifests + lockfile first so the dependency layer caches across source edits.
COPY package.json bun.lock bunfig.toml tsconfig.base.json ./
COPY apps/web/package.json ./apps/web/
COPY apps/admin/package.json ./apps/admin/
COPY pkg/ui/package.json ./pkg/ui/
COPY pkg/http/package.json ./pkg/http/
RUN bun install --frozen-lockfile

# Source for the app and the workspace packages it bundles (@qp/ui, @qp/http).
COPY pkg/ui ./pkg/ui
COPY pkg/http ./pkg/http
COPY apps/admin ./apps/admin
RUN bun run --filter @qp/admin build

# ---- runtime stage: nitro bundles all server deps into .output, so bare Node works ----
FROM node:24-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/apps/admin/.output ./.output
EXPOSE 3000
USER node
CMD ["node", ".output/server/index.mjs"]
