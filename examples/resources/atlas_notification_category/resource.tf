# A tenant-defined end-user notification category. `key` identifies it; `label`
# is shown on the user's preference toggle. Tenant categories are always optional.
resource "atlas_notification_category" "product_updates" {
  key   = "product_updates"
  label = "Product updates"
}

# A template gates itself against the category's preference toggle.
resource "atlas_notification_template" "release_note" {
  name     = "release_note"
  subject  = "What's new in {{product}}"
  body     = "Hi {{name}}, here is what shipped this week."
  category = atlas_notification_category.product_updates.key
}
