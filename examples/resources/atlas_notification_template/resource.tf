# A tenant notification template. POST /v1/notifications resolves it by `name`
# ahead of the built-ins, substituting {{variable}} placeholders from `data`.
resource "atlas_notification_template" "new_device" {
  name     = "new_device"
  subject  = "New sign-in on {{device}}"
  body     = "Hi {{name}}, we noticed a sign-in from {{device}} at {{time}}."
  category = "new_device"
}
