#!/bin/bash

set -e

# Database connection parameters for Docker environment
DB_HOST=${DB_HOST:-postgres}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-reviewer_service}
DB_USER=${DB_USER:-postgres}
DB_PASS=${DB_PASS:-postgres}

echo "Initializing database in Docker environment..."

# Set environment variables for psql
export PGPASSWORD=$DB_PASS

# Wait for database to be ready
echo "Waiting for database to be ready..."
until pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done

echo "PostgreSQL is ready!"

# Create database if it doesn't exist
echo "Creating database $DB_NAME if it doesn't exist..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME"

# Apply migrations
echo "Applying migrations..."
if [ -f "/api/migrations/001_init.up.sql" ]; then
    if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f /api/migrations/001_init.up.sql; then
        echo "Migrations applied successfully"
    else
        echo "Failed to apply migrations"
        exit 1
    fi
else
    echo "Migration file not found: /api/migrations/001_init.up.sql"
    exit 1
fi

echo "Database initialization completed!"
