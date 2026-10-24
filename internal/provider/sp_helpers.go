package provider

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Helpers shared by the 2026-10 resources (secret keys, partner links,
// integrations, audit exports, status, privacy orchestrator, referrals).

// emptyStringSet is the default for an optional set that the API reads back as
// [] when unset, so an omitted attribute plans and applies to the same value.
func emptyStringSet() defaults.Set {
	return setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{}))
}

// emptyStringMap is the default for an optional map the API reads back as {}.
func emptyStringMap() defaults.Map {
	return mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{}))
}

// metadataToAPI converts the Terraform metadata map into the request value. The
// API stores strings, numbers and booleans; Terraform carries every value as a
// string, which the API keeps as a string.
func metadataToAPI(ctx context.Context, m types.Map, diags *diag.Diagnostics) map[string]interface{} {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	strs := mapToStringMap(ctx, m, diags)
	out := make(map[string]interface{}, len(strs))
	for k, v := range strs {
		out[k] = v
	}
	return out
}

// metadataFromAPI converts the API's metadata object back to a string map.
// Numbers and booleans that were stored as such are rendered as text.
func metadataFromAPI(ctx context.Context, m map[string]interface{}, diags *diag.Diagnostics) types.Map {
	out := make(map[string]string, len(m))
	for k, v := range m {
		switch t := v.(type) {
		case string:
			out[k] = t
		case bool:
			out[k] = strconv.FormatBool(t)
		case float64:
			out[k] = strconv.FormatFloat(t, 'f', -1, 64)
		case nil:
		default:
			out[k] = fmt.Sprint(t)
		}
	}
	return stringMapToMap(ctx, out, diags)
}

// msToRFC3339 renders epoch milliseconds as an RFC 3339 UTC timestamp.
func msToRFC3339(ms int64) string {
	return time.UnixMilli(ms).UTC().Format(time.RFC3339)
}

// timeFromAPI maps a nullable epoch-ms field to an RFC 3339 string, keeping the
// configured spelling when it names the same instant (so "2027-01-01T01:00:00+01:00"
// in config is not reported as drift against the API's UTC reading).
func timeFromAPI(ms *int64, prior types.String) types.String {
	if ms == nil {
		return types.StringNull()
	}
	if !prior.IsNull() && !prior.IsUnknown() {
		if t, err := time.Parse(time.RFC3339, prior.ValueString()); err == nil && t.UnixMilli() == *ms {
			return prior
		}
	}
	return types.StringValue(msToRFC3339(*ms))
}

// stringSetEqual reports whether two string slices hold the same members.
func stringSetEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x := append([]string(nil), a...)
	y := append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// strOrNull maps a plain string to types.String, with "" as null.
func strOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// ptrStr returns a pointer to s (for optional request fields).
func ptrStr(s string) *string { return &s }

// isBlank reports whether a Terraform string is null, unknown or empty.
func isBlank(v types.String) bool {
	return v.IsNull() || v.IsUnknown() || v.ValueString() == ""
}
