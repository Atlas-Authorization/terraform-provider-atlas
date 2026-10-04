package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// schemaOf calls a resource's Schema hook and fails the test on any diagnostic,
// returning the produced schema for attribute assertions. This is a fully
// offline check — no provider configuration or live API is involved.
func schemaOf(t *testing.T, r resource.Resource) rschema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %+v", resp.Diagnostics)
	}
	return resp.Schema
}

// typeNameOf drives the Metadata hook the way the framework does, under the
// "atlas" provider type name.
func typeNameOf(t *testing.T, r resource.Resource) string {
	t.Helper()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "atlas"}, &resp)
	return resp.TypeName
}

func TestNewResourcesRegistered(t *testing.T) {
	want := map[string]bool{
		"atlas_instance_config":     false,
		"atlas_redirect_url":        false,
		"atlas_billing_plan":        false,
		"atlas_organization_policy": false,
	}
	for _, factory := range New("test")().Resources(context.Background()) {
		name := typeNameOf(t, factory())
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("resource %q is not registered in provider.Resources()", name)
		}
	}
}

func TestInstanceConfigSchema(t *testing.T) {
	s := schemaOf(t, NewInstanceConfigResource())
	if _, ok := s.Attributes["allowed_origins"]; !ok {
		t.Error("missing allowed_origins")
	}
	ac, ok := s.Attributes["auth_config"]
	if !ok {
		t.Fatal("missing auth_config")
	}
	if !ac.IsOptional() || !ac.IsComputed() {
		t.Error("auth_config should be Optional+Computed")
	}
	resolved, ok := s.Attributes["auth_config_resolved"]
	if !ok || !resolved.IsComputed() || resolved.IsRequired() {
		t.Error("auth_config_resolved should be read-only Computed")
	}
	if id := s.Attributes["id"]; id == nil || !id.IsComputed() {
		t.Error("id should be Computed")
	}
}

func TestRedirectURLSchema(t *testing.T) {
	s := schemaOf(t, NewRedirectURLResource())
	u, ok := s.Attributes["url"]
	if !ok || !u.IsRequired() {
		t.Error("url should be Required")
	}
	if id := s.Attributes["id"]; id == nil || !id.IsComputed() {
		t.Error("id should be Computed")
	}
}

func TestBillingPlanSchema(t *testing.T) {
	s := schemaOf(t, NewBillingPlanResource())
	for _, req := range []string{"name", "slug"} {
		a, ok := s.Attributes[req]
		if !ok || !a.IsRequired() {
			t.Errorf("%s should be Required", req)
		}
	}
	for _, oc := range []string{"audience", "interval", "currency", "pricing_model", "active", "features"} {
		a, ok := s.Attributes[oc]
		if !ok || !a.IsOptional() || !a.IsComputed() {
			t.Errorf("%s should be Optional+Computed", oc)
		}
	}
	if f := s.Attributes["free"]; f == nil || !f.IsComputed() || f.IsRequired() {
		t.Error("free should be read-only Computed")
	}
}

func TestOrganizationPolicySchema(t *testing.T) {
	s := schemaOf(t, NewOrganizationPolicyResource())
	org, ok := s.Attributes["organization_id"]
	if !ok || !org.IsRequired() {
		t.Error("organization_id should be Required")
	}
	for _, oc := range []string{
		"require_mfa", "sso_required", "session_idle_override_ms",
		"allowed_sign_in_methods", "ip_allowlist", "max_session_age_seconds",
	} {
		a, ok := s.Attributes[oc]
		if !ok || !a.IsOptional() || !a.IsComputed() {
			t.Errorf("%s should be Optional+Computed", oc)
		}
	}
}
