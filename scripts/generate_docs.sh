#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SWAGGER_FILE="$PROJECT_ROOT/docs/api/swagger.yaml"

echo "=== API Doc Generation Script ==="
echo "Project root: $PROJECT_ROOT"

if [ ! -f "$SWAGGER_FILE" ]; then
    echo "Error: swagger.yaml not found at $SWAGGER_FILE"
    exit 1
fi

echo "Base swagger spec: $SWAGGER_FILE"

# Check for swag CLI
if command -v swag &> /dev/null; then
    echo "Found swag CLI, generating docs from annotations..."

    # Generate from annotations in service handlers
    for svc in account-service auth-service payment-service notification-service config-service data-product-service; do
        SVC_DIR="$PROJECT_ROOT/$svc"
        if [ -d "$SVC_DIR" ]; then
            echo "Processing $svc..."
            swag init -g cmd/main.go -o "$PROJECT_ROOT/docs/api/$svc" --parseDependency --parseInternal 2>/dev/null || true
        fi
    done
else
    echo "swag CLI not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"
    echo "Skipping annotation-based generation."
fi

# Validate swagger spec
if command -v swagger &> /dev/null; then
    echo "Validating swagger spec..."
    swagger validate "$SWAGGER_FILE" || echo "Validation warnings found"
fi

echo "=== API Doc Generation Complete ==="
echo "Output: $PROJECT_ROOT/docs/api/"
