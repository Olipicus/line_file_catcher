FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o linefilecatcher ./cmd/linefilecatcher

# Use a minimal alpine image for the final container
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/linefilecatcher .

# Create storage directory
RUN mkdir -p /app/storage

# Set environment variables
ENV PORT=8080
ENV STORAGE_DIR=/app/storage

EXPOSE 8080

# Run the application
CMD ["./linefilecatcher"]