# atlas_user_legal_hold

Place a legal hold on a user. While it is active no path erases them: the delete call, a scheduled deletion (it stays scheduled and does not run), a data subject request, a workflow and a bulk delete all refuse or wait, and a merge that would delete this user is refused (409 LEGAL_HOLD). Placing a hold on a user already held returns the existing hold (200). Audited; emits user.legal_hold_placed. Honours Idempotency-Key.

Scopes: `data_subject_requests:read` to read, `data_subject_requests:write` to write.

Backend API: `POST /v1/users/{id}/legal_hold` (create and update); `GET /v1/users/{id}/legal_hold` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_user_legal_hold" "example" {
  user_id = "tf-user-legal-hold"
  reason = "tf-user-legal-hold"
}
```

## Schema

### Required

- `user_id` (String) The user id this object belongs to. Changing it forces replacement.
- `reason` (String) The `reason` field.

### Optional

- `case_ref` (String) Your own reference for the matter.

### Read-Only

- `id` (String) Atlas object id.
- `on_hold` (Boolean) The `on_hold` field.
- `active` (Boolean) The `active` field.
- `history` (String, JSON) The `history` field. A JSON document (use `jsonencode`).
- `placed_by` (String) The `placed_by` field.
- `placed_at` (Number) Epoch ms.
- `released_at` (Number) Epoch ms.
- `released_by` (String) The `released_by` field.
- `release_reason` (String) The `release_reason` field.

## Import

The import id is `<user_id>`.

```sh
terraform import atlas_user_legal_hold.example <user_id>
```
