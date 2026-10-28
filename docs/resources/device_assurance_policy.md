# atlas_device_assurance_policy

Create a device assurance policy: per-platform minimum OS version (or "newest major minus N"), disk encryption, screen lock, secure hardware, jailbreak/root, firewall and EDR checks, and a grace period. A device must pass every enabled policy for its platform; its `compliant` flag, and so the `deviceTrust` access-policy condition, follows the verdict. Existing devices are re-evaluated.

Scopes: `device_assurance:read` to read, `device_assurance:write` to write.

Backend API: `POST /v1/device_assurance_policies` (create); `GET /v1/device_assurance_policies/{id}` (read); `PATCH /v1/device_assurance_policies/{id}` (update); `DELETE /v1/device_assurance_policies/{id}` (destroy).

## Example Usage

```hcl
resource "atlas_device_assurance_policy" "example" {

}
```

## Schema

### Optional

- `name` (String) The `name` field.
- `platform` (String) The `platform` field. One of: `any`, `macos`, `windows`, `linux`, `ios`, `android`, `chromeos`.
- `enabled` (Boolean) The `enabled` field.
- `min_os_version` (String) The `min_os_version` field.
- `os_latest_minus` (Number) The `os_latest_minus` field.
- `require_disk_encryption` (Boolean) The `require_disk_encryption` field.
- `require_screen_lock` (Boolean) The `require_screen_lock` field.
- `require_secure_hardware` (Boolean) The `require_secure_hardware` field.
- `block_jailbroken` (Boolean) The `block_jailbroken` field.
- `require_firewall` (Boolean) The `require_firewall` field.
- `require_edr` (Boolean) The `require_edr` field.
- `grace_hours` (Number) The `grace_hours` field.

### Read-Only

- `id` (String) Atlas object id.
- `created_at` (Number) Created (epoch ms).
- `updated_at` (Number) Updated (epoch ms).

## Import

The import id is `<id>`.

```sh
terraform import atlas_device_assurance_policy.example <id>
```
