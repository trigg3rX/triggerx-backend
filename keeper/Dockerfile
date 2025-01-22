# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /keeper

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o keeper ./cmd/keeper

# Final stage
FROM alpine:3.18

WORKDIR /keeper

# Copy only the binary from builder
COPY --from=builder /keeper/keeper .
# Copy config file from the correct location
COPY config-files/triggerx_keeper.yaml /keeper/config-files/triggerx_keeper.yaml

# Run the keeper
ENTRYPOINT ["./keeper"]