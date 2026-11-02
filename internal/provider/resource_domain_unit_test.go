package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDomain_pathMount(t *testing.T) {
	ctx := context.Background()
	domain := map[string]interface{}{
		"object": "domain", "id": "dom_1", "role": "fapi", "host": "app.example.com", "status": "pending",
		"live": false, "cname_target": nil, "method": "path", "base_path": "/__atlas", "routing": nil,
		"frontend_api_url": "https://app.example.com/__atlas",
		"verification": map[string]interface{}{
			"method": "https_file", "url": "https://app.example.com/.well-known/atlas-verify/tok_1",
			"token": "tok_1", "content": "tok_1", "instructions": "Serve the token as the whole body.",
		},
	}
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodGet {
			return 200, map[string]interface{}{"object": "list", "data": []interface{}{domain}}
		}
		return 201, domain
	})
	r := NewDomainResource()
	configureResource(t, r, c)
	plan := domainModel{
		ID: types.StringUnknown(), Role: types.StringValue("fapi"), Host: types.StringValue("app.example.com"),
		Status: types.StringUnknown(), CnameTarget: types.StringUnknown(), Live: types.BoolUnknown(),
		CookieDomain: types.StringUnknown(), FailureReason: types.StringUnknown(), CertificateExpires: types.Int64Unknown(),
		Method: types.StringValue("path"), BasePath: types.StringUnknown(), Routing: types.StringUnknown(),
		FrontendAPIURL: types.StringUnknown(), VerificationURL: types.StringUnknown(),
		VerificationToken: types.StringUnknown(), VerificationBody: types.StringUnknown(),
	}
	state, resp := doCreate(t, r, planFrom(t, r, plan))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	post := api.last(http.MethodPost, "/v1/domains")
	if post.Body["method"] != "path" {
		t.Errorf("method must be sent, got %v", post.Body)
	}
	if _, sent := post.Body["base_path"]; sent {
		t.Error("an unset base_path must be left to Atlas's default")
	}
	var m domainModel
	state.Get(ctx, &m)
	if !m.CnameTarget.IsNull() || m.BasePath.ValueString() != "/__atlas" || m.VerificationBody.ValueString() != "tok_1" ||
		m.VerificationURL.ValueString() != "https://app.example.com/.well-known/atlas-verify/tok_1" || !m.Routing.IsNull() ||
		m.FrontendAPIURL.ValueString() != "https://app.example.com/__atlas" {
		t.Fatalf("state after create = %+v", m)
	}

	// A CNAME domain from an older projection (no method) reads as cname.
	delete(domain, "method")
	domain["cname_target"] = "edge.atlasauth.net"
	domain["verification"] = nil
	read, rr := doRead(t, r, state)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	var rm domainModel
	read.Get(ctx, &rm)
	if rm.Method.ValueString() != "cname" || rm.CnameTarget.ValueString() != "edge.atlasauth.net" || !rm.VerificationToken.IsNull() {
		t.Errorf("state after read = %+v", rm)
	}
}
