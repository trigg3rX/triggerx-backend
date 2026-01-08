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
    $0 -n <service> -v <version> [-w <network>]
    $0 -h|--help

OPTIONS:
    -n, --service    Service name to build (required)
    -v, --version    Version tag for the Docker image (required, format: MAJOR.MINOR.PATCH)
    -w, --network    Network name for keeper service (required for keeper, ignored for other services)
                     Valid values: mainnet, imua, sepolia
    -h, --help       Display this help message

AVAILABLE SERVICES:
    keeper              - TriggerX Keeper service (uses docker/Dockerfile.keeper)
    dbserver            - Database server service
    health              - Health monitoring service
    taskdispatcher      - Task dispatcher service
    taskmonitor         - Task monitor service
    eventmonitor        - Event monitor service
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
    echo "Usage: $0 -n <service> -v <version> [-w <network>]"
    echo "For keeper service: $0 -n keeper -v <version> -w <network>"
    echo "Use '$0 -h' for detailed help and examples"
    exit 1
}

# Function to get YAML file name from docker name
get_yaml_filename() {
    local docker_name=$1
    # Map docker names to YAML file names
    # For most services, they match directly
    # For scheduler services, handle special cases if needed
    case "$docker_name" in
        schedulers-time)
            # Check if schedulers-time.yaml exists, otherwise use time-scheduler.yaml
            if [ -f "config/services/schedulers-time.yaml" ]; then
                echo "schedulers-time.yaml"
            else
                echo "time-scheduler.yaml"
            fi
            ;;
        schedulers-condition)
            # Check if schedulers-condition.yaml exists, otherwise use condition-scheduler.yaml
            if [ -f "config/services/schedulers-condition.yaml" ]; then
                echo "schedulers-condition.yaml"
            else
                echo "condition-scheduler.yaml"
            fi
            ;;
        *)
            echo "${docker_name}.yaml"
            ;;
    esac
}

# Function to update version in YAML file
update_yaml_version() {
    local docker_name=$1
    local version=$2
    local yaml_filename=$(get_yaml_filename "$docker_name")
    local yaml_file="config/services/${yaml_filename}"
    
    # Check if YAML file exists
    if [ ! -f "$yaml_file" ]; then
        echo "Warning: YAML file $yaml_file not found, skipping version update" 1>&2
        return 1
    fi
    
    # Use awk to update or add version section
    awk -v version="$version" 'BEGIN {
        in_version = 0
        version_updated = 0
    }
    {
        if (/^version:/) {
            in_version = 1
            print $0
            next
        }
        if (in_version && /^  version:/) {
            printf "  version: \"%s\"                  # Service version\n", version
            version_updated = 1
            in_version = 0
            next
        }
        if (in_version && /^[^ ]/) {
            printf "  version: \"%s\"                  # Service version\n", version
            version_updated = 1
            in_version = 0
            print $0
            next
        }
        if (in_version && /^$/) {
            print $0
            next
        }
        if (in_version) {
            in_version = 0
        }
        print $0
    }
    END {
        if (!version_updated) {
            print ""
            print "version:"
            printf "  version: \"%s\"                  # Service version\n", version
        }
    }' "$yaml_file" > "${yaml_file}.tmp" && mv "${yaml_file}.tmp" "$yaml_file"
    
    if [ $? -eq 0 ]; then
        echo "Updated version in $yaml_file to $version"
    else
        echo "Error: Failed to update version in $yaml_file" 1>&2
        return 1
    fi
}

