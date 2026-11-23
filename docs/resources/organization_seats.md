# atlas_organization_seats

Set the seat limit (max_allowed_memberships; null = unlimited) and the warning level. Lowering the limit below the members already seated removes nobody; new seats are refused. Emits organization.seats_low when used seats reach warn_percent of the limit and organization.seat_limit_reached when the limit is reached or someone is turned away, each once per crossing.

Scopes: `organizations:read` to read, `organizations:write` to write.

Backend API: `PUT /v1/organizations/{id}/seats` (create and update); `GET /v1/organizations/{id}/seats` (read); no delete route: destroy forgets the object and leaves the setting in place.

## Example Usage

```hcl
resource "atlas_organization_seats" "example" {
  organization_id = "tf-organization-seats"
}
```

## Schema

### Required

- `organization_id` (String) The organization id this object belongs to. Changing it forces replacement.

### Optional

- `limit` (Number) The `limit` field.
- `warn_percent` (Number) The `warn_percent` field.

### Read-Only

- `id` (String) Atlas object id.
- `used` (Number) Members now.
- `pending_invitations` (Number) Unexpired pending invitations.
- `available` (Number) Seats left after members and pending invitations; null when unlimited.
- `percent_used` (Number) The `percent_used` field.
- `state` (String) The `state` field. One of: `ok`, `low`, `full`.

## Import

The import id is `<organization_id>`.

```sh
terraform import atlas_organization_seats.example <organization_id>
```
