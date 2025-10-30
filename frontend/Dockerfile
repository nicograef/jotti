# Build stage: compile the React (Vite) app
FROM node:22-alpine AS build

WORKDIR /app

# Install dependencies with pnpm via Corepack (respects pnpm-lock.yaml)
COPY package.json pnpm-lock.yaml ./
RUN corepack enable \
  && corepack prepare pnpm@latest --activate \
  && pnpm install --frozen-lockfile

# Copy the rest of the sources and build
COPY . .
RUN pnpm build

# Runtime stage: serve static files with nginx
FROM nginx:1.27-alpine AS runner

COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
