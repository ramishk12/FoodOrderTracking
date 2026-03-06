#!/bin/bash

# Default configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-pgtest}"
DB_NAME="${DB_NAME:-food_order_tracking}"

echo "=== Food Order Tracking API ==="
echo "Database Configuration:"
echo "  Host:     $DB_HOST"
echo "  Port:     $DB_PORT"
echo "  User:     $DB_USER"
echo "  Database: $DB_NAME"
echo ""
echo "To use different settings, set environment variables before running:"
echo "  export DB_HOST=<your-ip>"
echo "  export DB_PORT=<port>"
echo "  etc."
echo ""

# Export for the Go app
export DB_HOST
export DB_PORT
export DB_USER
export DB_PASSWORD
export DB_NAME

# Build if needed
if [ ! -f "./food-order-tracker" ]; then
    echo "Building application..."
    go build -o food-order-tracker ./cmd/main.go || exit 1
fi

echo "Starting API server..."
./food-order-tracker
