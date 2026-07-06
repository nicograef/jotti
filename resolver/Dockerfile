# --- Build Stage ---
FROM golang:1.26.4-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -o /go/bin/jotti-resolver .

# --- Deploy Stage ---
FROM alpine:3.24

COPY --from=builder /go/bin/jotti-resolver /usr/local/bin/jotti-resolver

EXPOSE 53/udp 53/tcp

ENTRYPOINT [ "jotti-resolver" ]
