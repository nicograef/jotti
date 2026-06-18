# Build stage: render the Astro + Starlight static site.
#
# The build context is the repository root (not website/), because the docs are
# the single source of truth: Astro's glob loader reads the published Markdown
# directly from the top-level docs/ via the relative path ../docs. Both website/
# and docs/ are copied in; everything else is excluded by .dockerignore.
FROM node:24-alpine AS build

WORKDIR /app

# Install dependencies with pnpm via Corepack (respects pnpm-lock.yaml)
COPY website/package.json website/pnpm-lock.yaml website/pnpm-workspace.yaml ./
RUN corepack enable \
  && corepack prepare pnpm@11 --activate \
  && pnpm install --frozen-lockfile

# Astro reads docs/ as ../docs relative to the project root (/app) — copy both
COPY website/ ./
COPY docs/ /docs/
RUN pnpm build

# Runtime stage: serve the static build with nginx
FROM nginx:1.27-alpine AS runner

COPY --from=build /app/dist /usr/share/nginx/html
COPY website/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
