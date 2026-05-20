#!/usr/bin/env bash
set -euo pipefail

# Dependency vulnerability scanning using Trivy and govulncheck
echo "=== Security Dependency Scan ==="
echo "Date: $(date)"
echo ""

SERVICES=(
    "api-gateway"
    "account-service"
    "auth-service"
    "credit-service"
    "notification-service"
    "compliance-service"
    "data-product-service"
    "payment-service"
)

REPORT_DIR="${REPORT_DIR:-reports/security}"
mkdir -p "${REPORT_DIR}"

# Check if tools are installed
check_tool() {
    if ! command -v "$1" &>/dev/null; then
        echo "WARNING: $1 not installed. Install: $2"
        return 1
    fi
    return 0
}

echo "--- Go Vulnerability Check (govulncheck) ---"
if check_tool govulncheck "go install golang.org/x/vuln/cmd/govulncheck@latest"; then
    for svc in "${SERVICES[@]}"; do
        if [ -d "$svc" ]; then
            echo "Scanning $svc..."
            govulncheck "./$svc/..." 2>&1 | tee "${REPORT_DIR}/${svc}-vulncheck.txt" || true
        fi
    done
fi

echo ""
echo "--- Trivy Container Scan ---"
if check_tool trivy "https://trivy.dev"; then
    for svc in "${SERVICES[@]}"; do
        IMAGE="account-center/${svc}:latest"
        if docker image inspect "$IMAGE" &>/dev/null 2>&1; then
            echo "Scanning container $IMAGE..."
            trivy image --format table --output "${REPORT_DIR}/${svc}-trivy.txt" "$IMAGE" || true
        fi
    done
fi

echo ""
echo "--- Trivy Filesystem Scan ---"
if check_tool trivy ""; then
    for svc in "${SERVICES[@]}"; do
        if [ -d "$svc" ]; then
            echo "Scanning filesystem $svc..."
            trivy fs --format table --output "${REPORT_DIR}/${svc}-fs-trivy.txt" "$svc" || true
        fi
    done
fi

echo ""
echo "=== Scan Complete ==="
echo "Reports saved to: ${REPORT_DIR}/"
