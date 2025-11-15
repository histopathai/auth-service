# Stage 1: Build the Go application
FROM golang:1.23.2-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod tidy && \
    go get -u github.com/swaggo/swag && \
    go mod tidy

RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12

RUN swag init --output ./docs --dir ./ --generalInfo ./cmd/main.go

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o auth-service ./cmd/main.go


# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder --chown=appuser:appgroup /app/auth-service .
COPY --from=builder --chown=appuser:appgroup /app/docs ./docs

USER appuser

CMD ["./auth-service"]
