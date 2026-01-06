#!/bin/bash

# Load environment variables from .env file if it exists
# This ensures we use the same database configuration as the application
# See: internal/dbserver/config/config.go and pkg/env/common.go
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

# Read database credentials from environment variables
# These match the same variables used by GetDatabaseConfig() in pkg/env/common.go:
# - DATABASE_USERNAME (default: "")
# - DATABASE_PASSWORD (default: "")
# - DATABASE_HOST_ADDRESS (default: "localhost") - used by application, not needed here
# - DATABASE_HOST_PORT (default: "9042") - used by application, not needed here
# Note: The script uses docker exec to connect directly to the container, so host/port aren't needed
DATABASE_USERNAME="${DATABASE_USERNAME:-}"
DATABASE_PASSWORD="${DATABASE_PASSWORD:-}"

# Function to execute CQL script with optional authentication
run_cql() {
    local script_file=$1
    if [ -n "$DATABASE_USERNAME" ] && [ -n "$DATABASE_PASSWORD" ]; then
        docker exec -i triggerx-scylla cqlsh -u "$DATABASE_USERNAME" -p "$DATABASE_PASSWORD" < "$script_file"
    else
        docker exec -i triggerx-scylla cqlsh < "$script_file"
    fi
}

# Function to create database user if credentials are provided
create_db_user() {
    if [ -n "$DATABASE_USERNAME" ] && [ -n "$DATABASE_PASSWORD" ]; then
        echo "Attempting to create database user: $DATABASE_USERNAME"
        
        # Try to create user as superuser (cassandra/cassandra)
        # Suppress warnings but capture actual errors
        if docker exec -i triggerx-scylla cqlsh -u cassandra -p cassandra 2>/dev/null <<EOF; then
CREATE KEYSPACE IF NOT EXISTS triggerx WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
CREATE USER IF NOT EXISTS '$DATABASE_USERNAME' WITH PASSWORD '$DATABASE_PASSWORD';
GRANT ALL ON KEYSPACE triggerx TO $DATABASE_USERNAME;
EOF
            echo "Database user '$DATABASE_USERNAME' created and permissions granted"
        # If that fails, try without authentication (auth might not be enabled)
        elif docker exec -i triggerx-scylla cqlsh 2>/dev/null <<EOF; then
CREATE KEYSPACE IF NOT EXISTS triggerx WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
CREATE USER IF NOT EXISTS '$DATABASE_USERNAME' WITH PASSWORD '$DATABASE_PASSWORD';
GRANT ALL ON KEYSPACE triggerx TO $DATABASE_USERNAME;
EOF
            echo "Database user '$DATABASE_USERNAME' created and permissions granted"
        else
            # Check if user already exists by trying to connect with the credentials
            if docker exec -i triggerx-scylla cqlsh -u "$DATABASE_USERNAME" -p "$DATABASE_PASSWORD" -e "DESCRIBE KEYSPACE triggerx" 2>/dev/null >/dev/null; then
                echo "Database user '$DATABASE_USERNAME' already exists with correct permissions"
            else
                echo "Note: User creation failed (authentication may not be enabled or user creation requires manual setup)"
                echo "If authentication is enabled, create the user manually:"
                echo "  docker exec -it triggerx-scylla cqlsh -u cassandra -p cassandra"
                echo "  CREATE USER IF NOT EXISTS '$DATABASE_USERNAME' WITH PASSWORD '$DATABASE_PASSWORD';"
                echo "  GRANT ALL ON KEYSPACE triggerx TO $DATABASE_USERNAME;"
            fi
        fi
    fi
}

# Check authentication status
if [ -n "$DATABASE_USERNAME" ] && [ -n "$DATABASE_PASSWORD" ]; then
    echo "Using database authentication with username: $DATABASE_USERNAME"
else
    echo "No database credentials provided, connecting without authentication"
fi

# Create database user first (if credentials provided)
create_db_user

# Execute the CQL script
echo "Initializing database schema..."
if run_cql scripts/database/init-db.cql; then
    echo "Database schema initialized successfully"
else
    echo "Failed to initialize database schema"
    exit 1
fi

sleep 3
echo "Inserting test data into Database..."
if run_cql scripts/database/test-data.cql; then
    echo "Test data inserted successfully"
else
    echo "Failed to insert test data"
    exit 1
fi