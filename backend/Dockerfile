# --- Build Stage ---
FROM golang:1.24-alpine AS builder

RUN apk update

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -o /go/bin/jotti ./main.go

# --- Deploy Stage ---
FROM alpine:latest

COPY --from=builder /go/bin/jotti /usr/local/bin/jotti

ENTRYPOINT [ "jotti" ]
