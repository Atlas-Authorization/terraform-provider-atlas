# atlas_organization_design

An organization's design override, layered over the instance design on that
organization's hosted pages (for example a customer's own colours and logo).

`design` is a JSON object (write it with `jsonencode`), compared
**semantically**. Each apply **replaces** the organization's whole override;
destroying the resource removes it, so the organization falls back to the
instance design.

API: `GET`/`PUT /v1/organizations/:id/design`.

## Example Usage

```hcl
resource "atlas_organization_design" "acme" {
  organization_id = atlas_organization.acme.id
  design = jsonencode({
    palette = { primary = "#0f766e" }
  })
}
```

## Schema

### Required

- `organization_id` (String) — The organization whose design override this manages. Changing it forces replacement.
- `design` (String, JSON object) — The override settings (any subset of the design document).

### Read-Only

- `id` (String) — Equals `organization_id`.
- `resolved` (String, JSON) — The design the organization's pages render (instance design with this override on top).
- `warnings` (String, JSON array) — Contrast and legibility warnings.

## Import

```sh
terraform import atlas_organization_design.acme org_123
```
