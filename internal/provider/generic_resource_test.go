package provider

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var genNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Every generated spec must be a valid, uniquely named resource whose schema the
// framework accepts, and must not collide with a hand-written resource or data source.
func TestGenericSpecsAreValid(t *testing.T) {
	ctx := context.Background()
	seen := map[string]string{}
	for _, f := range handWrittenResources() {
		seen[typeNameOf(t, f())] = "hand-written resource"
	}
	for _, spec := range genResourceSpecs() {
		spec := spec
		if spec.Kind == "dataonly" {
			continue
		}
		name := "atlas_" + spec.Type
		if prev, dup := seen[name]; dup {
			t.Errorf("%s collides with a %s", name, prev)
		}
		seen[name] = "generated resource"
		r := newGenResource(&spec)()
		sch := schemaOf(t, r)
		if d := sch.ValidateImplementation(ctx); d.HasError() {
			t.Errorf("%s: schema invalid: %v", name, d)
		}
		attrs := map[string]bool{}
		for _, a := range spec.Attrs {
			if !genNameRe.MatchString(a.Name) || a.Name == "id" || attrs[a.Name] {
				t.Errorf("%s: bad or duplicate attribute %q", name, a.Name)
			}
			attrs[a.Name] = true
			if a.Req && a.Comp {
				t.Errorf("%s.%s is both required and computed", name, a.Name)
			}
		}
		if spec.Kind == "coll" && !strings.HasSuffix(spec.Item, "/{id}") {
			t.Errorf("%s: coll item path %q must end in {id}", name, spec.Item)
		}
		if !spec.GetItem && spec.List == "" && !spec.NoRead {
			t.Errorf("%s: no way to read the object", name)
		}
		for _, p := range spec.params() {
			if !strings.Contains(spec.Item, "{"+p.Name+"}") {
				t.Errorf("%s: path key %s is not in %s", name, p.Name, spec.Item)
			}
		}
	}
	dsSeen := map[string]bool{}
	for _, f := range New("test")().DataSources(ctx) {
		n := dataSourceTypeName(f())
		if dsSeen[n] {
			t.Errorf("duplicate data source %s", n)
		}
		dsSeen[n] = true
	}
}

func TestGenericSpecsAreRegistered(t *testing.T) {
	ctx := context.Background()
	have := map[string]bool{}
	for _, f := range New("test")().Resources(ctx) {
		have[typeNameOf(t, f())] = true
	}
	for _, spec := range genResourceSpecs() {
		if spec.Kind != "dataonly" && !have["atlas_"+spec.Type] {
			t.Errorf("atlas_%s is not registered", spec.Type)
		}
	}
	// the handoff resources of the re-audit tracks
	for _, want := range []string{
		"workflow", "chat_integration", "iga_bundle", "iga_eligibility", "administrative_unit", "organization_application",
		"oauth_client_organization_access", "organization_connection", "mcp_gateway", "mcp_gateway_upstream", "mcp_registry_entry",
		"connection_action", "connection_trigger", "agent_lifecycle", "telemetry_export", "access_policy", "device_assurance_policy",
		"device_integration", "authenticator_policy", "email_risk_settings", "email_suppression", "fga_assertions", "fga_materialization",
		"abac_token_mapping", "billing_coupon", "audit_retention", "data_key_provider", "scim_provisioning_target", "scim_resource_type",
		"app_catalog_preset", "cloud_access_target", "wsfed_connection",
	} {
		if !have["atlas_"+want] {
			t.Errorf("atlas_%s is missing", want)
		}
	}
}

func TestGenericDataSourcesHaveSchemas(t *testing.T) {
	ctx := context.Background()
	n := 0
	for _, f := range genDataSourceFactories() {
		d := f()
		var resp datasource.SchemaResponse
		d.Schema(ctx, datasource.SchemaRequest{}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("%s: %v", dataSourceTypeName(d), resp.Diagnostics)
		}
		if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: %v", dataSourceTypeName(d), diags)
		}
		n++
	}
	if n == 0 {
		t.Fatal("no generated data sources")
	}
}

