package provider

import (
	"context"
	"fmt"

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
