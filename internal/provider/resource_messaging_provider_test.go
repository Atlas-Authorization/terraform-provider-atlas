package provider

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// messagingPlan builds a plan with one transport block set.
func messagingPlan(t *testing.T, r resource.Resource, from string, transport string, values map[string]attr.Value) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	s := schemaOf(t, r)
	p := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	mr := r.(*messagingProviderResource)
	for _, tr := range mr.transports {
		obj := types.ObjectNull(tr.objectType().AttrTypes)
		if tr.name == transport {
			full := map[string]attr.Value{}
			for _, f := range tr.fields {
				if v, ok := values[f.key]; ok {
					full[f.key] = v
					continue
				}
				switch f.kind {
				case kindInt:
					full[f.key] = types.Int64Null()
				case kindBool:
					full[f.key] = types.BoolNull()
				default:
					full[f.key] = types.StringNull()
				}
			}
			var d = p.SetAttribute(ctx, path.Root(tr.name), types.ObjectValueMust(tr.objectType().AttrTypes, full))
			if d.HasError() {
				t.Fatalf("set %s: %v", tr.name, d)
			}
			continue
		}
		p.SetAttribute(ctx, path.Root(tr.name), obj)
	}
	fromV := types.StringNull()
	if from != "" {
		fromV = types.StringValue(from)
	}
	p.SetAttribute(ctx, path.Root("from_address"), fromV)
	for _, name := range []string{"id", "transport", "status"} {
		p.SetAttribute(ctx, path.Root(name), types.StringUnknown())
	}
	p.SetAttribute(ctx, path.Root("secrets_set"), types.SetUnknown(types.StringType))
	p.SetAttribute(ctx, path.Root("updated_at"), types.Int64Unknown())
	return p
}

func TestEmailProvider_smtpTypedFieldsAndWriteOnlySecret(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		view := map[string]interface{}{
			"object": "messaging_provider", "channel": "email", "transport": "smtp",
			"from_address": "Acme <no-reply@acme.example>", "status": "active",
			"config":      map[string]string{"host": "smtp.acme.example", "port": "465", "secure": "true"},
			"secrets_set": map[string]bool{"password": true}, "updated_at": 5,
		}
		if call.Method == http.MethodGet {
			return 200, map[string]interface{}{"object": "messaging_providers", "email": view, "sms": nil}
		}
		return 200, view
	})
	r := NewEmailProviderResource()
	configureResource(t, r, c)
	if typeNameOf(t, r) != "atlas_email_provider" {
		t.Fatalf("type name %s", typeNameOf(t, r))
	}
	plan := messagingPlan(t, r, "Acme <no-reply@acme.example>", "smtp", map[string]attr.Value{
		"host": types.StringValue("smtp.acme.example"), "port": types.Int64Value(465),
		"secure": types.BoolValue(true), "password": types.StringValue("hunter2"),
	})
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	put := api.last(http.MethodPut, "/v1/messaging_providers/email")
	if put == nil || put.Body["transport"] != "smtp" {
		t.Fatalf("PUT = %+v", put)
	}
	cfg, _ := put.Body["config"].(map[string]interface{})
	if cfg["port"] != "465" || cfg["secure"] != "true" || cfg["password"] != "hunter2" {
		t.Errorf("config must be sent as strings incl. the secret: %v", cfg)
	}
	if _, sent := cfg["username"]; sent {
		t.Error("unset fields must be omitted")
	}

	// A refresh keeps the write-only secret from state and reads typed fields back.
	refreshed, rr := doRead(t, r, state)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	var smtp types.Object
	refreshed.GetAttribute(ctx, path.Root("smtp"), &smtp)
	attrs := smtp.Attributes()
	if attrs["password"].(types.String).ValueString() != "hunter2" {
		t.Error("secret must survive a read (never returned by the API)")
	}
	if attrs["port"].(types.Int64).ValueInt64() != 465 || !attrs["secure"].(types.Bool).ValueBool() {
		t.Errorf("typed fields not read back: %v", attrs)
	}
	var secrets types.Set
	refreshed.GetAttribute(ctx, path.Root("secrets_set"), &secrets)
	if len(secrets.Elements()) != 1 {
		t.Errorf("secrets_set = %v", secrets)
	}
	var resend types.Object
	refreshed.GetAttribute(ctx, path.Root("resend"), &resend)
	if !resend.IsNull() {
		t.Error("other transport blocks must be null")
	}

	if d := doDelete(t, r, state); d.Diagnostics.HasError() || api.last(http.MethodDelete, "/v1/messaging_providers/email") == nil {
		t.Error("delete must DELETE /v1/messaging_providers/email")
	}
}

func TestSMSProvider_twilioAndAbsent(t *testing.T) {
	present := true
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		view := map[string]interface{}{
			"channel": "sms", "transport": "twilio", "from_address": nil, "status": "active",
			"config": map[string]string{"account_sid": "AC1", "from": "+15550001"}, "secrets_set": map[string]bool{"auth_token": true}, "updated_at": 1,
		}
		if call.Method == http.MethodGet {
			if !present {
				return 200, map[string]interface{}{"email": nil, "sms": nil}
			}
			return 200, map[string]interface{}{"email": nil, "sms": view}
		}
		return 200, view
	})
	r := NewSMSProviderResource()
	configureResource(t, r, c)
	plan := messagingPlan(t, r, "", "twilio", map[string]attr.Value{
		"account_sid": types.StringValue("AC1"), "auth_token": types.StringValue("tok"), "from": types.StringValue("+15550001"),
	})
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	put := api.last(http.MethodPut, "/v1/messaging_providers/sms")
	if put == nil || put.Body["transport"] != "twilio" {
		t.Fatalf("PUT = %+v", put)
	}
	if _, sent := put.Body["from_address"]; sent {
		t.Error("unset from_address must be omitted")
	}
	present = false
	_, rr := doRead(t, r, state)
	if !rr.State.Raw.IsNull() {
		t.Error("an unconfigured channel must drop from state")
	}
}

