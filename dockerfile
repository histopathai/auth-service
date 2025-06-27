# Stage 1: Build the Go application
FROM golang:1.23.2-alpine AS builder

# Set working directory
WORKDIR /app


# Install necessary packages
COPY go.mod go.sum ./
RUN go mod download


# Copy the rest of the application source code
COPY . .


## Build the application
# CGO_ENABLED=0  is important static binaries when using alpine base image
# -a -installsuffix cgo reduces the image size
# -o authservice specifies the output executable name
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o auth-service ./cmd/main.go


# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/auth-service .

#Create a non-root user and group
# This is a security best practice for production containers
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./auth-service"]