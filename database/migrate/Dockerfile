FROM alpine:3.22

ARG MIGRATE_VERSION=v4.19.1

RUN apk add --no-cache curl \
    && curl -fsSL "https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz" \
    | tar -xz -C /usr/local/bin \
    && apk del curl

# Migrationen ins Image backen, damit der Stack selbst-enthalten ist (kein
# Bind-Mount vom Host noetig). Der Build-Context ist deshalb ./database statt
# ./database/migrate — siehe die compose-Dateien und .github/workflows/release.yml.
COPY migrations /migrations

CMD /usr/local/bin/migrate -path /migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/jotti?sslmode=disable" up