func TestMessagingProvider_exactlyOneTransport(t *testing.T) {
	ctx := context.Background()
	r := NewEmailProviderResource()
	vr := r.(resource.ResourceWithValidateConfig)
	none := messagingPlan(t, r, "", "", nil)
	var resp resource.ValidateConfigResponse
	vr.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: none.Schema, Raw: none.Raw}}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("no transport block must be rejected")
	}
	one := messagingPlan(t, r, "", "resend", map[string]attr.Value{"api_key": types.StringValue("re_x")})
	resp = resource.ValidateConfigResponse{}
	vr.ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: one.Schema, Raw: one.Raw}}, &resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("one transport block must pass: %v", resp.Diagnostics)
	}
	// Secrets are sensitive in the schema.
	s := schemaOf(t, r)
	resend, ok := s.Attributes["resend"].(rschema.SingleNestedAttribute)
	if !ok {
		t.Fatal("missing resend block")
	}
	if !resend.Attributes["api_key"].IsSensitive() {
		t.Error("secret settings must be sensitive")
	}
	if smtp := s.Attributes["smtp"].(rschema.SingleNestedAttribute); smtp.Attributes["host"].IsSensitive() {
		t.Error("non-secret settings must not be sensitive")
	}
}

func TestEmailProvider_forceDestroyAndInUse(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		if call.Method == http.MethodDelete {
			if call.Query != "force=true" {
				return 409, map[string]interface{}{"errors": []map[string]interface{}{{
					"code": "EMAIL_PROVIDER_IN_USE", "message": "This email provider is the only way this instance can send email.",
					"meta": map[string]interface{}{"features": []string{"Email code sign-in and step-up", "Password reset"}},
				}}}
			}
			return 200, map[string]interface{}{"object": "messaging_provider", "channel": "email", "deleted": true}
		}
		return 200, map[string]interface{}{
			"channel": "email", "transport": "resend", "from_address": nil, "status": "active",
			"config": map[string]string{}, "secrets_set": map[string]bool{"api_key": true}, "updated_at": 1,
		}
	})
	r := NewEmailProviderResource()
	configureResource(t, r, c)
	plan := messagingPlan(t, r, "", "resend", map[string]attr.Value{"api_key": types.StringValue("re_x")})
	plan.SetAttribute(ctx, path.Root("force_destroy"), false)
	state, resp := doCreate(t, r, plan)
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)

	// Refused: the error names the features and both ways out.
	d := doDelete(t, r, state)
	if !d.Diagnostics.HasError() {
		t.Fatal("a 409 EMAIL_PROVIDER_IN_USE must fail the destroy")
	}
	detail := d.Diagnostics[0].Detail()
	for _, want := range []string{"Email code sign-in and step-up, Password reset", "force_destroy = true", "turn those features off"} {
		if !strings.Contains(detail, want) {
			t.Errorf("error detail missing %q:\n%s", want, detail)
		}
	}
	if got := api.last(http.MethodDelete, "/v1/messaging_providers/email"); got == nil || got.Query != "" {
		t.Errorf("without force_destroy the DELETE must carry no query, got %+v", got)
	}

	// force_destroy only changes state: no request to Atlas.
	puts := api.count(http.MethodPut, "/v1/messaging_providers/email")
	next := messagingPlan(t, r, "", "resend", map[string]attr.Value{"api_key": types.StringValue("re_x")})
	next.SetAttribute(ctx, path.Root("force_destroy"), true)
	updated, ur := doUpdate(t, r, next, state)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	if api.count(http.MethodPut, "/v1/messaging_providers/email") != puts {
		t.Error("changing only force_destroy must not PUT the provider")
	}
	var force types.Bool
	updated.GetAttribute(ctx, path.Root("force_destroy"), &force)
	var status types.String
	updated.GetAttribute(ctx, path.Root("status"), &status)
	if !force.ValueBool() || status.ValueString() != "active" {
		t.Errorf("state after a force_destroy-only update: force=%v status=%v", force, status)
	}

	if d := doDelete(t, r, updated); d.Diagnostics.HasError() {
		t.Fatalf("force_destroy delete: %v", d.Diagnostics)
	}
	if got := api.last(http.MethodDelete, "/v1/messaging_providers/email"); got == nil || got.Query != "force=true" {
		t.Errorf("force_destroy must send ?force=true, got %+v", got)
	}

	// A real change alongside force_destroy still saves the provider.
	changed := messagingPlan(t, r, "Acme <no-reply@acme.example>", "resend", map[string]attr.Value{"api_key": types.StringValue("re_x")})
	changed.SetAttribute(ctx, path.Root("force_destroy"), false)
	if _, cr := doUpdate(t, r, changed, updated); cr.Diagnostics.HasError() || api.count(http.MethodPut, "/v1/messaging_providers/email") != puts+1 {
		t.Errorf("a from_address change must PUT: %v", cr.Diagnostics)
	}
}

func TestSMSProvider_noForceDestroy(t *testing.T) {
	if _, ok := schemaOf(t, NewSMSProviderResource()).Attributes["force_destroy"]; ok {
		t.Error("force_destroy is email-only: removing the SMS provider is never refused")
	}
	if _, ok := schemaOf(t, NewEmailProviderResource()).Attributes["force_destroy"]; !ok {
		t.Error("atlas_email_provider must have force_destroy")
	}
}
