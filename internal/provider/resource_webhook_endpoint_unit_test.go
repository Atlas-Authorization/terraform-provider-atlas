package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func webhookFake(t *testing.T, active bool) (*webhookEndpointResource, *fakeAPI) {
	endpoint := func(secret bool) map[string]interface{} {
		e := map[string]interface{}{"object": "webhook_endpoint", "id": "we_1", "url": "https://hooks.example/a",
			"enabled_events": []string{"user.created"}, "active": active, "failing_since": nil, "disabled_at": nil, "created_at": 1}
		if secret {
			e["secret"] = "whsec_create"
		}
		return e
	}
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		switch {
		case call.Method == http.MethodPost && call.Path == "/v1/webhook_endpoints":
			return 201, endpoint(true)
		case call.Path == "/v1/webhook_endpoints/we_1/reveal_secret":
			return 200, map[string]string{"object": "webhook_secret", "secret": "whsec_revealed"}
		case call.Path == "/v1/webhook_endpoints/we_1/rotate_secret":
			return 200, map[string]string{"object": "webhook_secret", "secret": "whsec_rotated"}
		case call.Method == http.MethodGet:
			return 200, endpoint(false)
		}
		return 200, endpoint(false)
	})
	r := NewWebhookEndpointResource().(*webhookEndpointResource)
	configureResource(t, r, c)
	return r, api
}

func webhookModel(ctx context.Context) webhookEndpointModel {
	events, _ := types.ListValueFrom(ctx, types.StringType, []string{"user.created"})
	m := webhookEndpointModel{
		ID: types.StringUnknown(), URL: types.StringValue("https://hooks.example/a"), EnabledEvents: events,
		Active: types.BoolUnknown(), Secret: types.StringUnknown(), RotateTrigger: types.StringNull(),
		FailingSince: types.Int64Unknown(), DisabledAt: types.Int64Unknown(), CreatedAt: types.Int64Unknown(),
	}
	withWebhookExtras(&m)
	return m
}

// nullWebhookExtras makes the 2026-10 attributes null, for a model used as state.
func nullWebhookExtras(m *webhookEndpointModel) {
	m.EventNaming = types.StringNull()
	m.External = types.BoolNull()
	m.SignatureVersions = types.SetNull(types.StringType)
	m.AlertContacts = types.SetNull(types.StringType)
	m.Destination = types.MapNull(types.StringType)
	m.DestinationCredsSet = types.BoolNull()
	m.Health = types.StringNull()
	m.LastSuccessAt = types.Int64Null()
	m.ConsecutiveFailures = types.Int64Null()
	m.LastFailingSince = types.Int64Null()
}

// withWebhookExtras fills the 2026-10 attributes with typed unknowns/nulls so a
// model built for a plan covers the whole schema.
func withWebhookExtras(m *webhookEndpointModel) {
	m.Description = types.StringNull()
	m.EventNaming = types.StringUnknown()
	m.External = types.BoolUnknown()
	m.SignatureVersions = types.SetUnknown(types.StringType)
	m.AlertContacts = types.SetUnknown(types.StringType)
	m.Destination = types.MapUnknown(types.StringType)
	m.DestinationCreds = types.MapNull(types.StringType)
	m.DestinationCredsSet = types.BoolUnknown()
	m.ExternalRef = types.StringNull()
	m.Metadata = types.MapValueMust(types.StringType, map[string]attr.Value{})
	m.Health = types.StringUnknown()
	m.LastSuccessAt = types.Int64Unknown()
	m.ConsecutiveFailures = types.Int64Unknown()
	m.LastFailingSince = types.Int64Unknown()
}

