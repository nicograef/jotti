# Build context is the repository root, not website/: the docs are the single
# source of truth and Astro reads them from the top-level docs/. .dockerignore
# excludes everything but website/ and docs/.
FROM node:24-alpine AS build

WORKDIR /app

COPY website/package.json website/pnpm-lock.yaml website/pnpm-workspace.yaml ./
RUN corepack enable \
  && corepack prepare pnpm@11 --activate \
  && pnpm install --frozen-lockfile

# Astro reads docs/ as ../docs relative to the project root (/app) — copy both
COPY website/ ./
COPY docs/ /docs/
RUN pnpm build

FROM nginx:1.31-alpine AS runner

COPY --from=build /app/dist /usr/share/nginx/html
COPY website/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
