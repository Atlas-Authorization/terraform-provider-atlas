# A cookie-consent banner on the hosted pages (any CMP).
resource "atlas_cookie_consent" "this" {
  script_url    = "https://consent.cookiebot.com/uc.js"
  script_id     = "Cookiebot"
  site_id       = "00000000-0000-0000-0000-000000000000"
  blocking_mode = "auto"

  # A "Cookie settings" link in the page footer that reopens the CMP.
  settings_link = {
    label  = "Cookie settings"
    preset = "cookiebot"
  }
}
