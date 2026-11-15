# Data residency for the instance: the pinned serving region, the failover region and FAPI isolation.
resource "atlas_region" "example" {
  pin = true
}
