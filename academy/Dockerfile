FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install necessary tools
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum files
COPY ./go.mod ./go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o academy ./cmd

# Create final lightweight image
FROM alpine:latest

WORKDIR /app

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/academy .

# Copy web assets and templates
COPY --from=builder /app/web ./web

# Create necessary directories
RUN mkdir -p /app/problems

# Expose the port
EXPOSE 8080

# Run the application
CMD ["./academy"]
