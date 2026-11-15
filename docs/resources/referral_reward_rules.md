# atlas_referral_reward_rules

Replace the referral reward terms (a field you leave out returns to its default). At least one of `referrer_amount` and `referee_amount` must be above zero to turn rewards on. Rewards are credits on the billing credit ledger; they are created only for referrals Atlas has attributed and not flagged, once the trigger is met, and are granted after `hold_days`.

Scopes: `billing:read` to read, `billing:write` to write.

Backend API: `PUT /v1/referral_reward_rules` (create and update); `GET /v1/referral_reward_rules` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_referral_reward_rules" "example" {

}
```

## Schema

### Optional

- `enabled` (Boolean) The `enabled` field.
- `trigger` (String) `signup`: the referral is attributed. `first_payment`: the referee's first paid invoice.
- `referrer_amount` (Number) Credit for the referrer, minor units.
- `referee_amount` (Number) Credit for the referee, minor units.
- `currency` (String) The `currency` field.
- `hold_days` (Number) Days a reward waits before it is granted.
- `max_rewards_per_referrer` (Number) Cap on rewards for one referrer; null = no cap.

### Read-Only

- `id` (String) Atlas object id.
- `updated_at` (Number) The `updated_at` field. (epoch ms)

## Import

Import with any id; the object is read from the singleton path.

```sh
terraform import atlas_referral_reward_rules.example referral_reward_rules
```
