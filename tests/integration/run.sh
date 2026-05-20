#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=== Account Center Integration Tests ==="
echo "Project root: $PROJECT_ROOT"

cleanup() {
    echo ""
    echo "Tearing down test environment..."
    docker compose \
        -f "$PROJECT_ROOT/docker-compose.yml" \
        -f "$PROJECT_ROOT/tests/integration/docker-compose.test.yml" \
        down -v --remove-orphans 2>/dev/null || true
    echo "Cleanup complete."
}

trap cleanup EXIT

echo "Building and starting services..."
docker compose \
    -f "$PROJECT_ROOT/docker-compose.yml" \
    -f "$PROJECT_ROOT/tests/integration/docker-compose.test.yml" \
    up -d --build --wait

echo ""
echo "Waiting for services to be healthy..."
sleep 5

echo ""
echo "Running integration tests..."
docker compose \
    -f "$PROJECT_ROOT/docker-compose.yml" \
    -f "$PROJECT_ROOT/tests/integration/docker-compose.test.yml" \
    run --rm integration-test

echo ""
echo "=== Integration tests completed ==="
