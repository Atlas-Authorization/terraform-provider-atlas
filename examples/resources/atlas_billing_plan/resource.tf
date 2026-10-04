# A paid plan. slug becomes the `pla` session claim; features become `fea`.
resource "atlas_billing_plan" "pro" {
  name            = "Pro"
  slug            = "pro"
  stripe_price_id = "price_123"
  audience        = "org"
  interval        = "month"
  currency        = "usd"
  features        = ["seats", "sso", "audit_log"]
  pricing_model   = "per_seat"
  trial_days      = 14
}

# The free tier: a plan with no Stripe price (free = true).
resource "atlas_billing_plan" "free" {
  name     = "Free"
  slug     = "free"
  features = ["dashboard"]
}
