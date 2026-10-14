# atlas_instance_design

The design of your hosted sign-in pages for the instance (a singleton): preset,
colours (light and dark), typography, shape, card, layout, buttons, inputs,
background and logo — the same document the dashboard design editor saves.

`design` is a JSON object (write it with `jsonencode`), compared
**semantically**: the key order, formatting and `null` members of what Atlas
returns never cause drift. (Editing only the whitespace or key order of a hand-written JSON string still shows one
in-place update, because Terraform compares configuration text; `jsonencode` output
is canonical, so it never does.)
Each apply **replaces** the whole stored design; destroying the resource removes
it, and the pages return to the default look. Unknown keys are refused with
their full path, and images must be uploaded to the instance first.

API: `GET`/`PUT /v1/instance/design`.

## Example Usage

```hcl
resource "atlas_instance_design" "this" {
  design = jsonencode({
    preset      = "midnight"
    colorScheme = "auto"
    palette = {
      primary = "#4f46e5"
      accent  = "#22d3ee"
    }
    shape = { radius = 12 }
  })
}

output "resolved_design" {
  value = jsondecode(atlas_instance_design.this.resolved)
}
```

## Schema

### Required

- `design` (String, JSON object) — The design settings. See the design reference in the Atlas docs for every section and value.

### Read-Only

- `id` (String) — Always `instance`.
- `resolved` (String, JSON) — The fully-resolved design every page renders (defaults, then the preset, then your settings).
- `warnings` (String, JSON array) — Contrast and legibility warnings for the resolved design.

## Import

```sh
terraform import atlas_instance_design.this instance
```
