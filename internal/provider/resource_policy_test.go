package provider

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fakePolicies is an in-memory policy store with the API's rules: keys are
// unique (409), GET works by id or key, DELETE retires a published policy.
type fakePolicies struct {
	mu        sync.Mutex
	next      int
	byID      map[string]map[string]interface{}
	published map[string]bool
}

func newFakePolicies() *fakePolicies {
	return &fakePolicies{byID: map[string]map[string]interface{}{}, published: map[string]bool{}}
}

func (f *fakePolicies) find(idOrKey string) map[string]interface{} {
	if p, ok := f.byID[idOrKey]; ok {
		return p
	}
	for _, p := range f.byID {
		if p["key"] == idOrKey {
			return p
		}
	}
	return nil
}

func (f *fakePolicies) handle(call apiCall) (int, interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	notFound := map[string]interface{}{"errors": []map[string]interface{}{{"code": "NOT_FOUND", "message": "not found"}}}
	switch {
	case call.Method == http.MethodPost && call.Path == "/v1/policies":
		if f.find(call.Body["key"].(string)) != nil {
			return 409, map[string]interface{}{"errors": []map[string]interface{}{{"code": "CONFLICT", "message": "key taken", "param": "key"}}}
		}
		f.next++
		id := "pol_" + string(rune('a'+f.next))
		p := map[string]interface{}{"id": id, "key": call.Body["key"], "source": "atlas", "position": 0, "status": "active"}
		for _, k := range []string{"title", "display", "required", "applies_to", "reaccept", "grace_days"} {
			p[k] = call.Body[k]
		}
		f.byID[id] = p
		return 201, p
	case strings.HasPrefix(call.Path, "/v1/policies/"):
		p := f.find(strings.TrimPrefix(call.Path, "/v1/policies/"))
		if p == nil {
			return 404, notFound
		}
		switch call.Method {
		case http.MethodGet:
			return 200, p
		case http.MethodPatch:
			for k, v := range call.Body {
				p[k] = v
			}
			return 200, p
		case http.MethodDelete:
			if f.published[p["id"].(string)] {
				p["status"] = "retired"
				return 200, map[string]interface{}{"object": "policy", "id": p["id"], "retired": true}
			}
			delete(f.byID, p["id"].(string))
			return 200, map[string]interface{}{"object": "policy", "id": p["id"], "deleted": true}
		}
	}
	return 500, nil
}

func policyPlan(key, title, display string, required bool, appliesTo, reaccept string, grace types.Int64) policyModel {
	return policyModel{
		ID: types.StringUnknown(), Key: types.StringValue(key), Title: types.StringValue(title),
		Display: types.StringValue(display), Required: types.BoolValue(required), AppliesTo: types.StringValue(appliesTo),
		Reaccept: types.StringValue(reaccept), GraceDays: grace, Status: types.StringUnknown(),
	}
}

func TestPolicy_lifecycle(t *testing.T) {
	ctx := context.Background()
	fake := newFakePolicies()
	c, api := newFakeAPI(t, fake.handle)
	r := NewPolicyResource()
	configureResource(t, r, c)
	if typeNameOf(t, r) != "atlas_policy" {
		t.Fatalf("type name %s", typeNameOf(t, r))
	}

	state, resp := doCreate(t, r, planFrom(t, r, policyPlan("tos", "Terms of service", "scroll", true, "sign_in_and_sign_up", "grace", types.Int64Value(14))))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	post := api.last(http.MethodPost, "/v1/policies")
	if post.Body["key"] != "tos" || post.Body["display"] != "scroll" || post.Body["reaccept"] != "grace" || post.Body["grace_days"] != float64(14) {
		t.Fatalf("POST body = %v", post.Body)
	}
	var m policyModel
	state.Get(ctx, &m)
	if m.ID.ValueString() == "" || m.Status.ValueString() != "active" || m.GraceDays.ValueInt64() != 14 {
		t.Fatalf("state = %+v", m)
	}

	// Refresh is stable.
	refreshed, rr := doRead(t, r, state)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	if !refreshed.Raw.Equal(state.Raw) {
		t.Errorf("refresh drifted:\n%v\n%v", refreshed.Raw, state.Raw)
	}

	// Updating to a policy with no grace window clears grace_days with an explicit null.
	plan2 := policyPlan("tos", "Terms of service", "scroll", true, "sign_in_and_sign_up", "next_sign_in", types.Int64Null())
	plan2.ID = m.ID
	plan2.Status = types.StringValue("active")
	state2, ur := doUpdate(t, r, planFrom(t, r, plan2), state)
	mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
	patch := api.last(http.MethodPatch, "/v1/policies/"+m.ID.ValueString())
	if v, sent := patch.Body["grace_days"]; !sent || v != nil || patch.Body["reaccept"] != "next_sign_in" || patch.Body["status"] != "active" {
		t.Errorf("PATCH body = %v", patch.Body)
	}
	if _, sent := patch.Body["key"]; sent {
		t.Error("the key must not be sent on update")
	}
	state2.Get(ctx, &m)
	if !m.GraceDays.IsNull() {
		t.Errorf("grace_days = %v", m.GraceDays)
	}

	// Destroy deletes a never-published policy; a refresh then drops it.
	if d := doDelete(t, r, state2); d.Diagnostics.HasError() || api.count(http.MethodDelete, "/v1/policies/"+m.ID.ValueString()) != 1 {
		t.Errorf("delete: %v", d.Diagnostics)
	}
	_, rr = doRead(t, r, state2)
	if !rr.State.Raw.IsNull() {
		t.Error("a deleted policy must drop from state")
	}
}

