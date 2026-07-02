# Provider Refund Integration Design

Date: 2026-06-15
Status: Approved for planning
Scope: Phase 6 / P0 payment and finance closure

## Goal

Upgrade refund approval from an internal-only status change into a real external payment-channel refund flow for both WeChat Pay and Alipay.

The target business flow is:

1. User requests refund for a paid order.
2. Admin approves the refund.
3. payment-service calls the matching payment provider refund API.
4. The provider refund number, provider status, and any provider error are persisted.
5. Only after provider refund success, payment-service marks the order refunded and triggers downstream reversal:
   - subscription cancellation
   - entitlement revocation through account-service
   - credit reversal

## Current State

The repository already contains provider-level refund methods:

- `WeChatPayProvider.Refund(ctx, req)`
- `AlipayProvider.Refund(ctx, req)`

However, `RefundService.ApproveRefund()` currently does not call those provider methods. It immediately changes the internal refund status to `approved`, marks the order `refunded`, and triggers subscription and credit reversal.

This means the internal system can say a refund was approved before the external payment channel has actually accepted the refund.

## Recommended Approach

Use a minimal-intrusion synchronous provider refund flow.

`RefundService.ApproveRefund()` will remain the orchestration point, but it will first call the provider refund API based on the order payment method. The internal business reversal will only continue if the provider call succeeds.

This keeps the implementation compatible with the current service shape and avoids introducing a new asynchronous refund callback subsystem in this phase.

## Data Model

Extend `refunds` with provider refund tracking fields:

- `refund_no`: internal refund number used as WeChat `out_refund_no` and Alipay `out_request_no`.
- `provider`: payment provider name, usually `wechat` or `alipay`.
- `provider_refund_id`: provider refund identifier returned by the channel.
- `provider_status`: provider refund status returned by the channel.
- `provider_error`: provider failure message when a channel call fails.
- `approved_at`: timestamp when internal approval and reversal completes.
- `failed_at`: timestamp when provider refund fails.

`model.Refund` will expose the same fields so backend APIs and the finance UI can display channel refund evidence.

## Service Boundaries

### RefundService

Responsibilities:

- Load refund and order.
- Calculate refundable amount.
- Resolve provider from order payment method.
- Generate stable `refund_no` if missing.
- Call provider refund API.
- Persist provider result.
- Continue downstream reversal only after provider refund success.
- Persist provider error and stop downstream reversal on failure.

### Provider Registry

`RefundService` will depend on a small provider lookup abstraction, not on concrete provider types. The abstraction should expose enough to resolve a `provider.PaymentProvider` by provider name.

### Repositories

`RefundRepository` will gain methods to persist provider result and provider failure. The SQL repository will update the new fields atomically for each step.

## Provider Selection

Provider name will be derived from `order.PaymentMethod`.

Initial mapping:

- values starting with `wechat` map to `wechat`
- values starting with `alipay` map to `alipay`

Unknown or empty payment methods are treated as refund failure. The failure reason is persisted in `provider_error`, and no order/subscription/credit reversal is executed.

## Success Flow

1. Admin calls approve refund endpoint.
2. `RefundService` loads refund and order.
3. Service validates the order is paid.
4. Service calculates refund amount.
5. Service generates `refund_no`, for example `RF<refundID>` or a timestamped deterministic variant.
6. Service calls provider refund with:
   - order number
   - refund number
   - total order amount
   - calculated refund amount
   - reason
7. Provider returns refund ID and status.
8. Service persists provider result.
9. Service marks refund `approved` and sets `approved_at`.
10. Service marks order `refunded`.
11. If the order is a subscription order, account-service is notified to cancel refunded-order subscription.
12. Credits are reversed if a credit service is configured.

## Failure Flow

If provider lookup or provider refund fails:

1. Service persists `status = failed`.
2. Service persists `provider_error` and `failed_at`.
3. Service returns an error to the approve endpoint.
4. Order remains paid.
5. Subscription remains active.
6. Entitlements remain active.
7. Credits are not reversed.

This prevents internal benefits from being revoked when the external refund did not happen.

If downstream reversal fails after provider refund success, the existing rollback model is not sufficient to undo a real provider refund. In this phase, the service will still return an error and keep the existing conservative behavior where possible, but provider refund evidence remains persisted. A later phase should introduce an outbox/compensation workflow for this edge case.

## API and UI Impact

The existing approve/reject endpoints remain unchanged.

Refund JSON responses gain provider evidence fields. The finance admin page can display:

- internal refund number
- provider
- provider refund ID
- provider status
- provider error
- approved/failed timestamps

No new finance page is required in this step.

## Testing Strategy

Use TDD around `RefundService`.

Required tests:

1. Provider success:
   - provider refund is called
   - provider refund ID/status are persisted
   - refund becomes approved
   - order becomes refunded
   - subscription cancellation is called for subscription orders

2. Provider failure:
   - refund becomes failed
   - provider error is persisted
   - order is not marked refunded
   - subscription cancellation is not called
   - credit reversal is not called

3. Unknown payment method:
   - refund becomes failed
   - explicit provider error is persisted
   - no downstream reversal occurs

4. SQL repository mapping:
   - new refund columns are selected and scanned correctly
   - provider result update writes the intended fields

## Out of Scope

This design intentionally excludes:

- asynchronous refund callback endpoints
- automatic refund retry button
- outbox or saga compensation framework
- multiple partial refunds for one order
- tax invoice red-letter cancellation
- production credential provisioning

These are separate follow-up closures after this synchronous refund flow is verified.

## Acceptance Criteria

The implementation is acceptable when:

- WeChat and Alipay refund providers are invoked from `ApproveRefund()`.
- Provider refund evidence is persisted in PostgreSQL.
- Provider failure prevents internal order/subscription/credit reversal.
- Tests cover success, provider failure, and unknown payment method failure.
- Existing payment-service, account-service, api-gateway tests still pass.
- web-ui build still passes.
- `docker compose config --quiet` still passes.
