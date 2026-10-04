package provider

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// authConfigSubsetModifier suppresses the perpetual diff on `auth_config`.
//
// `auth_config` is written as a PARTIAL patch (a `jsonencode(...)` of only the
// keys the tenant wants to set), but Atlas stores and returns the full
// server-side MERGED object. The refreshed state therefore holds a SUPERSET of
// the config literal, so a plain string comparison never byte-matches and every
// `terraform plan` shows a spurious diff.
//
// This modifier parses the planned config value and the prior STATE value as
// JSON objects. If every key present in the config is present AND deep-equal in
// the prior state (i.e. config ⊆ state), the config expresses no real change,
// so it copies the prior state value into the plan — suppressing the diff. If
// the config is NOT a subset (a genuinely changed value, or a key not yet in
// state), the plan is left untouched so the diff shows. Null/unknown config is
// skipped, and non-object JSON on either side falls back to an exact string
// comparison.
type authConfigSubsetModifier struct{}

func (authConfigSubsetModifier) Description(_ context.Context) string {
	return "Suppresses the diff when the configured auth_config is a subset of the server-merged value already in state."
}

func (m authConfigSubsetModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (authConfigSubsetModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// No prior state (e.g. on create) → there is nothing to suppress against.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	// Null/unknown config → the attribute was removed or is still being
	// computed; leave the plan for the framework to resolve.
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	configStr := req.ConfigValue.ValueString()
	stateStr := req.StateValue.ValueString()

	var configObj, stateObj map[string]json.RawMessage
	configIsObj := json.Unmarshal([]byte(configStr), &configObj) == nil && configObj != nil
	stateIsObj := json.Unmarshal([]byte(stateStr), &stateObj) == nil && stateObj != nil

	// Non-object JSON on either side → cannot reason about subsets; fall back
	// to an exact string comparison and suppress only an identical value.
	if !configIsObj || !stateIsObj {
		if configStr == stateStr {
			resp.PlanValue = req.StateValue
		}
		return
	}

	// config ⊆ state → no real change; keep the server-merged state value so the
	// plan is clean.
	if jsonObjectSubset(configObj, stateObj) {
		resp.PlanValue = req.StateValue
	}
}

// jsonObjectSubset reports whether every key in sub is present in super with a
// deep-equal value.
func jsonObjectSubset(sub, super map[string]json.RawMessage) bool {
	for k, subVal := range sub {
		superVal, ok := super[k]
		if !ok {
			return false
		}
		if !jsonDeepEqual(subVal, superVal) {
			return false
		}
	}
	return true
}

// jsonDeepEqual compares two JSON documents for structural equality, ignoring
// whitespace and object key order.
func jsonDeepEqual(a, b json.RawMessage) bool {
	var av, bv interface{}
	if err := json.Unmarshal(a, &av); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		return false
	}
	return reflect.DeepEqual(av, bv)
}
