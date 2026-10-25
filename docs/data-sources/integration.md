# atlas_integration (data source)

Read-only lookup of the matching `atlas_integration` object; secrets are never available.

The `cmp` attribute is the consent manager (with its nested `consent_api`) the object brings, or null; it includes a derived default.

The `per_call_audience` attribute is the per-call audience in effect (`param`, `template`, `allowed`, `any_matching`), including a preset's default, or null when off.