func TestJSONCovers(t *testing.T) {
	cases := []struct {
		cfg, server any
		want        bool
	}{
		{map[string]any{"a": 1.0}, map[string]any{"a": 1.0, "b": "default"}, true},
		{map[string]any{"a": 1.0}, map[string]any{"a": 2.0}, false},
		{map[string]any{"a": nil}, map[string]any{}, true},
		{[]any{1.0, 2.0}, []any{1.0, 2.0}, true},
		{[]any{1.0}, []any{1.0, 2.0}, false},
		{"x", "x", true},
		{"x", "y", false},
	}
	for i, c := range cases {
		if got := jsonCovers(c.cfg, c.server); got != c.want {
			t.Errorf("case %d: jsonCovers = %v, want %v", i, got, c.want)
		}
	}
}

func TestFillPathEscapes(t *testing.T) {
	got := fillPath("/v1/x/{a}/y/{id}", map[string]string{"a": "p/q", "id": "u@e.com"})
	if got != "/v1/x/p%2Fq/y/u@e.com" {
		t.Fatalf("got %s", got)
	}
}

// sampleFor gives a plan value of the attribute's kind.
func sampleFor(a genAttr) attr.Value {
	switch a.Kind {
	case "bool":
		return types.BoolValue(true)
	case "int":
		return types.Int64Value(5)
	case "float":
		return types.Float64Value(1.5)
	case "strlist":
		l, _ := types.ListValueFrom(context.Background(), types.StringType, []string{"a"})
		return l
	case "json":
		return types.StringValue(`{"a":1}`)
	}
	return types.StringValue("v-" + a.Name)
}

func unknownFor(a genAttr) attr.Value {
	switch a.Kind {
	case "bool":
		return types.BoolUnknown()
	case "int":
		return types.Int64Unknown()
	case "float":
		return types.Float64Unknown()
	case "strlist":
		return types.ListUnknown(types.StringType)
	}
	return types.StringUnknown()
}

// buildPlan sets every attribute the way the framework would for a new object.
func buildPlan(t *testing.T, r resource.Resource, spec *genSpec, id string) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	s := schemaOf(t, r)
	p := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	set := func(name string, v attr.Value) {
		if d := p.SetAttribute(ctx, path.Root(name), v); d.HasError() {
			t.Fatalf("%s: set %s: %v", spec.Type, name, d)
		}
	}
	if id == "" {
		set("id", types.StringUnknown())
	} else {
		set("id", types.StringValue(id))
	}
	for _, a := range spec.Attrs {
		switch {
		case a.Param:
			set(a.Name, types.StringValue("p-"+a.Name))
		case a.Req:
			set(a.Name, sampleFor(a))
		case a.Comp:
			set(a.Name, unknownFor(a))
		case a.WriteOnly:
			set(a.Name, sampleFor(a))
		default: // optional + computed, left to the server
			set(a.Name, unknownFor(a))
		}
	}
	return p
}

// genIDFor is the id the fake API hands out: an id that is also a configured
// field (an email, a key) is whatever the plan carries.
func genIDFor(spec *genSpec) string {
	if f := spec.idField(); f != "id" {
		for _, a := range spec.Attrs {
			if a.API == f && !a.Comp && a.Kind == "string" {
				return "v-" + a.Name
			}
		}
	}
	return "obj_1"
}

func genObjectFor(spec *genSpec) map[string]interface{} {
	obj := map[string]interface{}{"id": "obj_1"}
	if f := spec.idField(); f != "id" {
		obj[f] = genIDFor(spec)
	}
	// a keyed object found by scanning a list carries its key in the list item
	if ps := spec.params(); spec.Kind == "keyed" && spec.Match != "" && len(ps) > 0 {
		obj[spec.Match] = "p-" + ps[len(ps)-1].Name
	}
	return obj
}

