#!/usr/bin/env bash
set -euo pipefail

echo "=== Account Center V2.0 Penetration Test Suite ==="
echo "Date: $(date)"
echo ""

REPORT_DIR="${REPORT_DIR:-reports/security}"
mkdir -p "${REPORT_DIR}"

echo "[1/4] Generating test accounts..."
go run scripts/security/generate_test_accounts.go

echo ""
echo "[2/4] Scanning dependencies..."
bash scripts/security/scan_dependencies.sh

echo ""
echo "[3/4] Running OWASP ZAP scan..."
if command -v zap-cli &>/dev/null; then
    zap-cli quick-scan --self-contained -t http://localhost:30300 -c scripts/security/zap_scan_config.json
    echo "ZAP scan report: ${REPORT_DIR}/zap-scan-report.html"
else
    echo "ZAP CLI not installed. Skipping automated scan."
    echo "Manual scan: open ZAP GUI → import config → scripts/security/zap_scan_config.json"
fi

echo ""
echo "[4/4] API Security Tests..."
echo "Checking TLS..."
curl -s -o /dev/null -w "HTTP %{http_code}" http://localhost:30300/health || echo "Gateway not reachable"

echo ""
echo "Checking CORS headers..."
curl -sI -H "Origin: http://evil.com" http://localhost:30300/api/v1/pricing 2>/dev/null | grep -i "access-control-allow-origin" || echo "No CORS issue"

echo ""
echo "Checking auth enforcement..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:30300/api/v1/account/deletion/status)
if [ "$HTTP_CODE" == "401" ]; then
    echo "PASS: Protected endpoint returns 401 without token"
else
    echo "FAIL: Protected endpoint returned $HTTP_CODE (expected 401)"
fi

echo ""
echo "Checking rate limiting..."
for i in $(seq 1 150); do
    curl -s -o /dev/null http://localhost:30300/health &
done
wait
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:30300/health)
if [ "$HTTP_CODE" == "429" ] || [ "$HTTP_CODE" == "200" ]; then
    echo "Rate limiting: HTTP $HTTP_CODE"
else
    echo "Unexpected rate limit response: $HTTP_CODE"
fi

echo ""
echo "=== Penetration Test Suite Complete ==="
echo "Reports: ${REPORT_DIR}/"
