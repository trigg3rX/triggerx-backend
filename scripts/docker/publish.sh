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
if [[ ! "$SERVICE" =~ ^(keeper|dbserver|health|taskdispatcher|taskmonitor|schedulers/time|schedulers/condition|all)$ ]]; then
    echo "Error: Invalid service. Allowed services are: keeper, dbserver, health, taskdispatcher, taskmonitor, schedulers/time, schedulers/condition" 1>&2
    exit 1
fi

# Validate version format (basic regex for semantic versioning)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Use MAJOR.MINOR.PATCH (e.g., 0.0.1)" 1>&2
    exit 1
fi

# Login to Docker Hub
docker login

if [[ "$SERVICE" == "all" ]]; then
    # Push all services in parallel
    services=(dbserver health taskdispatcher taskmonitor schedulers/time schedulers/condition)
    publish_pids=()
    failed_services=()
    
    # Function to publish a single service
    publish_service() {
        local service=$1
        local version=$2
        local docker_name=$(echo $service | sed 's/\//-/g')
        
        echo "[$(date '+%H:%M:%S')] Starting publish for $service..."
        
        # Tag images with version and latest
        if ! docker tag triggerx-${docker_name}:${version} trigg3rx/triggerx-${docker_name}:${version} > "publish_${docker_name}.log" 2>&1; then
            echo "[$(date '+%H:%M:%S')] ❌ Failed to tag $service for version ${version}"
            return 1
        fi
        
        # Push images
        if ! docker push trigg3rx/triggerx-${docker_name}:${version} >> "publish_${docker_name}.log" 2>&1; then
            echo "[$(date '+%H:%M:%S')] ❌ Failed to push $service version ${version}"
            return 1
        fi
        
        # Clean up remote tags
        docker rmi trigg3rx/triggerx-${docker_name}:${version} >> "publish_${docker_name}.log" 2>&1
        
        echo "[$(date '+%H:%M:%S')] ✅ Successfully published $service"
        return 0
    }
    
    echo "Starting parallel publishes for ${#services[@]} services..."
    
    # Start all publishes in parallel
    for service in "${services[@]}"; do
        publish_service "$service" "$VERSION" &
        publish_pids+=($!)
    done
    
    # Wait for all publishes to complete and collect results
    echo "Waiting for all publishes to complete..."
    for i in "${!publish_pids[@]}"; do
        if ! wait "${publish_pids[$i]}"; then
            failed_services+=("${services[$i]}")
        fi
    done
    
    # Report results
    echo ""
    echo "=== Publish Summary ==="
    if [[ ${#failed_services[@]} -eq 0 ]]; then
        echo "✅ All ${#services[@]} services published successfully!"
    else
        echo "❌ ${#failed_services[@]} service(s) failed to publish:"
        for service in "${failed_services[@]}"; do
            echo "  - $service"
        done
        echo ""
        echo "Check individual log files (publish_*.log) for detailed error information."
        exit 1
    fi
    
    # Clean up log files on success
    for service in "${services[@]}"; do
        docker_name=$(echo $service | sed 's/\//-/g')
        rm -f "publish_${docker_name}.log"
    done
else
    # Push a single service
    echo "Pushing $SERVICE..."
    # Convert service name to Docker-compatible name
    DOCKER_NAME=$(echo $SERVICE | sed 's/\//-/g')

    # Tag images with version and latest
    docker tag triggerx-${DOCKER_NAME}:${VERSION} trigg3rx/triggerx-${DOCKER_NAME}:${VERSION}

    echo "Pushing $SERVICE..."
    docker push trigg3rx/triggerx-${DOCKER_NAME}:${VERSION}

    docker rmi trigg3rx/triggerx-${DOCKER_NAME}:${VERSION}
fi

if [[ "$SERVICE" == "all" ]]; then
    echo "All services published successfully with version ${VERSION}"
else
    echo "Successfully tagged and pushed: triggerx-${SERVICE}:${VERSION}"
fi