func TestWebhookEndpoint_rotateTriggerAndImportReveal(t *testing.T) {
	ctx := context.Background()
	r, api := webhookFake(t, true)
	state, resp := doCreate(t, r, planFrom(t, r, webhookModel(ctx)))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	var m webhookEndpointModel
	state.Get(ctx, &m)
	if m.Secret.ValueString() != "whsec_create" {
		t.Fatalf("create secret = %q", m.Secret.ValueString())
	}

	// Changing the trigger rotates in place.
	next := m
	next.RotateTrigger = types.StringValue("v2")
	updated, uresp := doUpdate(t, r, planFrom(t, r, next), state)
	mustOK(t, "update", uresp.Diagnostics.HasError(), uresp.Diagnostics)
	var um webhookEndpointModel
	updated.Get(ctx, &um)
	if um.Secret.ValueString() != "whsec_rotated" || api.count(http.MethodPost, "/v1/webhook_endpoints/we_1/rotate_secret") != 1 {
		t.Errorf("rotation did not run / store the new secret: %q", um.Secret.ValueString())
	}
	if api.last(http.MethodPatch, "/v1/webhook_endpoints/we_1") == nil {
		t.Error("an update must PATCH in place")
	}

	// Import: state has only the id → read reveals the secret.
	imported := webhookModel(ctx)
	imported.ID = types.StringValue("we_1")
	imported.Secret = types.StringNull()
	imported.Active = types.BoolNull()
	imported.FailingSince, imported.DisabledAt, imported.CreatedAt = types.Int64Null(), types.Int64Null(), types.Int64Null()
	nullWebhookExtras(&imported)
	read, rr := doRead(t, r, stateFrom(t, r, imported))
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	var im webhookEndpointModel
	read.Get(ctx, &im)
	if im.Secret.ValueString() != "whsec_revealed" {
		t.Errorf("import must reveal the secret, got %q", im.Secret.ValueString())
	}
	// A normal refresh (secret already known) never calls reveal again.
	before := api.count(http.MethodPost, "/v1/webhook_endpoints/we_1/reveal_secret")
	doRead(t, r, read)
	if api.count(http.MethodPost, "/v1/webhook_endpoints/we_1/reveal_secret") != before {
		t.Error("refresh with a known secret must not reveal again")
	}
}

func TestWebhookEndpoint_inactiveStaysInStateAndPauseOnCreate(t *testing.T) {
	ctx := context.Background()
	r, api := webhookFake(t, false)
	plan := webhookModel(ctx)
	plan.Active = types.BoolValue(false)
	state, resp := doCreate(t, r, planFrom(t, r, plan))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	patch := api.last(http.MethodPatch, "/v1/webhook_endpoints/we_1")
	if patch == nil || patch.Body["active"] != false {
		t.Fatalf("active=false must pause the new endpoint, got %+v", patch)
	}
	var m webhookEndpointModel
	state.Get(ctx, &m)
	if m.Secret.ValueString() != "whsec_create" {
		t.Error("pausing must keep the create-time secret")
	}
	// An auto-disabled (inactive) endpoint still exists.
	_, rr := doRead(t, r, state)
	if rr.State.Raw.IsNull() {
		t.Error("an inactive endpoint must stay in state")
	}
}

func TestWebhookEndpoint_modifyPlanMarksSecretUnknown(t *testing.T) {
	ctx := context.Background()
	r := NewWebhookEndpointResource().(resource.ResourceWithModifyPlan)
	prior := webhookModel(ctx)
	prior.ID, prior.Secret, prior.Active = types.StringValue("we_1"), types.StringValue("whsec_a"), types.BoolValue(true)
	prior.FailingSince, prior.DisabledAt, prior.CreatedAt = types.Int64Null(), types.Int64Null(), types.Int64Value(1)
	nullWebhookExtras(&prior)
	planned := prior
	planned.RotateTrigger = types.StringValue("x")
	res := NewWebhookEndpointResource()
	req := resource.ModifyPlanRequest{State: stateFrom(t, res, prior), Plan: planFrom(t, res, planned)}
	resp := resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(ctx, req, &resp)
	var out webhookEndpointModel
	resp.Plan.Get(ctx, &out)
	if !out.Secret.IsUnknown() {
		t.Error("a trigger change must plan the secret as unknown")
	}
}
