package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// pathSecretKey names the provider secret_key attribute for diagnostics.
func pathSecretKey() path.Path { return path.Root("secret_key") }

// clientFromProviderData resolves the configured *client.Client that the
// provider stashed in ResourceData/DataSourceData. It tolerates a nil (the
// framework calls Configure with nil data during early graph walks).
func clientFromProviderData(providerData any, diags *diag.Diagnostics) *client.Client {
	if providerData == nil {
		return nil
	}
	c, ok := providerData.(*client.Client)
	if !ok {
		diags.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.Client, got %T. This is a provider bug.", providerData),
		)
		return nil
	}
	return c
}

// ── Terraform <-> Go conversions ────────────────────────────────────────────

// stringSliceToList converts a Go []string into a types.List of strings.
func stringSliceToList(ctx context.Context, values []string, diags *diag.Diagnostics) types.List {
	if values == nil {
		values = []string{}
	}
	list, d := types.ListValueFrom(ctx, types.StringType, values)
	diags.Append(d...)
	return list
}

// listToStringSlice converts a types.List of strings into a Go []string. A null
// or unknown list yields nil so a create omits the field entirely.
func listToStringSlice(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var out []string
	diags.Append(list.ElementsAs(ctx, &out, false)...)
	return out
}

// stringSliceToSet converts a Go []string into a types.Set of strings.
func stringSliceToSet(ctx context.Context, values []string, diags *diag.Diagnostics) types.Set {
	if values == nil {
		values = []string{}
	}
	set, d := types.SetValueFrom(ctx, types.StringType, values)
	diags.Append(d...)
	return set
}

// setToStringSlice converts a types.Set of strings into a Go []string.
func setToStringSlice(ctx context.Context, set types.Set, diags *diag.Diagnostics) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var out []string
	diags.Append(set.ElementsAs(ctx, &out, false)...)
	return out
}

// mapToStringMap converts a types.Map of strings into a Go map[string]string.
func mapToStringMap(ctx context.Context, m types.Map, diags *diag.Diagnostics) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	out := map[string]string{}
	diags.Append(m.ElementsAs(ctx, &out, false)...)
	return out
}

// stringMapToMap converts a Go map[string]string into a types.Map of strings.
func stringMapToMap(ctx context.Context, values map[string]string, diags *diag.Diagnostics) types.Map {
	if values == nil {
		values = map[string]string{}
	}
	m, d := types.MapValueFrom(ctx, types.StringType, values)
	diags.Append(d...)
	return m
}

// optionalString returns nil for a null/unknown value, else a *string.
func optionalString(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// stringPtrToValue maps a *string (a nullable API field) onto types.String.
func stringPtrToValue(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

// int64PtrToValue maps a *int64 onto types.Int64.
func int64PtrToValue(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

// optionalInt64 returns nil for a null/unknown value, else a *int64.
func optionalInt64(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

// optionalBool returns nil for a null/unknown value, else a *bool.
func optionalBool(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

// boolPtrToValue maps a *bool onto types.Bool.
func boolPtrToValue(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*v)
}

// canonicalJSON re-encodes a JSON document into a compact, key-sorted form so a
// value stored in state is stable regardless of the input's whitespace or key
// order (Go's json.Marshal sorts map keys). Terraform's own jsonencode() emits
// the same compact, sorted form, so a round-trip does not thrash the plan.
func canonicalJSON(b []byte) (string, error) {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return "", err
	}
	out, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// jsonRawToValue maps a freeform json.RawMessage API field onto a canonical JSON
// string attribute. An empty/absent message becomes null; an unparseable one is
// surfaced verbatim rather than dropped.
func jsonRawToValue(raw json.RawMessage) types.String {
	if len(raw) == 0 || string(raw) == "null" {
		return types.StringNull()
	}
	canon, err := canonicalJSON(raw)
	if err != nil {
		return types.StringValue(string(raw))
	}
	return types.StringValue(canon)
}

// stringToJSONRaw parses a JSON-string attribute into a json.RawMessage. A
// null/unknown/empty value yields nil so the field is omitted from a PATCH; a
// non-null value must be valid JSON or the (reported) error fails the apply.
func stringToJSONRaw(v types.String) (json.RawMessage, error) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	s := strings.TrimSpace(v.ValueString())
	if s == "" {
		return nil, nil
	}
	canon, err := canonicalJSON([]byte(s))
	if err != nil {
		return nil, err
	}
	return json.RawMessage(canon), nil
}

// floatMapToPtr converts a types.Map of numbers into a *map[string]float64. A
// null/unknown map yields nil so the field is omitted; an empty map yields a
// pointer to an empty map so it is sent as {} (clearing the server value).
func floatMapToPtr(ctx context.Context, m types.Map, diags *diag.Diagnostics) *map[string]float64 {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	out := map[string]float64{}
	diags.Append(m.ElementsAs(ctx, &out, false)...)
	return &out
}

// floatMapToMap converts a Go map[string]float64 into a types.Map of numbers.
func floatMapToMap(ctx context.Context, values map[string]float64, diags *diag.Diagnostics) types.Map {
	if values == nil {
		values = map[string]float64{}
	}
	m, d := types.MapValueFrom(ctx, types.Float64Type, values)
	diags.Append(d...)
	return m
}
