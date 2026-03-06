#!/bin/bash

echo "=== Database Connection Troubleshooting ==="
echo ""

# Check if running in WSL
if grep -qi microsoft /proc/version; then
    echo "✓ Running in WSL"
    echo ""
    
    # Get Windows host IP
    echo "Your Windows host IP (from WSL perspective):"
    WINDOWS_IP=$(grep nameserver /etc/resolv.conf | awk '{print $2}')
    echo "  $WINDOWS_IP"
    echo ""
    
    echo "Common approaches to connect from WSL:"
    echo "  1. Use Windows host IP: $WINDOWS_IP"
    echo "  2. Use host.docker.internal (if Docker is running)"
    echo "  3. Use 127.0.0.1 if port forwarded to WSL localhost"
    echo ""
else
    echo "⚠ Not running in WSL"
fi

echo "Testing PostgreSQL connection with different hosts:"
echo ""

# Test with localhost
echo "1. Testing with localhost:5432..."
psql -h localhost -p 5432 -U postgres -d food_order_tracking -c "SELECT 1;" 2>&1 | head -3

# Test with Windows IP if available
if [ ! -z "$WINDOWS_IP" ]; then
    echo ""
    echo "2. Testing with $WINDOWS_IP:5432..."
    psql -h $WINDOWS_IP -p 5432 -U postgres -d food_order_tracking -c "SELECT 1;" 2>&1 | head -3
fi

# Test with host.docker.internal
echo ""
echo "3. Testing with host.docker.internal:5432..."
psql -h host.docker.internal -p 5432 -U postgres -d food_order_tracking -c "SELECT 1;" 2>&1 | head -3

echo ""
echo "=== Configuration ==="
echo "Set environment variables before running:"
echo "  export DB_HOST=<working-host>"
echo "  export DB_PORT=5432"
echo "  export DB_USER=postgres"
echo "  export DB_PASSWORD=pgtest"
echo "  export DB_NAME=food_order_tracking"
