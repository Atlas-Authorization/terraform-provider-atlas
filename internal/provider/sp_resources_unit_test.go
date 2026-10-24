package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Every new resource and data source must produce a valid schema, and every
// data source must be derivable (the spec data sources panic on an unsupported
// attribute type at Schema time).
func TestSP_schemasAreValid(t *testing.T) {
	ctx := context.Background()
	p := New("test")()
	for _, f := range p.Resources(ctx) {
		r := f()
		var resp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &resp)
		if d := resp.Schema.ValidateImplementation(ctx); d.HasError() {
			var m resource.MetadataResponse
			r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "atlas"}, &m)
			t.Errorf("%s: %v", m.TypeName, d)
		}
	}
	names := map[string]bool{}
	for _, f := range p.DataSources(ctx) {
		d := f()
		var m datasource.MetadataResponse
		d.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "atlas"}, &m)
		names[m.TypeName] = true
		var resp datasource.SchemaResponse
		d.Schema(ctx, datasource.SchemaRequest{}, &resp)
		if dg := resp.Schema.ValidateImplementation(ctx); dg.HasError() {
			t.Errorf("%s: %v", m.TypeName, dg)
		}
	}
	for _, want := range []string{"atlas_whoami", "atlas_secret_key", "atlas_partner_link", "atlas_integration", "atlas_audit_export", "atlas_privacy_orchestrator", "atlas_referral_settings"} {
		if !names[want] {
			t.Errorf("missing data source %s", want)
		}
	}
}

func TestSecretKey_createSendsNewFieldsAndRevokesOnDestroy(t *testing.T) {
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodPost && call.Path == "/v1/secret_keys" {
			return 201, map[string]interface{}{"id": "key_1", "prefix": "atlas_live_sk_ab", "format": "atlas", "scopes": []string{"users:read"},
				"status": "active", "owner": "ops@example.com", "expires_at": 1893456000000, "allowed_ips": []string{}, "allowed_origins": []string{"https://a.example"},
				"created_at": 1, "external_ref": "tf-1", "metadata": map[string]interface{}{"env": "prod"}, "secret": "atlas_live_sk_secret"}
		}
		return 200, map[string]interface{}{}
	})
	r := NewSecretKeyResource().(*secretKeyResource)
	configureResource(t, r, c)
	ctx := context.Background()
	var m secretKeyModel
	m.Scopes = stringSliceToSet(ctx, []string{"users:read"}, &diag.Diagnostics{})
	m.AllowedIPs = stringSliceToSet(ctx, nil, &diag.Diagnostics{})
	m.AllowedOrigins = stringSliceToSet(ctx, []string{"https://a.example"}, &diag.Diagnostics{})
	m.Metadata = stringMapToMap(ctx, map[string]string{"env": "prod"}, &diag.Diagnostics{})
	m.ID, m.Name, m.Owner, m.ExpiresAt = unkS(), nullS(), nullS(), unkS()
	m.ExternalRef = strVal("tf-1")
	m.Secret, m.Prefix, m.Environment, m.Status = unkS(), unkS(), unkS(), unkS()
	m.CreatedAt, m.LastUsedAt = unkI(), unkI()
	state, resp := doCreate(t, r, planFrom(t, r, m))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	call := api.last(http.MethodPost, "/v1/secret_keys")
	if call.Body["external_ref"] != "tf-1" || call.Body["allowed_origins"] == nil {
		t.Errorf("create body missing new fields: %+v", call.Body)
	}
	var got secretKeyModel
	state.Get(ctx, &got)
	if got.Secret.ValueString() != "atlas_live_sk_secret" {
		t.Errorf("secret not stored: %q", got.Secret.ValueString())
	}
	doDelete(t, r, state)
	if api.count(http.MethodPost, "/v1/secret_keys/key_1/revoke") != 1 {
		t.Error("destroy must revoke the key")
	}
}

func TestAPIError_newCodes(t *testing.T) {
	e := &client.APIError{Status: 403, Errors: []client.ErrorItem{{Code: "INSUFFICIENT_SCOPE", Message: "m", Meta: map[string]interface{}{"required_scope": "audit:write"}}}}
	if !e.IsInsufficientScope() || e.RequiredScope() != "audit:write" {
		t.Error("INSUFFICIENT_SCOPE not read")
	}
	rnf := &client.APIError{Status: 404, Errors: []client.ErrorItem{{Code: "ROUTE_NOT_FOUND", Message: "m", DidYouMean: []string{"/v1/x"}}}}
	if rnf.IsNotFound() || !rnf.IsRouteNotFound() {
		t.Error("ROUTE_NOT_FOUND must not look like a missing object")
	}
	if !(&client.APIError{Status: 403, Errors: []client.ErrorItem{{Code: "STEP_UP_REQUIRED"}}}).IsStepUpRequired() {
		t.Error("STEP_UP_REQUIRED")
	}
}
