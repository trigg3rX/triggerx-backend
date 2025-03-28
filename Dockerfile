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
FROM golang:1.23-alpine

# Install ca-certificates, curl, Docker, and Docker Compose
RUN apk --no-cache add ca-certificates curl \
    && apk --no-cache add docker \
    && apk --no-cache add docker-compose

# Install npm for othentic-cli
RUN apk --no-cache add nodejs npm

# Install othentic-cli globally
RUN npm i -g @othentic/othentic-cli@1.8.1

# Set working directory
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/keeper-execution .

# Copy the docker-compose.yaml file from the builder stage
COPY --from=builder /app/cmd/keeper/docker-compose.yaml ./

RUN touch .env

COPY ./scripts/services/start-keeper.sh /root/start-keeper.sh
# Expose the port the service runs on
EXPOSE 9005

# Set environment variables (customize as needed)
ENV GIN_MODE=release
ENV LOG_LEVEL=info

# Disable SSL verification for HTTP client
ENV GODEBUG=http2client=0
ENV HTTPS_PROXY=""
ENV HTTP_PROXY=""

# Run the startup script
CMD ["sh", "./start-keeper.sh"]
# CMD ["sleep", "7200"]