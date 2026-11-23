# Route SMS to a country through a provider.
resource "atlas_messaging_country_route" "example" {
  country = "tf-messaging-country-route"
  provider_id = "tf-messaging-country-route"
}
