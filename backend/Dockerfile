FROM golang:1.27-alpine AS builder

# The image tag names the Go minor; the go directive in go.mod names the patch.
ENV GOTOOLCHAIN=auto

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Der Release-Workflow uebergibt den Tag per --build-arg VERSION=<tag>.
ARG VERSION=dev
RUN go build -ldflags "-X main.version=${VERSION}" -o /go/bin/jotti ./main.go

FROM alpine:3.24

# Unprivilegiert: das Backend schreibt nicht ins Dateisystem und lauscht auf 3000 (>1024).
RUN addgroup -S jotti && adduser -S -G jotti jotti

COPY --from=builder /go/bin/jotti /usr/local/bin/jotti

USER jotti

ENTRYPOINT [ "jotti" ]
