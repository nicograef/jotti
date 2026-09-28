FROM alpine:3.24

ARG MIGRATE_VERSION=v4.20.1
# From sha256sum.txt of the release; bump it together with MIGRATE_VERSION.
ARG MIGRATE_SHA256=b1e13121d4178b7804a9fb29dca62d57b732390f7e5ae33f2cd65e9b192e1bd4

RUN apk add --no-cache curl \
    && curl -fsSL -o /tmp/migrate.tar.gz "https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz" \
    && echo "${MIGRATE_SHA256}  /tmp/migrate.tar.gz" | sha256sum -c - \
    && tar -xzf /tmp/migrate.tar.gz -C /usr/local/bin migrate \
    && rm /tmp/migrate.tar.gz \
    && apk del curl

# Unprivilegierter Benutzer: migrate liest nur /migrations und spricht die
# Datenbank an, schreibt also nichts ins Dateisystem.
RUN addgroup -S jotti && adduser -S -G jotti jotti

# Migrationen ins Image backen, damit der Stack selbst-enthalten ist (kein
# Bind-Mount vom Host noetig). Der Build-Context ist deshalb ./database statt
# ./database/migrate — siehe die compose-Dateien und .github/workflows/release.yml.
COPY migrations /migrations

USER jotti

CMD /usr/local/bin/migrate -path /migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/jotti?sslmode=disable" up

