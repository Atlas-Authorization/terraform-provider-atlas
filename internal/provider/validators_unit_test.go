package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func runStringValidator(v validator.String, value string) bool {
	req := validator.StringRequest{Path: path.Root("x"), ConfigValue: types.StringValue(value)}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	return !resp.Diagnostics.HasError()
}

func TestRoleKeyValidator(t *testing.T) {
	for key, ok := range map[string]bool{
		"org:billing_admin": true,
		"org:ab":            true,
		"billing_admin":     false, // missing org: prefix (the API rejects it)
		"org:Billing":       false,
		"org:1abc":          false,
		"org:a":             false,
	} {
		if got := runStringValidator(roleKeyValidator{}, key); got != ok {
			t.Errorf("role key %q: valid=%v, want %v", key, got, ok)
		}
	}
}

func TestSigningAlgValidator(t *testing.T) {
	if !runStringValidator(signingAlgValidator{}, "RS256") {
		t.Error("RS256 must be accepted")
	}
	if runStringValidator(signingAlgValidator{}, "ES256") {
		t.Error("ES256 must be rejected: the API never honours it")
	}
	if got := resourceServerSigningAlg(""); got != "RS256" {
		t.Errorf("empty API value must map to RS256, got %q", got)
	}
}

func TestAddDroppedPermissionsError(t *testing.T) {
	var diags diag.Diagnostics
	addDroppedPermissionsError([]string{"org:a:b"}, []string{"org:a:b"}, &diags)
	if diags.HasError() {
		t.Fatal("no error expected when every permission was stored")
	}
	addDroppedPermissionsError([]string{"org:a:b", "org:z:y", "org:c:d"}, []string{"org:a:b"}, &diags)
	if !diags.HasError() || !strings.Contains(diags[0].Detail(), "org:c:d, org:z:y") {
		t.Fatalf("expected an error naming the dropped keys, got %v", diags)
	}
}
