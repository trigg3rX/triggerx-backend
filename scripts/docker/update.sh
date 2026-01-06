#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 -n <service>"
    echo "Example: $0 -n schedulers/time"
    echo ""
    echo "This script updates the YAML configuration file for a running service"
    echo "and restarts the container to apply the changes."
    echo ""
    echo "Supported services:"
    echo "  - schedulers/time (uses config/services/time-scheduler.yaml)"
    echo "  - schedulers/condition (uses config/services/condition-scheduler.yaml)"
    echo "  - dbserver (uses config/services/docker-executor.yaml)"
    exit 1
}

# Parse command-line arguments
while getopts ":n:" opt; do
    case ${opt} in
        n )
            SERVICE=$OPTARG
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

# Check if service is provided
if [ -z "$SERVICE" ]; then
    echo "Error: Service is required" 1>&2
    usage
fi

# Validate the service from the list of allowed services
if [[ ! "$SERVICE" =~ ^(dbserver|schedulers/time|schedulers/condition)$ ]]; then
    echo "Error: Invalid service. Allowed services are: dbserver, schedulers/time, schedulers/condition" 1>&2
    exit 1
fi

# Convert service name to Docker-compatible name
DOCKER_NAME=$(echo $SERVICE | sed 's/\//-/g')
CONTAINER_NAME="triggerx-${DOCKER_NAME}"

# Check if container exists and is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Error: Container ${CONTAINER_NAME} is not running" 1>&2
    echo "Please start the container first using: ./scripts/docker/pull-and-run.sh -n ${SERVICE} -v <version> -p <port>"
    exit 1
fi

# Map service names to their YAML config files
declare -A SERVICE_CONFIG_MAP
SERVICE_CONFIG_MAP["schedulers/time"]="config/services/time-scheduler.yaml"
SERVICE_CONFIG_MAP["schedulers/condition"]="config/services/condition-scheduler.yaml"
SERVICE_CONFIG_MAP["dbserver"]="config/services/docker-executor.yaml"

# Get the config file path for this service
CONFIG_FILE="${SERVICE_CONFIG_MAP[$SERVICE]}"

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Config file not found: ${CONFIG_FILE}" 1>&2
    echo "Please ensure the YAML config file exists in the expected location."
    exit 1
fi

# Get the container's config directory path
CONTAINER_CONFIG_DIR="/home/appuser/config/services"
CONTAINER_CONFIG_FILE="${CONTAINER_CONFIG_DIR}/$(basename ${CONFIG_FILE})"

echo "Updating configuration for service: ${SERVICE}"
echo "Config file: ${CONFIG_FILE}"
echo "Container: ${CONTAINER_NAME}"

# Check if config is mounted as a volume (preferred method)
# If mounted, we just need to ensure the local file is up to date
# If not mounted, we'll copy it into the container
MOUNT_SOURCE=$(docker inspect ${CONTAINER_NAME} --format '{{range .Mounts}}{{if eq .Destination "/home/appuser/config/services"}}{{.Source}}{{end}}{{end}}' 2>/dev/null)

if [ -n "$MOUNT_SOURCE" ]; then
    echo "✓ Config directory is mounted as volume from: ${MOUNT_SOURCE}"
    echo "  Local file changes will be reflected in the container"
    echo "  Ensuring local file is up to date..."
    # The file should already be in the right place if volume is mounted
    # Just verify it exists
    if [ ! -f "${CONFIG_FILE}" ]; then
        echo "Error: Config file ${CONFIG_FILE} not found locally" 1>&2
        exit 1
    fi
    echo "✓ Local config file verified"
else
    # Config not mounted, copy it into the container
    echo "Config directory not mounted, copying file to container..."
    if docker cp "${CONFIG_FILE}" "${CONTAINER_NAME}:${CONTAINER_CONFIG_FILE}"; then
        echo "✓ Config file copied successfully"
        # Set proper permissions on the config file in the container
        docker exec ${CONTAINER_NAME} chmod 644 ${CONTAINER_CONFIG_FILE} 2>/dev/null || true
    else
        echo "Error: Failed to copy config file to container" 1>&2
        exit 1
    fi
fi

# Restart the container to apply the new configuration
echo "Restarting container to apply new configuration..."
if docker restart ${CONTAINER_NAME}; then
    echo "✓ Container restarted successfully"
    echo ""
    echo "Configuration update complete!"
    echo "To view logs: docker logs -f ${CONTAINER_NAME}"
else
    echo "Error: Failed to restart container" 1>&2
    exit 1
fi
