# Register a schema for an action: the target types and actor types it allows and a JSON Schema its metadata must satisfy.
resource "atlas_audit_event_schema" "example" {
  action = "tf-audit-event-schema"
}
