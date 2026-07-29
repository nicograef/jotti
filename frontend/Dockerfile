# Build stage: compile the React (Vite) app
FROM node:24-alpine AS build

WORKDIR /app

# Install dependencies with pnpm via Corepack (respects pnpm-lock.yaml)
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN corepack enable \
  && corepack prepare pnpm@11 --activate \
  && pnpm install --frozen-lockfile

# Copy the rest of the sources and build
COPY . .

# Version als Build-Konstante ins Bundle eingebrannt (Default dev). Der
# Release-Workflow uebergibt den Tag per --build-arg VERSION=<tag>. Das
# Argument wird im RUN konsumiert, damit ein Versionswechsel die Build-Schicht
# invalidiert und nicht das gecachte alte Bundle im neuen Image landet.
ARG VERSION=dev
RUN VERSION="${VERSION}" pnpm build

# Runtime stage: serve static files with nginx
FROM nginx:1.31-alpine AS runner

COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
