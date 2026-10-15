package provider

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func oauthClientPlanModel(ctx context.Context, firstParty, skip types.Bool) oauthClientModel {
	uris, _ := types.ListValueFrom(ctx, types.StringType, []string{"https://app.example.com/callback"})
	return oauthClientModel{
		ID: types.StringUnknown(), ClientID: types.StringUnknown(), Name: types.StringValue("App"),
		RedirectURIs: uris, AllowedScopes: types.ListUnknown(types.StringType), GrantTypes: types.ListUnknown(types.StringType),
		TokenEndpointAuthMethod: types.StringUnknown(), LogoURI: types.StringNull(), FirstParty: firstParty,
		SkipLogoutConfirmation: skip, SecretPrefix: types.StringUnknown(), IsPublic: types.BoolUnknown(),
		ClientSecret: types.StringUnknown(), CreatedAt: types.Int64Unknown(), UpdatedAt: types.Int64Unknown(),
	}
}

func oauthClientView(skip bool) map[string]interface{} {
	return map[string]interface{}{
		"id": "oac_1", "client_id": "client_1", "name": "App", "logo_url": nil,
		"redirect_uris": []string{"https://app.example.com/callback"}, "allowed_scopes": []string{"openid"},
		"grant_types": []string{"authorization_code"}, "token_endpoint_auth_method": "client_secret_basic",
		"secret_prefix": "sec_", "is_public": false, "first_party": true, "skip_logout_confirmation": skip,
		"created_at": 1, "updated_at": 2,
	}
}

func TestOAuthClient_skipLogoutConfirmationSentAndRead(t *testing.T) {
	ctx := context.Background()
	skip := true
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) { return 200, oauthClientView(skip) })
	r := NewOAuthClientResource()
	configureResource(t, r, c)

	state, resp := doCreate(t, r, planFrom(t, r, oauthClientPlanModel(ctx, types.BoolValue(true), types.BoolValue(true))))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	post := api.last(http.MethodPost, "/v1/oauth_clients")
	if post == nil || post.Body["skip_logout_confirmation"] != true || post.Body["first_party"] != true {
		t.Fatalf("POST body = %+v", post)
	}
	var m oauthClientModel
	state.Get(ctx, &m)
	if !m.SkipLogoutConfirmation.ValueBool() {
		t.Error("skip_logout_confirmation not read back")
	}

	// Turning it off sends an explicit false on the PATCH.
	skip = false
	next := oauthClientPlanModel(ctx, types.BoolValue(true), types.BoolValue(false))
	next.ID = types.StringValue("oac_1")
	_, ur := doUpdate(t, r, planFrom(t, r, next), state)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	patch := api.last(http.MethodPatch, "/v1/oauth_clients/oac_1")
	if patch == nil {
		t.Fatal("update must PATCH /v1/oauth_clients/oac_1")
	}
	if v, sent := patch.Body["skip_logout_confirmation"]; !sent || v != false {
		t.Errorf("PATCH must send skip_logout_confirmation=false, got %+v", patch.Body)
	}
}

func TestOAuthClient_skipLogoutConfirmationNeedsFirstParty(t *testing.T) {
	ctx := context.Background()
	r := NewOAuthClientResource()
	cfg := func(firstParty, skip types.Bool) tfsdk.Config {
		p := planFrom(t, r, oauthClientPlanModel(ctx, firstParty, skip))
		return tfsdk.Config{Schema: p.Schema, Raw: p.Raw}
	}
	validate := func(firstParty, skip types.Bool) bool {
		var resp resource.ValidateConfigResponse
		r.(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: cfg(firstParty, skip)}, &resp)
		return resp.Diagnostics.HasError()
	}
	if !validate(types.BoolValue(false), types.BoolValue(true)) {
		t.Error("skip on an explicit third-party client must fail validation")
	}
	if validate(types.BoolValue(true), types.BoolValue(true)) || validate(types.BoolValue(false), types.BoolValue(false)) {
		t.Error("first-party + skip, or no skip, must validate")
	}
	if validate(types.BoolNull(), types.BoolValue(true)) || validate(types.BoolUnknown(), types.BoolValue(true)) {
		t.Error("an unset / unknown first_party is left to ModifyPlan")
	}

	// first_party unset: a new client is third-party → refused at plan time.
	modify := func(state tfsdk.State) (bool, string) {
		c := cfg(types.BoolNull(), types.BoolValue(true))
		req := resource.ModifyPlanRequest{Config: c, Plan: tfsdk.Plan{Schema: c.Schema, Raw: c.Raw}, State: state}
		resp := resource.ModifyPlanResponse{Plan: req.Plan}
		r.(resource.ResourceWithModifyPlan).ModifyPlan(ctx, req, &resp)
		if !resp.Diagnostics.HasError() {
			return false, ""
		}
		return true, resp.Diagnostics[0].Detail()
	}
	if failed, detail := modify(nullState(t, r)); !failed || !strings.Contains(detail, "first_party = true") {
		t.Errorf("create with first_party unset must fail with guidance, got %v %q", failed, detail)
	}
	// An existing client that is first-party in state may keep first_party unset.
	prior := oauthClientPlanModel(ctx, types.BoolValue(true), types.BoolValue(false))
	prior.ID, prior.ClientID, prior.AllowedScopes, prior.GrantTypes = types.StringValue("oac_1"), types.StringValue("c"), types.ListNull(types.StringType), types.ListNull(types.StringType)
	prior.TokenEndpointAuthMethod, prior.SecretPrefix, prior.IsPublic = types.StringValue("none"), types.StringNull(), types.BoolValue(true)
	prior.ClientSecret, prior.CreatedAt, prior.UpdatedAt = types.StringValue(""), types.Int64Value(1), types.Int64Value(1)
	if failed, _ := modify(stateFrom(t, r, prior)); failed {
		t.Error("a first-party client in state must be allowed to skip")
	}
	prior.FirstParty = types.BoolValue(false)
	if failed, _ := modify(stateFrom(t, r, prior)); !failed {
		t.Error("a third-party client in state must be refused")
	}
}

func TestOAuthClientDataSource_skipLogoutConfirmation(t *testing.T) {
	ctx := context.Background()
	c, _ := newFakeAPI(t, func(apiCall) (int, interface{}) { return 200, oauthClientView(true) })
	d := NewOAuthClientDataSource()
	d.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{ProviderData: c}, &datasource.ConfigureResponse{})
	var sr datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &sr)
	st := tfsdk.State{Schema: sr.Schema, Raw: tftypes.NewValue(sr.Schema.Type().TerraformType(ctx), nil)}
	if diags := st.SetAttribute(ctx, path.Root("id"), "oac_1"); diags.HasError() {
		t.Fatal(diags)
	}
	cfg := tfsdk.Config{Schema: sr.Schema, Raw: st.Raw}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: sr.Schema, Raw: cfg.Raw}}
	d.Read(ctx, datasource.ReadRequest{Config: cfg}, &resp)
	mustOK(t, "read", resp.Diagnostics.HasError(), resp.Diagnostics)
	var m oauthClientDataModel
	resp.State.Get(ctx, &m)
	if !m.SkipLogoutConfirmation.ValueBool() || !m.FirstParty.ValueBool() {
		t.Errorf("data source model = %+v", m)
	}
}
