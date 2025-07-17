#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 -n <service> -v <version>"
    echo "Example: $0 -n keeper -v 0.0.1"
    exit 1
}

# Parse command-line arguments
while getopts ":n:v:" opt; do
    case ${opt} in
        n )
            SERVICE=$OPTARG
            ;;
        v )
            VERSION=$OPTARG
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

# Validate the service from the list of allowed services
if [[ ! "$SERVICE" =~ ^(keeper|dbserver|registrar|health|redis|schedulers/time|schedulers/condition|all)$ ]]; then
    echo "Error: Invalid service. Allowed services are: keeper, dbserver, registrar, health, redis, schedulers/time, schedulers/condition" 1>&2
    exit 1
fi

# Validate version format (basic regex for semantic versioning)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Use MAJOR.MINOR.PATCH (e.g., 0.0.1)" 1>&2
    exit 1
fi

if [[ "$SERVICE" == "all" ]]; then
    # Build all services in parallel
    services=(dbserver registrar health redis schedulers/time schedulers/condition)
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
