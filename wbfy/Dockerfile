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
RUN go build -o wbfy .

# Create final lightweight image
FROM alpine:latest

WORKDIR /app

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata docker-cli

# Copy the binary from the builder stage
COPY --from=builder /app/wbfy .

# Copy web assets if any
COPY --from=builder /app/web ./web

# Expose the port
EXPOSE 8081

# Run the application
CMD ["./wbfy"]
