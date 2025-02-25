# Use golang:1.23-alpine as the base image
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app


# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Update dependencies and ensure go.sum is up to date
RUN go mod tidy
RUN go mod download

# Copy the entire codebase into the container
COPY . .

# Run go mod tidy again with the full codebase to ensure all imports are accounted for
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o keeper-execution ./cmd/keeper

# Use a minimal alpine image for the final stage
FROM alpine:3.18

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/keeper-execution .

COPY .env .env

# Expose the port the service runs on
EXPOSE 8080

# Set environment variables (customize as needed)
ENV GIN_MODE=release
ENV LOG_LEVEL=info

# Run the binary
CMD ["./keeper-execution"]