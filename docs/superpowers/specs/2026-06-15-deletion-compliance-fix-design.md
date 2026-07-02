# NF-01 Account Deletion Compliance Fix Design

Date: 2026-06-15
Status: Approved for implementation
Scope: Phase 6 / P0 — account deletion compliance gaps

## Gaps Being Fixed

1. **Redis session cleanup no-op** — wrong key pattern. Fix: delete `user_sessions:<userID>` set + call auth-service `POST /sessions/invalidate-all`.
2. **Entitlement cache cleanup no-op** — wrong key pattern. Fix: delete `entitlement:<userID>`.
3. **No audit record** — Fix: write to `audit_logs` table with SM3 hash chain after deletion.
4. **Same-DB PII not anonymized** — Fix: anonymize `enterprises` table KYC fields in same transaction.
5. **No DB transaction** — Fix: wrap anonymize + entitlement delete + enterprise anonymize in `BeginTx`.
6. **email_otp verification bypass** — Fix: reject if verification code empty or no verification mechanism configured.

## Out of Scope

- Cross-service credit/notification PII cleanup (Phase 7 outbox event design)
- compliance-service Kafka audit pipeline (Phase 7)
- subscriptions/credit_transactions history retention (financial compliance requires keeping)

## audit_logs Table Schema (already exists)

```sql
CREATE TABLE audit_logs (
    log_id VARCHAR(100) PRIMARY KEY,
    user_id BIGINT,
    event_time TIMESTAMP WITH TIME ZONE NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    target_resource VARCHAR(200),
    source_ip VARCHAR(50),
    result VARCHAR(20) NOT NULL,
    details JSONB,
    sm3_hash VARCHAR(128) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Acceptance Criteria

- Redis cleanup uses correct key patterns
- auth-service session invalidation is called via HTTP
- audit_logs record is written after successful deletion
- enterprises PII is anonymized in same transaction
- DB transaction wraps all DB writes
- email_otp without configured verifier is rejected
- All existing tests pass, new tests cover each fix
- `go test ./...` PASS for account-service
