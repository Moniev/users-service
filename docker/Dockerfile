FROM golang:1.24 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -v -o /app/users-service ./app/cmd/main.go && \
    echo "--- Build successful, checking for binary in builder stage ---" && \
    ls -l /app/users-service

FROM alpine:latest

RUN apk --no-cache add bash ca-certificates curl libc6-compat

WORKDIR /app

COPY --from=builder /app/users-service /app/users-service

RUN echo "--- Checking for binary in final stage before chmod ---" && \
    ls -l /app/users-service && \
    chmod +x /app/users-service && \
    echo "--- Chmod successful ---"

RUN curl -fsSL -o /app/wait-for-it.sh https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh && \
    chmod +x /app/wait-for-it.sh

RUN cat <<EOF > /app/start.sh
#!/bin/sh
echo "Waiting for PostgreSQL..."
/app/wait-for-it.sh postgres.postgres.svc.cluster.local:5432 --timeout=60 -- echo "PostgreSQL is up"

echo "Waiting for Redis..."
/app/wait-for-it.sh redis.redis.svc.cluster.local:6379 --timeout=60 -- echo "Redis is up"

echo "Waiting for Kafka..."
/app/wait-for-it.sh kafka.kafka.svc.cluster.local:9092 --timeout=60 -- echo "Kafka is up"

echo "Starting users-service..."
./users-service
EOF

RUN chmod +x /app/start.sh

EXPOSE 8000

ENTRYPOINT ["/app/start.sh"]