// TestGenericCRUD drives create, read, update, delete and import for every
// generated resource against a fake API that stores what it is sent.
func TestGenericCRUD(t *testing.T) {
	for _, spec := range genResourceSpecs() {
		spec := spec
		if spec.Kind == "dataonly" {
			continue
		}
		t.Run(spec.Type, func(t *testing.T) {
			stored := genObjectFor(&spec)
			vals := map[string]string{}
			for _, p := range spec.params() {
				vals[p.Name] = "p-" + p.Name
			}
			wantID := spec.Type
			itemVals := map[string]string{}
			for k, v := range vals {
				itemVals[k] = v
			}
			switch spec.Kind {
			case "coll":
				wantID = genIDFor(&spec)
				itemVals["id"] = wantID
			case "keyed":
				parts := []string{}
				for _, p := range spec.params() {
					parts = append(parts, vals[p.Name])
				}
				wantID = strings.Join(parts, "/")
			}
			itemPath := fillPath(spec.Item, itemVals)

			c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
				switch call.Method {
				case http.MethodPost, http.MethodPut, http.MethodPatch:
					if call.Path == itemPath+spec.Revoke && spec.Revoke != "" {
						return 200, map[string]interface{}{}
					}
					for k, v := range call.Body {
						stored[k] = v
					}
					if spec.Pick == "data0" {
						return 201, map[string]interface{}{"data": []interface{}{stored}}
					}
					return 200, stored
				case http.MethodGet:
					if call.Path == itemPath {
						return 200, stored
					}
					return 200, map[string]interface{}{"data": []interface{}{stored}, "has_more": false}
				}
				return 200, map[string]interface{}{"deleted": true}
			})
			r := newGenResource(&spec)()
			configureResource(t, r, c)

			// create
			plan := buildPlan(t, r, &spec, "")
			state, cr := doCreate(t, r, plan)
			mustOK(t, "create", cr.Diagnostics.HasError(), cr.Diagnostics)
			var gotID types.String
			if d := state.GetAttribute(context.Background(), path.Root("id"), &gotID); d.HasError() {
				t.Fatalf("id: %v", d)
			}
			if gotID.ValueString() != wantID {
				t.Fatalf("id = %q, want %q", gotID.ValueString(), wantID)
			}
			wantCreate := spec.collPath(vals)
			if spec.Kind != "coll" {
				wantCreate = fillPath(spec.Item, vals)
			}
			if call := api.last(spec.Write, wantCreate); call == nil {
				t.Fatalf("no %s %s; calls: %+v", spec.Write, wantCreate, api.all())
			} else {
				for _, a := range spec.Attrs {
					if a.Req && !a.Param {
						if _, ok := call.Body[a.API]; !ok {
							t.Errorf("create body is missing required %s: %v", a.API, call.Body)
						}
					}
					if a.Param || a.Comp {
						if _, ok := call.Body[a.API]; ok {
							t.Errorf("create body carries %s %s", map[bool]string{true: "path key", false: "computed field"}[a.Param], a.API)
						}
					}
				}
			}
			// write-only attributes survive from the plan
			for _, a := range spec.Attrs {
				if a.WriteOnly && !a.Req {
					want := sampleFor(a)
					got := nullFor(a.Kind)
					switch a.Kind {
					case "bool":
						var v types.Bool
						state.GetAttribute(context.Background(), path.Root(a.Name), &v)
						got = v
					case "int":
						var v types.Int64
						state.GetAttribute(context.Background(), path.Root(a.Name), &v)
						got = v
					case "float":
						var v types.Float64
						state.GetAttribute(context.Background(), path.Root(a.Name), &v)
						got = v
					case "strlist":
						var v types.List
						state.GetAttribute(context.Background(), path.Root(a.Name), &v)
						got = v
					default:
						var v types.String
						state.GetAttribute(context.Background(), path.Root(a.Name), &v)
						got = v
					}
					if !want.Equal(got.(attr.Value)) {
						t.Errorf("write-only %s = %v, want the planned %v", a.Name, got, want)
					}
				}
			}

			// read: found keeps the resource, and a 404 removes it
			if !spec.NoRead {
				read, rr := doRead(t, r, state)
				mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
				if read.Raw.IsNull() {
					t.Fatal("read dropped a live resource")
				}
			}

			// update
			if spec.Update != "" {
				plan2 := buildPlan(t, r, &spec, wantID)
				_, ur := doUpdate(t, r, plan2, state)
				mustOK(t, "update", ur.Diagnostics.HasError(), ur.Diagnostics)
				wantUpd := itemPath
				if spec.Kind != "coll" {
					wantUpd = fillPath(spec.Item, vals)
				}
				if api.last(spec.Update, wantUpd) == nil {
					t.Errorf("no %s %s; calls: %+v", spec.Update, wantUpd, api.all())
				}
			}

			// delete
			dr := doDelete(t, r, state)
			mustOK(t, "delete", dr.Diagnostics.HasError(), dr.Diagnostics)
			switch {
			case spec.Revoke != "":
				if api.last(http.MethodPost, itemPath+spec.Revoke) == nil {
					t.Errorf("destroy did not POST %s", itemPath+spec.Revoke)
				}
			case spec.Delete:
				wantDel := itemPath
				if spec.Kind != "coll" {
					wantDel = fillPath(spec.Item, vals)
				}
				if api.last(http.MethodDelete, wantDel) == nil {
					t.Errorf("destroy did not DELETE %s; calls: %+v", wantDel, api.all())
				}
			default:
				for _, call := range api.all() {
					if call.Method == http.MethodDelete {
						t.Errorf("destroy sent a DELETE for a resource without a delete route")
					}
				}
			}
		})
	}
}

