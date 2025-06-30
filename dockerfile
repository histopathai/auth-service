# Stage 1: Build the Go application
FROM golang:1.23.2-alpine AS builder

# Set working directory
WORKDIR /app

# Install git for go mod download and swag tool
RUN apk add --no-cache git

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./

# --- KRİTİK GÜNCELLEMELER BAŞLANGICI ---
# Go modül bağımlılıklarını temizliyoruz ve sabit uyumlu sürümleri alıyoruz
RUN go mod tidy && \
    go get github.com/swaggo/gin-swagger@v1.6.0 && \
    go get github.com/swaggo/swag@v1.8.12 && \
    go mod tidy
# --- KRİTİK GÜNCELLEMELER SONU ---

# İlgili bağımlılıkları indiriyoruz
RUN go mod download

# Uygulamanın geri kalan kaynak kodlarını kopyalıyoruz
COPY . .

# Sabitlenmiş sürümde swag aracını kur
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12

# Swagger dökümantasyonunu oluştur
RUN swag init --output ./docs --dir ./ --generalInfo ./cmd/main.go

# Go uygulamasını derle
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o auth-service ./cmd/main.go

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

# Builder'dan derlenen binary'i kopyala
COPY --from=builder /app/auth-service .

# Non-root kullanıcı oluştur
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Uygulamanın çalıştığı portu aç
EXPOSE 8080

# Uygulama çalıştırma komutu
CMD ["./auth-service"]
