#!/bin/bash

# Function to display comprehensive help menu
show_help() {
    cat << 'EOF'
TriggerX Docker Build Script
============================

DESCRIPTION:
    Builds Docker images for TriggerX microservices with specified versions.
    Supports building individual services or all services in parallel.

USAGE:
    $0 -n <service> -v <version>
    $0 -h|--help

OPTIONS:
    -n, --service    Service name to build (required)
    -v, --version    Version tag for the Docker image (required, format: MAJOR.MINOR.PATCH)
    -h, --help       Display this help message

AVAILABLE SERVICES:
    keeper              - TriggerX Keeper service (uses Dockerfile.keeper)
    imua-keeper         - TriggerX Imua Keeper service (uses Dockerfile.imua-keeper)
    dbserver            - Database server service
    registrar           - Registrar service
    health              - Health monitoring service
    taskdispatcher      - Task dispatcher service
    taskmonitor         - Task monitor service
    schedulers/time     - Time-based scheduler service
    schedulers/condition - Condition-based scheduler service
    all                 - Build all services in parallel (excluding keepers)

VERSION FORMAT:
    Must follow semantic versioning: MAJOR.MINOR.PATCH
    Examples: 0.0.1, 1.2.3, 2.0.0
EOF
    exit 0
}

# Function to display usage (simplified version for errors)
usage() {
    echo "Usage: $0 -n <service> -v <version>"
    echo "Use '$0 -h' for detailed help and examples"
    exit 1
}

# Parse command-line arguments
while getopts ":n:v:h-:" opt; do
    case ${opt} in
        n )
            SERVICE=$OPTARG
            ;;
        v )
            VERSION=$OPTARG
            ;;
        h )
            show_help
            ;;
        - )
            # Handle long options
            case "${OPTARG}" in
                service=* )
                    SERVICE="${OPTARG#*=}"
                    ;;
                version=* )
                    VERSION="${OPTARG#*=}"
                    ;;
                help )
                    show_help
                    ;;
                * )
                    echo "Unknown long option: --${OPTARG}" 1>&2
                    usage
                    ;;
            esac
            ;;
        \? )
            echo "Invalid option: -$OPTARG" 1>&2
            usage
            ;;
        : )
            echo "Option -$OPTARG requires an argument" 1>&2
            usage
            ;;
    esac
done

# Check if no arguments were provided
if [ $# -eq 0 ]; then
    echo "Error: No arguments provided" 1>&2
    show_help
fi

# Check if name is provided
if [ -z "$SERVICE" ]; then
    echo "Error: Service (-n) is required" 1>&2
    usage
fi

# Check if version is provided
if [ -z "$VERSION" ]; then
    echo "Error: Version (-v) is required (e.g., 0.0.1)" 1>&2
    usage
fi

# Validate the service from the list of allowed services
if [[ ! "$SERVICE" =~ ^(keeper|imua-keeper|dbserver|health|taskdispatcher|taskmonitor|schedulers/time|schedulers/condition|all)$ ]]; then
    echo "Error: Invalid service. Allowed services are: keeper, imua-keeper, dbserver, health, taskdispatcher, taskmonitor, schedulers/time, schedulers/condition" 1>&2
    exit 1
fi

# Validate version format (basic regex for semantic versioning)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Use MAJOR.MINOR.PATCH (e.g., 0.0.1)" 1>&2
    exit 1
fi

if [[ "$SERVICE" == "all" ]]; then
    # Build all services in parallel
    services=(dbserver health taskdispatcher taskmonitor schedulers/time schedulers/condition)
    build_pids=()
    failed_services=()
    
    # Function to build a single service
    build_service() {
        local service=$1
        local version=$2
        local docker_name=$(echo $service | sed 's/\//-/g')
        
        echo "[$(date '+%H:%M:%S')] Starting build for $service..."
        if docker build --no-cache \
            -f Dockerfile.backend \
            --build-arg SERVICE=${service} \
            --build-arg DOCKER_NAME=${docker_name} \
            -t triggerx-${docker_name}:${version} . > "build_${docker_name}.log" 2>&1; then
            echo "[$(date '+%H:%M:%S')] ✅ Successfully built $service"
            return 0
        else
            echo "[$(date '+%H:%M:%S')] ❌ Failed to build $service (check build_${docker_name}.log)"
            return 1
        fi
    }
    
    echo "Starting parallel builds for ${#services[@]} services..."
    
    # Start all builds in parallel
    for service in "${services[@]}"; do
        build_service "$service" "$VERSION" &
        build_pids+=($!)
    done
    
    # Wait for all builds to complete and collect results
    echo "Waiting for all builds to complete..."
    for i in "${!build_pids[@]}"; do
        if ! wait "${build_pids[$i]}"; then
            failed_services+=("${services[$i]}")
        fi
    done
    
    # Report results
    echo ""
    echo "=== Build Summary ==="
    if [[ ${#failed_services[@]} -eq 0 ]]; then
        echo "✅ All ${#services[@]} services built successfully!"
    else
        echo "❌ ${#failed_services[@]} service(s) failed to build:"
        for service in "${failed_services[@]}"; do
            echo "  - $service"
        done
        echo ""
        echo "Check individual log files (build_*.log) for detailed error information."
        exit 1
    fi
    
    # Clean up log files on success
    for service in "${services[@]}"; do
        docker_name=$(echo $service | sed 's/\//-/g')
        rm -f "build_${docker_name}.log"
    done
elif [[ "$SERVICE" == "keeper" ]]; then
    echo "Building $SERVICE..."
    docker build --no-cache \
        -f Dockerfile.keeper \
        -t triggerx-keeper:${VERSION} .
elif [[ "$SERVICE" == "imua-keeper" ]]; then
    echo "Building $SERVICE..."
    docker build --no-cache \
        -f Dockerfile.imua-keeper \
        -t triggerx-imua-keeper:${VERSION} .
else
    echo "Building $SERVICE..."
    # Convert service name to Docker-compatible name
    DOCKER_NAME=$(echo $SERVICE | sed 's/\//-/g')

    echo "DOCKER_NAME: $DOCKER_NAME"
    # Build a single service
    echo "Building $SERVICE..."
    docker build --no-cache \
        -f Dockerfile.backend \
        --build-arg SERVICE=${SERVICE} \
        --build-arg DOCKER_NAME=${DOCKER_NAME} \
        -t triggerx-${DOCKER_NAME}:${VERSION} .
fi

if [[ "$SERVICE" == "all" ]]; then
    echo "All services built successfully with version ${VERSION}"
else
    echo "Successfully built: triggerx-${SERVICE}:${VERSION}"
fi
