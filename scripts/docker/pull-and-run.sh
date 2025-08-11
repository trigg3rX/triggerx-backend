#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 -n <service> -v <version> -p <port>"
    echo "Example: $0 -n health -v 0.0.1 -p 8080"
    exit 1
}

# Parse command-line arguments
while getopts ":n:v:p:" opt; do
    case ${opt} in
        n )
            SERVICE=$OPTARG
            ;;
        v )
            VERSION=$OPTARG
            ;;
        p )
            PORT=$OPTARG
            ;;
        \? )
            echo "Invalid option: $OPTARG" 1>&2
            usage
            ;;
        : )
            echo "Invalid option: $OPTARG requires an argument" 1>&2
            usage
            ;;
    esac
done

# Check if name is provided
if [ -z "$SERVICE" ]; then
    echo "Error: Service is required" 1>&2
    usage
fi

# Check if version is provided
if [ -z "$VERSION" ]; then
    echo "Error: Version is required (e.g., 0.0.1)" 1>&2
    usage
fi

# Check if port is provided
if [ -z "$PORT" ]; then
    echo "Error: Port is required" 1>&2
    usage
fi

# Validate the service from the list of allowed services
if [[ ! "$SERVICE" =~ ^(dbserver|health|taskdispatcher|taskmonitor|schedulers/time|schedulers/condition)$ ]]; then
    echo "Error: Invalid service. Allowed services are: dbserver, health, taskdispatcher, taskmonitor, schedulers/time, schedulers/condition" 1>&2
    exit 1
fi

# Validate version format (basic regex for semantic versioning)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Use MAJOR.MINOR.PATCH (e.g., 0.0.1)" 1>&2
    exit 1
fi

# Convert service name to Docker-compatible name
DOCKER_NAME=$(echo $SERVICE | sed 's/\//-/g')

# Check if container already exists and remove it
if docker ps -a | grep -q "triggerx-${DOCKER_NAME}"; then
    echo "Stopping and removing existing container triggerx-${DOCKER_NAME}..."
    docker stop triggerx-${DOCKER_NAME} 2>/dev/null
    docker rm triggerx-${DOCKER_NAME} 2>/dev/null
fi

# Check if local image exists
if ! docker images --filter reference="triggerx-${DOCKER_NAME}:${VERSION}" | grep "triggerx-${DOCKER_NAME}"; then
    echo "Local image triggerx-${DOCKER_NAME}:${VERSION} does not exist" 1>&2
    IMAGE_NAME="trigg3rx/triggerx-${DOCKER_NAME}:${VERSION}"
    echo "Pulling ${IMAGE_NAME}..."
    docker pull ${IMAGE_NAME}
else
    echo "Using local image triggerx-${DOCKER_NAME}:${VERSION}"
    IMAGE_NAME="triggerx-${DOCKER_NAME}:${VERSION}"
fi

# Check if .env file exists and set up volume mount
if [ -f .env ]; then
    echo "Found .env file, mounting it to container..."
    ENV_FILE="-v $(pwd)/.env:/home/appuser/.env"
else
    echo "Warning: .env file not found in current directory. Container will run without environment variables."
    echo "To use environment variables, create a .env file in: $(pwd)/.env"
    ENV_FILE=""
fi

# Ensure the log directory exists and has proper permissions
echo "Setting up log directory: ./data/logs/${DOCKER_NAME}"
mkdir -p "./data/logs/${DOCKER_NAME}"

# Ensure the cache directory exists and has proper permissions
echo "Setting up cache directory: ./data/cache"
mkdir -p "./data/cache"

# Get current user and group IDs
CURRENT_UID=$(id -u)
CURRENT_GID=$(id -g)
CURRENT_USER=$(id -un)
CURRENT_GROUP=$(id -gn)

# Get promtail group ID if it exists
PROMTAIL_GID=""
if command -v getent >/dev/null 2>&1 && getent group promtail >/dev/null 2>&1; then
    PROMTAIL_GID=$(getent group promtail | cut -d: -f3)
    echo "Found promtail group with GID: ${PROMTAIL_GID}"
fi

# Get docker group ID for services that need docker socket access
DOCKER_GID=""
if command -v getent >/dev/null 2>&1 && getent group docker >/dev/null 2>&1; then
    DOCKER_GID=$(getent group docker | cut -d: -f3)
    echo "Found docker group with GID: ${DOCKER_GID}"
fi

# Set appropriate ownership and permissions
# For dbserver, use docker group if available for socket access
if [[ "$SERVICE" == "dbserver" && -n "$DOCKER_GID" ]]; then
    echo "Setting ownership for dbserver to UID ${CURRENT_UID} (${CURRENT_USER}) and GID ${DOCKER_GID} (docker group)..."
    sudo chown ${CURRENT_UID}:${DOCKER_GID} "./data/logs/${DOCKER_NAME}" "./data/cache" 2>/dev/null || {
        echo "Warning: Could not set docker group ownership. Using fallback permissions..."
        chmod 775 "./data/logs/${DOCKER_NAME}" "./data/cache" 2>/dev/null
    }
    chmod ug+rws,o+r "./data/logs/${DOCKER_NAME}" "./data/cache" 2>/dev/null
    USER_MAPPING="--user ${CURRENT_UID}:${DOCKER_GID}"
