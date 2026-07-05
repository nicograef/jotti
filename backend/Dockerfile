# --- Build Stage ---
FROM golang:1.26.0-alpine AS builder

RUN apk update

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Version per ldflags einkompiliert (Default dev). Der Release-Workflow
# uebergibt den Tag per --build-arg VERSION=<tag>.
ARG VERSION=dev
RUN go build -ldflags "-X main.version=${VERSION}" -o /go/bin/jotti ./main.go

# --- Deploy Stage ---
FROM alpine:3.22

# Unprivilegierter Benutzer: das Backend schreibt nicht ins Dateisystem und
# lauscht auf 3000 (>1024), braucht also keine root-Rechte.
RUN addgroup -S jotti && adduser -S -G jotti jotti

COPY --from=builder /go/bin/jotti /usr/local/bin/jotti

USER jotti

ENTRYPOINT [ "jotti" ]
