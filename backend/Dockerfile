# --- Build Stage ---
FROM golang:1.26.0-alpine AS builder

RUN apk update

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -o /go/bin/jotti ./main.go

# --- Deploy Stage ---
FROM alpine:3.21

COPY --from=builder /go/bin/jotti /usr/local/bin/jotti

ENTRYPOINT [ "jotti" ]
