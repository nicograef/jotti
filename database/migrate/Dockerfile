FROM alpine:latest

# Install migrate tool by copying the executable
COPY migrate /usr/local/bin/migrate

# run the migrations on container startup with the provided environment variables and mounted migration files
CMD /usr/local/bin/migrate -path /migrations -database postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@postgres:5432/jotti?sslmode=disable up

