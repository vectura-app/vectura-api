# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

RUN apk --no-cache add ca-certificates sqlite gcc musl-dev sqlite-dev

# Build optimized binary
RUN go build \
    -ldflags="-w -s" \
    -trimpath \
    -o vectura-api \
    .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/vectura-api .

# Copy cities.yaml config file
COPY --from=builder /app/cities.yaml .

# Make sure the binary is executable
RUN chmod +x vectura-api

# Run the application
CMD ["./vectura-api"]
