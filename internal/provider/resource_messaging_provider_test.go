package provider

import (
	"context"
	"net/http"
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
