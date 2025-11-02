FROM alpine:latest

# install golang-migrate CLI
RUN apk add --no-cache curl tar
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz

# run the migrations on container startup with the provided environment variables and mounted migration files
CMD ./migrate -path /migrations -database postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@postgres:5432/jotti?sslmode=disable up