elif [ -n "$PROMTAIL_GID" ]; then
    # Set owner to current user and group to promtail for log reading
    echo "Setting ownership to UID ${CURRENT_UID} (${CURRENT_USER}) and GID ${PROMTAIL_GID} (promtail)..."
    sudo chown ${CURRENT_UID}:${PROMTAIL_GID} "./data/logs/${DOCKER_NAME}" "./data/cache" 2>/dev/null || {
        echo "Warning: Could not set specific ownership. Using fallback permissions..."
        chmod 775 "./data/logs/${DOCKER_NAME}" "./data/cache" 2>/dev/null
    }
    # User can read/write, group (promtail) can read/write, others can read
    # Set setgid bit (s) so new files automatically inherit the directory's group ownership
    chmod ug+rws,o+r "./data/logs/${DOCKER_NAME}" "./data/cache" 2>/dev/null
    
    # Set user mapping for docker container to run with promtail group
    USER_MAPPING="--user ${CURRENT_UID}:${PROMTAIL_GID}"
    echo "Container will run as UID ${CURRENT_UID} (${CURRENT_USER}), GID ${PROMTAIL_GID} (promtail group)"
else
    # No promtail group, set to current user's group
    echo "No promtail group found. Setting ownership to UID ${CURRENT_UID} (${CURRENT_USER})..."
    sudo chown ${CURRENT_UID}:${CURRENT_GID} "./data/logs/${DOCKER_NAME}" "./data/cache" 2>/dev/null || {
        echo "Warning: Could not set ownership. Using world-writable fallback..."
        chmod 777 "./data/logs/${DOCKER_NAME}" "./data/cache" 2>/dev/null
    }
    # Set setgid bit so new files inherit group ownership
    chmod ug+rws,o+r "./data/logs/${DOCKER_NAME}" "./data/cache" 2>/dev/null
    USER_MAPPING="--user ${CURRENT_UID}:${CURRENT_GID}"
fi

# Fix permissions on any existing log files
if [ "$(ls -A "./data/logs/${DOCKER_NAME}" 2>/dev/null)" ]; then
    echo "Fixing permissions on existing log files..."
    if [[ "$SERVICE" == "dbserver" && -n "$DOCKER_GID" ]]; then
        sudo chown ${CURRENT_UID}:${DOCKER_GID} "./data/logs/${DOCKER_NAME}"/* 2>/dev/null
        chmod 664 "./data/logs/${DOCKER_NAME}"/* 2>/dev/null
    elif [ -n "$PROMTAIL_GID" ]; then
        sudo chown ${CURRENT_UID}:${PROMTAIL_GID} "./data/logs/${DOCKER_NAME}"/* 2>/dev/null
        chmod 664 "./data/logs/${DOCKER_NAME}"/* 2>/dev/null
    else
        sudo chown ${CURRENT_UID}:${CURRENT_GID} "./data/logs/${DOCKER_NAME}"/* 2>/dev/null
        chmod 664 "./data/logs/${DOCKER_NAME}"/* 2>/dev/null
    fi
fi

# Fix permissions on any existing cache files
if [ "$(ls -A "./data/cache" 2>/dev/null)" ]; then
    echo "Fixing permissions on existing cache files..."
    if [[ "$SERVICE" == "dbserver" && -n "$DOCKER_GID" ]]; then
        sudo chown ${CURRENT_UID}:${DOCKER_GID} "./data/cache"/* 2>/dev/null
        chmod 664 "./data/cache"/* 2>/dev/null
    elif [ -n "$PROMTAIL_GID" ]; then
        sudo chown ${CURRENT_UID}:${PROMTAIL_GID} "./data/cache"/* 2>/dev/null
        chmod 664 "./data/cache"/* 2>/dev/null
    else
        sudo chown ${CURRENT_UID}:${CURRENT_GID} "./data/cache"/* 2>/dev/null
        chmod 664 "./data/cache"/* 2>/dev/null
    fi
fi

echo "Log and cache directory permissions set successfully."

if [[ "$SERVICE" == "dbserver" ]]; then
    # Special handling for dbserver which needs docker socket access
    if [ -n "$DOCKER_GID" ]; then
        DBSERVER_USER_MAPPING="--user ${CURRENT_UID}:${DOCKER_GID}"
        echo "Container will run as UID ${CURRENT_UID} (${CURRENT_USER}), GID ${DOCKER_GID} (docker group) for socket access"
    else
        echo "Warning: Docker group not found. Container may not be able to access Docker socket."
        DBSERVER_USER_MAPPING="${USER_MAPPING}"
    fi
    
    # Run the container
    echo "Starting container triggerx-${DOCKER_NAME}..."
    docker run -d \
        --name triggerx-${DOCKER_NAME} \
        ${DBSERVER_USER_MAPPING} \
        ${ENV_FILE} \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v ./data/logs/${DOCKER_NAME}:/home/appuser/data/logs/${DOCKER_NAME} \
        -v ./data/cache:/home/appuser/data/cache \
        -p ${PORT}:${PORT} \
        ${IMAGE_NAME}
else
    # Run the container
    echo "Starting container triggerx-${DOCKER_NAME}..."
    docker run -d \
        --name triggerx-${DOCKER_NAME} \
        ${USER_MAPPING} \
        ${ENV_FILE} \
        --network host \
        -v ./data/logs/${DOCKER_NAME}:/home/appuser/data/logs/${DOCKER_NAME} \
        -v ./data/cache:/home/appuser/data/cache \
        --restart unless-stopped \
        ${IMAGE_NAME}
fi

# Check if container started successfully
if [ $? -eq 0 ]; then
    echo "Container started successfully!"
    echo "Container ID: $(docker ps -q -f name=triggerx-${DOCKER_NAME})"
    echo "Port: ${PORT}"
    echo "To view logs: docker logs triggerx-${DOCKER_NAME}"
    echo "To stop: docker stop triggerx-${DOCKER_NAME}"
else
    echo "Error: Failed to start container"
    exit 1
fi
