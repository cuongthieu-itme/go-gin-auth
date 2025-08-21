#!/bin/bash

# Migration script
set -e

# Default values
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USER:-root}
DB_PASS=${DB_PASS:-root}
DB_NAME=${DB_NAME:-authdb}

DSN="mysql://${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?multiStatements=true"
MIGRATION_PATH="internal/storage/migrations"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if migrate tool is installed
check_migrate_tool() {
    if ! command -v migrate &> /dev/null; then
        log_error "migrate tool is not installed"
        log_info "Install it with: go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        exit 1
    fi
}

# Wait for database to be ready
wait_for_db() {
    log_info "Waiting for database to be ready..."
    for i in {1..30}; do
        if mysql -h"${DB_HOST}" -P"${DB_PORT}" -u"${DB_USER}" -p"${DB_PASS}" -e "SELECT 1" >/dev/null 2>&1; then
            log_info "Database is ready!"
            return 0
        fi
        log_warn "Waiting for database... ($i/30)"
        sleep 2
    done
    log_error "Database is not ready after 60 seconds"
    exit 1
}

case "$1" in
  up)
    check_migrate_tool
    log_info "Running migrations up..."
    migrate -path $MIGRATION_PATH -database "$DSN" up
    log_info "Migrations completed successfully!"
    ;;
  down)
    check_migrate_tool
    STEPS=${2:-1}
    log_warn "Rolling back $STEPS migration(s)..."
    migrate -path $MIGRATION_PATH -database "$DSN" down $STEPS
    log_info "Rollback completed!"
    ;;
  force)
    check_migrate_tool
    if [ -z "$2" ]; then
      log_error "Please provide version: $0 force VERSION"
      exit 1
    fi
    log_warn "Force setting version to $2..."
    migrate -path $MIGRATION_PATH -database "$DSN" force $2
    log_info "Version forced to $2"
    ;;
  version)
    check_migrate_tool
    log_info "Current migration version:"
    migrate -path $MIGRATION_PATH -database "$DSN" version
    ;;
  create)
    check_migrate_tool
    if [ -z "$2" ]; then
      log_error "Please provide migration name: $0 create MIGRATION_NAME"
      exit 1
    fi
    log_info "Creating migration: $2"
    migrate create -ext sql -dir $MIGRATION_PATH -seq $2
    log_info "Migration files created successfully!"
    ;;
  wait)
    wait_for_db
    ;;
  *)
    echo "Usage: $0 {up|down|force|version|create|wait} [args]"
    echo ""
    echo "Commands:"
    echo "  up              - Run all up migrations"
    echo "  down [N]        - Run N down migrations (default: 1)"
    echo "  force VERSION   - Force set version"
    echo "  version         - Show current version"
    echo "  create NAME     - Create new migration"
    echo "  wait            - Wait for database to be ready"
    echo ""
    echo "Environment variables:"
    echo "  DB_HOST         - Database host (default: localhost)"
    echo "  DB_PORT         - Database port (default: 3306)"
    echo "  DB_USER         - Database user (default: root)"
    echo "  DB_PASS         - Database password (default: root)"
    echo "  DB_NAME         - Database name (default: authdb)"
    exit 1
    ;;
esac
