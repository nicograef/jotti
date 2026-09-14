FROM node:24-alpine AS build

WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN corepack enable \
  && corepack prepare pnpm@11 --activate \
  && pnpm install --frozen-lockfile

COPY . .

# Das VERSION-Argument wird im RUN konsumiert, damit ein Versionswechsel die
# Build-Schicht invalidiert und nicht das alte Bundle im neuen Image landet.
ARG VERSION=dev
RUN VERSION="${VERSION}" pnpm build

FROM nginx:1.31-alpine AS runner

COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
