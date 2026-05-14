#!/bin/bash
set -euo pipefail

BASE=${1:-"http://127.0.0.1:30300"}
API="$BASE/api/v1"
PASS=0
FAIL=0

ok()   { PASS=$((PASS+1)); echo "  ✅ $1"; }
fail() { FAIL=$((FAIL+1)); echo "  ❌ $1"; }

check_json() {
  local label="$1" method="$2" url="$3" exp_code="$4" extra="$5"
  local resp
  resp=$(curl -s -o /tmp/resp.json -w "%{http_code}" -X "$method" "$url" -H "Content-Type: application/json" $extra 2>/dev/null || true)
  if [ "$resp" = "$exp_code" ]; then
    ok "$label (HTTP $resp)"
  else
    fail "$label — expected $exp_code got $resp"
    cat /tmp/resp.json 2>/dev/null | head -3
  fi
}

echo "================================================"
echo " w004 Account Center — Integration Test Suite"
echo " Target: $BASE"
echo "================================================"
echo ""

echo "--- 1. Infrastructure Health Check ---"
for port in 30300 30301 30302 30311 30312 30313 30314; do
  check_json ":$port/health" GET "http://127.0.0.1:$port/health" 200 ""
done
echo ""

PHONE="138$(printf '%08d' $((RANDOM % 100000000)))"
USER_ID=""
TOKEN=""

echo "--- 2. User Registration ---"
check_json "Register user" POST "$API/account/register" 200 "-d {\"phone_number\":\"$PHONE\",\"password\":\"Test@1234\",\"sms_code\":\"000000\",\"account_id\":\"test_$(date +%s)\"}"
echo ""

echo "--- 3. User Login ---"
LOGIN_RESP=$(curl -s -X POST "$API/auth/login" -H "Content-Type: application/json" -d "{\"credential\":\"$PHONE\",\"password\":\"Test@1234\"}" 2>/dev/null || true)
LOGIN_CODE=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code',''))" 2>/dev/null || true)
TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('access_token',''))" 2>/dev/null || true)
USER_ID=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('user_id',''))" 2>/dev/null || true)
if [ -n "$TOKEN" ] && [ "$TOKEN" != "" ]; then
  ok "Login succeeded, got token"
else
  fail "Login — no token in response"
fi
echo ""

echo "--- 4. Get User Tier ---"
if [ -n "$USER_ID" ] && [ "$USER_ID" != "" ]; then
  check_json "Get user tier" GET "$API/account/$USER_ID/tier" 200 "-H \"Authorization: Bearer $TOKEN\""
else
  check_json "Get user tier" GET "$API/account/1/tier" 200 "-H \"Authorization: Bearer test\""
fi
echo ""

echo "--- 5. Entitlements ---"
if [ -n "$USER_ID" ] && [ "$USER_ID" != "" ]; then
  check_json "Get entitlements" GET "$API/entitlements/$USER_ID" 200 "-H \"Authorization: Bearer $TOKEN\""
else
  check_json "Get entitlements" GET "$API/entitlements/1" 200 "-H \"Authorization: Bearer test\""
fi
echo ""

echo "--- 6. Credit Account ---"
if [ -n "$USER_ID" ] && [ "$USER_ID" != "" ]; then
  check_json "Get credit account" GET "$API/credits/$USER_ID/account" 200 "-H \"Authorization: Bearer $TOKEN\""
else
  check_json "Get credit account" GET "$API/credits/1/account" 200 "-H \"Authorization: Bearer test\""
fi
echo ""

echo "--- 7. Referral Link Generation ---"
if [ -n "$USER_ID" ] && [ "$USER_ID" != "" ]; then
  check_json "Generate referral link" POST "$API/referral/generate-link" 200 "-H \"Authorization: Bearer $TOKEN\" -d {\"user_id\":\"$USER_ID\"}"
else
  check_json "Generate referral link" POST "$API/referral/generate-link" 200 "-H \"Authorization: Bearer test\" -d {\"user_id\":\"1\"}"
fi
echo ""

echo "--- 8. QR Code Generation ---"
check_json "Generate QR code" POST "$API/qrcode/generate" 200 '-d {"content":"test-qr-data","type":"text"}'
echo ""

echo "--- 9. KYB Submit ---"
check_json "KYB submit (no auth)" POST "$API/kyb/submit" 401 '-d {}'
echo ""

echo "--- 10. Audit Log ---"
check_json "Audit log (no auth)" POST "$API/audit/logs" 401 '-d {"action_type":"TEST","target_resource":"/test","source_ip":"127.0.0.1","result":"SUCCESS"}'
echo ""

echo "--- 11. Risk Assessment ---"
check_json "Risk assess (no auth)" POST "$API/risk/assess" 401 '-d {"user_id":"1","ip_address":"127.0.0.1"}'
echo ""

echo "--- 12. SMS Send ---"
check_json "Send SMS" POST "$API/sms/send" 200 "-d {\"phone_number\":\"$PHONE\"}"
echo ""

echo "--- 13. Email OTP ---"
check_json "Send email OTP" POST "$API/email/otp/send" 200 '-d {"email":"test@example.com"}'
echo ""

echo "--- 14. API Gateway Proxy ---"
check_json "Gateway health" GET "$BASE/health" 200 ""
check_json "Gateway metrics" GET "$BASE/metrics" 200 ""
check_json "Proxy to account service" GET "$API/account/register" 405 ""
echo ""

echo "--- 15. Subscription Purchase ---"
check_json "Purchase subscription (no auth)" POST "$API/subscriptions/purchase" 401 '-d {"plan_id":"free","payment_method":"free"}'
echo ""

echo "================================================"
TOTAL=$((PASS+FAIL))
echo " Results: $PASS passed / $FAIL failed / $TOTAL total"
if [ "$FAIL" -eq 0 ]; then
  echo " Status: ✅ ALL TESTS PASSED"
else
  echo " Status: ❌ $FAIL TESTS FAILED"
fi
echo "================================================"
