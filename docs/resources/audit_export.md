# atlas_audit_export

Streams the instance audit log as CloudEvents to a webhook or S3 bucket.

Attributes, defaults and replace-on-change rules are in the schema (`terraform providers schema -json`).

## Example Usage

```hcl
resource "atlas_audit_export" "siem" {
  name         = "siem"
  type         = "webhook"
  destination  = { url = "https://siem.example.com/atlas" }
  action_filter = ["user.", "organization.*"]
}
```
