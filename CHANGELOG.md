# Changelog

All notable changes to `terraform-provider-atlas` are documented here. The
released version is set from the git tag at build time (see `.goreleaser.yml`);
tag a release `vX.Y.Z` to match an entry below.

## 0.3.0

### Fixed

- **`atlas_billing_plan.amount` documentation corrected.** The docs claimed
  `amount` was the "smallest currency unit (cents)". Atlas stores `amount` as an
  opaque free-form string verbatim — it is not parsed, validated or converted.
  The attribute, resource docs and client comment now say so and tell you to use
  whatever units you display (e.g. `12.00` for $12, or `1200` for cents).
- **`atlas_instance_config.auth_config` no longer shows a perpetual diff.**
  `auth_config` is a partial patch, but Atlas returns the full server-merged
  object, so the refreshed state never byte-matched the config literal. A new
  plan modifier suppresses the diff when your config is a subset of the merged
  value already in state, and still shows a diff for a genuine change.

### Added

- **`atlas_instance_config.auth_config_resolved`** (computed, read-only) exposes
  the full effective configuration Atlas resolved from your `auth_config` plus
  defaults, while `auth_config` continues to hold your partial patch.
