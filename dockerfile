# Stage 1: Build the Go application
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod tidy && \
    go get github.com/swaggo/gin-swagger@v1.6.0 && \
    go get github.com/swaggo/swag@v1.8.12 && \
    go mod tidy

RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12

RUN swag init --output ./docs --dir ./ --generalInfo ./cmd/main.go

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o auth-service ./cmd/main.go

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/auth-service .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser


CMD ["./auth-service"]
