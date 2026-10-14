package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func oauthModel(ctx context.Context) oauthProviderModel {
	return oauthProviderModel{
		Provider: types.StringValue("siwe"), Credentials: types.MapNull(types.StringType),
		Settings: types.MapNull(types.StringType), Scopes: types.ListUnknown(types.StringType),
		Enabled: types.BoolUnknown(), AllowSignIn: types.BoolUnknown(), AllowSignUp: types.BoolUnknown(),
		ID: types.StringUnknown(), ClientID: types.StringUnknown(), HasSecret: types.BoolUnknown(),
		Configured: types.BoolUnknown(), RedirectURI: types.StringUnknown(),
		DisplayName: types.StringUnknown(), IconURL: types.StringUnknown(), Mode: types.StringUnknown(),
		UpdatedAt: types.Int64Unknown(), SettingValues: types.MapUnknown(types.StringType),
		Native: types.BoolUnknown(), AtlasCallback: types.StringUnknown(), CustomRedirectURI: types.StringUnknown(),
		JavascriptOrigins: types.ListUnknown(types.StringType),
	}
}

// A wallet sign-in (no credentials): only settings and display overrides are
// sent, and neither scope nor enabled is touched unless configured.
func TestOAuthProvider_walletSettingsAndDisplay(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodGet {
			return 200, map[string]interface{}{
				"provider": "siwe", "display_name": "Ethereum", "icon_url": "https://cdn.example/eth.svg", "mode": "redirect",
				"native": false, "category": "web3", "tier": "core", "redirect_uri": "https://x/cb", "atlas_callback": "https://x/cb",
				"custom_redirect_uri": nil, "javascript_origins": []string{}, "default_scopes": []string{},
				"configured": false, "client_id": nil, "has_secret": false, "enabled": false,
				"allow_sign_in": true, "allow_sign_up": true, "scopes": []string{}, "updated_at": nil,
				"setting_values": map[string]string{"domain": "app.example.com", "chain_ids": "1,10"},
			}
		}
		return 200, map[string]interface{}{"object": "oauth_provider", "provider": "siwe"}
	})
	r := NewOAuthProviderResource()
	configureResource(t, r, c)
	m := oauthModel(ctx)
	m.Settings, _ = types.MapValueFrom(ctx, types.StringType, map[string]string{"domain": "app.example.com", "chain_ids": "1, 10"})
	m.DisplayName = types.StringValue("Ethereum")
	m.IconURL = types.StringValue("https://cdn.example/eth.svg")
	state, resp := doCreate(t, r, planFrom(t, r, m))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)

	put := api.last(http.MethodPut, "/v1/oauth_providers/siwe")
	if put == nil {
		t.Fatal("no PUT")
	}
	for _, k := range []string{"values", "scopes", "client_id"} {
		if _, sent := put.Body[k]; sent {
			t.Errorf("a credential-less provider must not be sent %q: %v", k, put.Body)
		}
	}
	if put.Body["display_name"] != "Ethereum" || put.Body["icon_url"] != "https://cdn.example/eth.svg" {
		t.Errorf("display overrides not sent: %v", put.Body)
	}
	if s, _ := put.Body["settings"].(map[string]interface{}); s["chain_ids"] != "1, 10" {
		t.Errorf("settings not sent: %v", put.Body["settings"])
	}
	if api.count(http.MethodPost, "/v1/oauth_providers/siwe/scope") != 0 || api.count(http.MethodPost, "/v1/oauth_providers/siwe/enabled") != 0 {
		t.Error("scope/enabled must not be called unless configured for a credential-less provider")
	}
	var got oauthProviderModel
	state.Get(ctx, &got)
	var sv map[string]string
	got.SettingValues.ElementsAs(ctx, &sv, false)
	if sv["chain_ids"] != "1,10" || got.Mode.ValueString() != "redirect" {
		t.Errorf("setting_values/mode not read back: %v %s", sv, got.Mode.ValueString())
	}
	// Refresh keeps it (not configured, but it has settings).
	if _, rr := doRead(t, r, state); rr.State.Raw.IsNull() {
		t.Error("a configured wallet provider must stay in state")
	}
}

func TestOAuthProviderGone(t *testing.T) {
	ctx := context.Background()
	creds, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"client_id": "x"})
	withCreds := oauthProviderModel{Credentials: creds}
	if !oauthProviderGone(&client.OAuthProvider{Configured: false}, &withCreds) {
		t.Error("credentials removed → gone")
	}
	bare := oauthProviderModel{Credentials: types.MapNull(types.StringType), Settings: types.MapNull(types.StringType),
		DisplayName: types.StringNull(), Mode: types.StringNull()}
	if !oauthProviderGone(&client.OAuthProvider{}, &bare) {
		t.Error("nothing left → gone")
	}
	if oauthProviderGone(&client.OAuthProvider{SettingValues: map[string]string{"domain": "a.b"}}, &bare) {
		t.Error("stored wallet settings → present")
	}
}