func TestGenericReadGoneRemovesResource(t *testing.T) {
	for _, spec := range genResourceSpecs() {
		spec := spec
		if spec.Kind != "coll" || !spec.GetItem {
			continue
		}
		c, _ := newFakeAPI(t, func(call apiCall) (int, interface{}) {
			return 404, map[string]interface{}{"errors": []map[string]string{{"code": "NOT_FOUND", "message": "gone"}}}
		})
		r := newGenResource(&spec)()
		configureResource(t, r, c)
		st := stateFrom404(t, r, &spec)
		out, rr := doRead(t, r, st)
		mustOK(t, "read", rr.Diagnostics.HasError(), rr.Diagnostics)
		if !out.Raw.IsNull() {
			t.Errorf("%s: a 404 should remove the resource from state", spec.Type)
		}
		return // one is enough; the path is shared
	}
}

func stateFrom404(t *testing.T, r resource.Resource, spec *genSpec) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	s := schemaOf(t, r)
	st := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	if d := st.SetAttribute(ctx, path.Root("id"), types.StringValue("obj_gone")); d.HasError() {
		t.Fatal(d)
	}
	for _, p := range spec.params() {
		if d := st.SetAttribute(ctx, path.Root(p.Name), types.StringValue("p-"+p.Name)); d.HasError() {
			t.Fatal(d)
		}
	}
	return st
}

func TestGenericImport(t *testing.T) {
	for _, spec := range genResourceSpecs() {
		spec := spec
		if spec.Kind == "dataonly" {
			continue
		}
		r := newGenResource(&spec)()
		s := schemaOf(t, r)
		ctx := context.Background()
		var id string
		switch spec.Kind {
		case "single":
			id = "anything"
		case "keyed":
			var parts []string
			for _, p := range spec.params() {
				parts = append(parts, "k-"+p.Name)
			}
			id = strings.Join(parts, "/")
		default:
			var parts []string
			for _, p := range spec.params() {
				parts = append(parts, "k-"+p.Name)
			}
			id = strings.Join(append(parts, "obj_9"), "/")
		}
		resp := resource.ImportStateResponse{State: tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}}
		r.(resource.ResourceWithImportState).ImportState(ctx, resource.ImportStateRequest{ID: id}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("%s: import %q: %v", spec.Type, id, resp.Diagnostics)
			continue
		}
		for _, p := range spec.params() {
			var v types.String
			resp.State.GetAttribute(ctx, path.Root(p.Name), &v)
			if v.ValueString() != "k-"+p.Name {
				t.Errorf("%s: imported %s = %q", spec.Type, p.Name, v.ValueString())
			}
		}
		if len(spec.params()) > 0 || spec.Kind == "coll" {
			bad := resource.ImportStateResponse{State: resp.State}
			r.(resource.ResourceWithImportState).ImportState(ctx, resource.ImportStateRequest{ID: "a/b/c/d/e/f"}, &bad)
			if !bad.Diagnostics.HasError() {
				t.Errorf("%s: a malformed import id should be refused", spec.Type)
			}
		}
	}
}
