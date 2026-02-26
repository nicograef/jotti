FROM alpine:3.22

ARG MIGRATE_VERSION=v4.19.1

RUN apk add --no-cache curl \
    && curl -fsSL "https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz" \
    | tar -xz -C /usr/local/bin \
    && apk del curl

CMD /usr/local/bin/migrate -path /migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/jotti?sslmode=disable" up