func TestPolicy_retiredIsGoneAndIsRevivedOnCreate(t *testing.T) {
	ctx := context.Background()
	fake := newFakePolicies()
	c, api := newFakeAPI(t, fake.handle)
	r := NewPolicyResource()
	configureResource(t, r, c)

	state, resp := doCreate(t, r, planFrom(t, r, policyPlan("privacy", "Privacy", "checkbox", true, "sign_up", "none", types.Int64Null())))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	var m policyModel
	state.Get(ctx, &m)
	fake.published[m.ID.ValueString()] = true // it has a published version now

	// Destroy retires it; a refresh treats it as gone.
	if d := doDelete(t, r, state); d.Diagnostics.HasError() {
		t.Fatal(d.Diagnostics)
	}
	if fake.byID[m.ID.ValueString()]["status"] != "retired" {
		t.Fatal("the fake should have retired the policy")
	}
	_, rr := doRead(t, r, state)
	if !rr.State.Raw.IsNull() {
		t.Error("a retired policy must drop from state")
	}

	// Creating the same key again revives the retired policy with the new configuration.
	state2, cr := doCreate(t, r, planFrom(t, r, policyPlan("privacy", "Privacy notice", "link", false, "sign_up", "none", types.Int64Null())))
	mustOK(t, "revive", cr.Diagnostics.HasError(), cr.Diagnostics)
	patch := api.last(http.MethodPatch, "/v1/policies/"+m.ID.ValueString())
	if patch == nil || patch.Body["status"] != "active" || patch.Body["title"] != "Privacy notice" {
		t.Fatalf("revive PATCH = %+v", patch)
	}
	var revived policyModel
	state2.Get(ctx, &revived)
	if revived.ID.ValueString() != m.ID.ValueString() || revived.Status.ValueString() != "active" || revived.Display.ValueString() != "link" {
		t.Errorf("revived = %+v", revived)
	}

	// An existing ACTIVE policy with the key is an error pointing at import.
	_, cr = doCreate(t, r, planFrom(t, r, policyPlan("privacy", "Again", "checkbox", true, "sign_up", "none", types.Int64Null())))
	if !cr.Diagnostics.HasError() || !strings.Contains(cr.Diagnostics[0].Detail(), "terraform import") {
		t.Errorf("a duplicate active key must point at import: %v", cr.Diagnostics)
	}
}

func TestPolicy_importByKey(t *testing.T) {
	ctx := context.Background()
	fake := newFakePolicies()
	c, _ := newFakeAPI(t, fake.handle)
	r := NewPolicyResource()
	configureResource(t, r, c)
	state, resp := doCreate(t, r, planFrom(t, r, policyPlan("tos", "Terms", "checkbox", true, "sign_up", "none", types.Int64Null())))
	mustOK(t, "create", resp.Diagnostics.HasError(), resp.Diagnostics)
	var created policyModel
	state.Get(ctx, &created)

	var ir resource.ImportStateResponse
	ir.State = nullState(t, r)
	r.(resource.ResourceWithImportState).ImportState(ctx, resource.ImportStateRequest{ID: "tos"}, &ir)
	mustOK(t, "import", ir.Diagnostics.HasError(), ir.Diagnostics)
	st, rr := doRead(t, r, ir.State)
	mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
	var id types.String
	st.GetAttribute(ctx, path.Root("id"), &id)
	if id.ValueString() != created.ID.ValueString() {
		t.Errorf("an import by key must resolve to the real id, got %q want %q", id.ValueString(), created.ID.ValueString())
	}
}

func TestPolicy_validation(t *testing.T) {
	ctx := context.Background()
	r := NewPolicyResource()
	validate := func(m policyModel) []string {
		p := planFrom(t, r, m)
		var resp resource.ValidateConfigResponse
		r.(resource.ResourceWithValidateConfig).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: p.Schema, Raw: p.Raw}}, &resp)
		var out []string
		for _, d := range resp.Diagnostics {
			out = append(out, d.Summary())
		}
		return out
	}
	if got := validate(policyPlan("tos", "Terms", "scroll", true, "sign_in_and_sign_up", "grace", types.Int64Value(30))); len(got) != 0 {
		t.Errorf("a valid config must pass: %v", got)
	}
	cases := map[string]policyModel{
		"Invalid key":                           policyPlan("Terms of Service", "T", "checkbox", true, "sign_up", "none", types.Int64Null()),
		"Invalid display":                       policyPlan("tos", "T", "popup", true, "sign_up", "none", types.Int64Null()),
		"Invalid applies_to":                    policyPlan("tos", "T", "checkbox", true, "always", "none", types.Int64Null()),
		"Invalid reaccept":                      policyPlan("tos", "T", "checkbox", true, "sign_up", "sometimes", types.Int64Null()),
		"grace_days is required":                policyPlan("tos", "T", "checkbox", true, "sign_in_and_sign_up", "grace", types.Int64Null()),
		"grace_days needs reaccept = \"grace\"": policyPlan("tos", "T", "checkbox", true, "sign_in_and_sign_up", "none", types.Int64Value(7)),
		"Invalid grace_days":                    policyPlan("tos", "T", "checkbox", true, "sign_in_and_sign_up", "grace", types.Int64Value(400)),
		"Invalid title":                         policyPlan("tos", "", "checkbox", true, "sign_up", "none", types.Int64Null()),
	}
	for want, m := range cases {
		got := validate(m)
		if len(got) != 1 || got[0] != want {
			t.Errorf("%s: got %v", want, got)
		}
	}
}
