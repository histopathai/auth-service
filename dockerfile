# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go mod tidy

RUN go run github.com/swaggo/swag/cmd/swag init --output ./docs --dir ./... --generalInfo ./cmd/main.go

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o auth-service ./cmd/main.go

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder --chown=appuser:appgroup /app/auth-service .
COPY --from=builder --chown=appuser:appgroup /app/docs ./docs

USER appuser

CMD ["./auth-service"]