# Function to update network in keeper.yaml file
update_keeper_network() {
    local network=$1
    local yaml_file="config/services/keeper.yaml"
    
    # Check if YAML file exists
    if [ ! -f "$yaml_file" ]; then
        echo "Error: Keeper YAML file $yaml_file not found" 1>&2
        return 1
    fi
    
    # Validate network value
    if [[ ! "$network" =~ ^(mainnet|imua|sepolia)$ ]]; then
        echo "Error: Invalid network value. Must be one of: mainnet, imua, sepolia" 1>&2
        return 1
    fi
    
    # Use awk to update or add network field
    awk -v network="$network" 'BEGIN {
        network_updated = 0
    }
    {
        if (/^network:/) {
            printf "network: \"%s\"                  # Network: mainnet, imua, or sepolia\n", network
            network_updated = 1
            next
        }
        print $0
    }
    END {
        if (!network_updated) {
            # If network field doesn't exist, add it at the beginning after comments
            print "network: \"" network "\"                  # Network: mainnet, imua, or sepolia"
        }
    }' "$yaml_file" > "${yaml_file}.tmp" && mv "${yaml_file}.tmp" "$yaml_file"
    
    if [ $? -eq 0 ]; then
        echo "Updated network in $yaml_file to $network"
    else
        echo "Error: Failed to update network in $yaml_file" 1>&2
        return 1
    fi
}

# Parse command-line arguments
while getopts ":n:v:w:h-:" opt; do
    case ${opt} in
        n )
            SERVICE=$OPTARG
            ;;
        v )
            VERSION=$OPTARG
            ;;
        w )
            NETWORK=$OPTARG
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
                network=* )
                    NETWORK="${OPTARG#*=}"
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
if [[ ! "$SERVICE" =~ ^(keeper|dbserver|health|taskdispatcher|taskmonitor|eventmonitor|schedulers/time|schedulers/condition|all)$ ]]; then
    echo "Error: Invalid service. Allowed services are: keeper, dbserver, health, taskdispatcher, taskmonitor, eventmonitor, schedulers/time, schedulers/condition" 1>&2
    exit 1
fi

# Validate version format (basic regex for semantic versioning)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Use MAJOR.MINOR.PATCH (e.g., 0.0.1)" 1>&2
    exit 1
fi

if [[ "$SERVICE" == "all" ]]; then
    # Build all services in parallel
    services=(dbserver health taskdispatcher taskmonitor eventmonitor schedulers/time schedulers/condition)
    build_pids=()
    failed_services=()
    
    # Function to build a single service
    build_service() {
        local service=$1
        local version=$2
        local docker_name=$(echo $service | sed 's/\//-/g')
        
        # Update version in YAML file BEFORE building
        update_yaml_version "$docker_name" "$version"
        
        echo "[$(date '+%H:%M:%S')] Starting build for $service..."
        if docker build --no-cache \
            -f docker/Dockerfile.backend \
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
    echo "Successfully built all services: ${VERSION}"
    exit 0
elif [[ "$SERVICE" == "keeper" ]]; then
    # Update network in YAML file BEFORE building (mandatory for keeper)
    update_keeper_network "$NETWORK"
    
    # Update version in YAML file BEFORE building
    update_yaml_version "keeper" "$VERSION"
    
    echo "Building $SERVICE with network=$NETWORK..."
    if docker build --no-cache \
        -f docker/Dockerfile.keeper \
        -t triggerx-keeper:${VERSION} .; then
        echo "Successfully built keeper:${VERSION} (network: $NETWORK)"
        echo "Successfully built: ${VERSION}"
        exit 0
    else
        exit 1
    fi
else
    # Convert service name to Docker-compatible name
    DOCKER_NAME=$(echo $SERVICE | sed 's/\//-/g')

    # Update version in YAML file BEFORE building
    update_yaml_version "$DOCKER_NAME" "$VERSION"
    
    echo "Building $SERVICE..."
    if docker build --no-cache \
        -f docker/Dockerfile.backend \
        --build-arg SERVICE=${SERVICE} \
        --build-arg DOCKER_NAME=${DOCKER_NAME} \
        -t triggerx-${DOCKER_NAME}:${VERSION} .; then
        echo "Successfully built ${DOCKER_NAME}:${VERSION}"
        echo "Successfully built: ${VERSION}"
        exit 0
    else
        exit 1
    fi
fi
