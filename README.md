# PR Reviewer Assignment Service

## Overview
This service manages pull request reviewer assignments within teams.

## Prerequisites
- Docker and Docker Compose
- PostgreSQL client (psql)
- Go 1.24.7

## Setup

### Quick Start
1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```
2. Run the service with database:
   ```bash
   make run-with-db
   ```

### Manual Setup

#### Database Setup
1. Start the database:
   ```bash
   make start-db
   ```

2. Initialize the database:
   ```bash
   make init-db
   ```

#### Running the Service
1. Ensure the database is running and initialized
2. Run the service:
   ```bash
   make run
   ```
3. The service will be available at http://localhost:8080

## Makefile Targets
- `make run-with-db` - Start database and run application
- `make start-db` - Start PostgreSQL database
- `make init-db` - Initialize database schema
- `make run` - Run the application
- `make stop-db` - Stop database
- `make clean-db` - Clean up database (removes data)

## Testing
You can test if the service is running correctly by making a request to the health endpoint:
```bash
curl http://localhost:8080/health
```
