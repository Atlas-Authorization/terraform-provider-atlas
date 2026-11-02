package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Table-driven resources. Each genSpec (generic_specs_gen.go, generated from the
// Atlas OpenAPI document) describes one Backend API collection, singleton or
// keyed object: its attributes, paths and verbs. One engine implements the
// framework CRUD contract for all of them, so every resource behaves the same:
//
//   - scalars are typed attributes; nested objects and arrays of objects are JSON
//     strings (the same convention as atlas_instance_config.auth_config);
//   - fields the API accepts but never returns (credentials, secrets) are
//     Sensitive and write-only: kept from the plan, never read back;
//   - a JSON attribute only drifts when the server value no longer covers the
//     configured document, so server-added defaults do not thrash the plan;
//   - a 404 on read drops the resource from state (out-of-band delete).

// genAttr is one attribute of a generated resource.
type genAttr struct {
	// Name is the Terraform attribute name; API is the JSON key (snake_case).
	Name, API string
	// Kind is string | bool | int | float | strlist | json.
	Kind string
	// Req: required in config. Comp: response-only (computed). Sens: Sensitive.
	// WriteOnly: sent to the API, never read back. Immut: change forces
	// replacement. Param: a path variable (parent id / key), not sent in a body.
	// Volatile: a computed value that changes on update (no UseStateForUnknown).
	Req, Comp, Sens, WriteOnly, Immut, Param, Volatile bool
	Desc                                               string
}

// genSpec describes one generated resource.
type genSpec struct {
	// Type is the Terraform type name without the provider prefix.
	Type, Desc string
	// Kind: coll (create under a collection, id from the response), single (one
	// object at a fixed path) or keyed (an object addressed only by its path keys).
	Kind string
	// Item is the object path; {id} is the object id and {<attr>} a path-key attribute.
	Item string
	// IDField is the response key holding the object id (default "id").
	IDField string
	// Write is the create method (POST for coll, PUT/POST/PATCH otherwise);
	// Update is the in-place update method ("" = every change replaces).
	Write, Update string
	// GetItem: the object has a GET route. Otherwise reads list the collection and
	// pick the object by IDField.
	GetItem bool
	// Delete: the object has a DELETE route. Otherwise destroy only forgets it.
	Delete bool
	// Pick: "data0" when a create returns {"data":[object]}.
	Pick string
	// DS also publishes a read-only data source of the same name; Kind "dataonly"
	// publishes the data source only.
	DS bool
	// Init: a method sent once, with no body, to the write path before the first
	// write (enable-then-configure singletons); a 409 means already enabled.
	Init string
	// Revoke: a POST suffix used to destroy instead of DELETE (for example /revoke).
	Revoke string
	// List and Match locate the object when it has no GET route: the list path
	// (attribute placeholders allowed) is scanned for an object whose Match field
	// equals the object's key.
	List, Match string
	// NoRead: the API has no read route for the object, so refresh keeps state.
	NoRead bool
	Attrs  []genAttr
}

func (s *genSpec) idField() string {
	if s.IDField == "" {
		return "id"
	}
	return s.IDField
}

func (s *genSpec) params() []genAttr {
	var out []genAttr
	for _, a := range s.Attrs {
		if a.Param {
			out = append(out, a)
		}
	}
	return out
}

var (
	_ resource.Resource                = &genResource{}
	_ resource.ResourceWithConfigure   = &genResource{}
	_ resource.ResourceWithImportState = &genResource{}
)

type genResource struct {
	spec   *genSpec
	client *client.Client
}

func newGenResource(s *genSpec) func() resource.Resource {
	return func() resource.Resource { return &genResource{spec: s} }
}

// genericResourceFactories returns the factory for every generated resource.
func genericResourceFactories() []func() resource.Resource {
	specs := genResourceSpecs()
	out := make([]func() resource.Resource, 0, len(specs))
	for i := range specs {
		if specs[i].Kind == "dataonly" {
			continue
		}
		out = append(out, newGenResource(&specs[i]))
	}
	return out
}

func (r *genResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.spec.Type
}

func (r *genResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *genResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.spec.resourceSchema()
}

func (s *genSpec) resourceSchema() schema.Schema {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Atlas object id (the import id; see the Import section).",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
	}
	for _, a := range s.Attrs {
		attrs[a.Name] = a.build(s.Update == "")
	}
	return schema.Schema{MarkdownDescription: s.Desc, Attributes: attrs}
}

func (a genAttr) build(replaceAll bool) schema.Attribute {
	computed := a.Comp || (!a.Req && !a.WriteOnly && !a.Param)
	optional := !a.Req && !a.Comp
	stable := computed && (!a.Comp || !a.Volatile)
	replace := a.Immut || a.Param || (replaceAll && !a.Comp)
	desc := a.Desc
	switch a.Kind {
	case "bool":
		at := schema.BoolAttribute{Required: a.Req, Optional: optional, Computed: computed, Sensitive: a.Sens, MarkdownDescription: desc}
		if stable {
			at.PlanModifiers = append(at.PlanModifiers, boolplanmodifier.UseStateForUnknown())
		}
		if replace {
			at.PlanModifiers = append(at.PlanModifiers, boolplanmodifier.RequiresReplace())
		}
		return at
	case "int":
		at := schema.Int64Attribute{Required: a.Req, Optional: optional, Computed: computed, Sensitive: a.Sens, MarkdownDescription: desc}
		if stable {
			at.PlanModifiers = append(at.PlanModifiers, int64planmodifier.UseStateForUnknown())
		}
		if replace {
			at.PlanModifiers = append(at.PlanModifiers, int64planmodifier.RequiresReplace())
		}
		return at
	case "float":
		at := schema.Float64Attribute{Required: a.Req, Optional: optional, Computed: computed, Sensitive: a.Sens, MarkdownDescription: desc}
		if stable {
			at.PlanModifiers = append(at.PlanModifiers, float64planmodifier.UseStateForUnknown())
		}
		if replace {
			at.PlanModifiers = append(at.PlanModifiers, float64planmodifier.RequiresReplace())
		}
		return at
	case "strlist":
		at := schema.ListAttribute{ElementType: types.StringType, Required: a.Req, Optional: optional, Computed: computed, Sensitive: a.Sens, MarkdownDescription: desc}
		if stable {
			at.PlanModifiers = append(at.PlanModifiers, listplanmodifier.UseStateForUnknown())
		}
		if replace {
			at.PlanModifiers = append(at.PlanModifiers, listplanmodifier.RequiresReplace())
		}
		return at
	default: // string and json
		at := schema.StringAttribute{Required: a.Req, Optional: optional, Computed: computed, Sensitive: a.Sens, MarkdownDescription: desc}
		if stable {
			at.PlanModifiers = append(at.PlanModifiers, stringplanmodifier.UseStateForUnknown())
		}
		if replace {
			at.PlanModifiers = append(at.PlanModifiers, stringplanmodifier.RequiresReplace())
		}
		return at
	}
}

// ── plan / state <-> JSON ───────────────────────────────────────────────────

type attrGetter interface {
	GetAttribute(ctx context.Context, p path.Path, target interface{}) diag.Diagnostics
}

// fromModel reads the attribute's configured value as a JSON-encodable Go value.
// ok is false for null and unknown values (the field is then omitted).
func (a genAttr) fromModel(ctx context.Context, g attrGetter, diags *diag.Diagnostics) (any, bool) {
	p := path.Root(a.Name)
	switch a.Kind {
	case "bool":
		var v types.Bool
		diags.Append(g.GetAttribute(ctx, p, &v)...)
		if v.IsNull() || v.IsUnknown() {
			return nil, false
		}
		return v.ValueBool(), true
	case "int":
		var v types.Int64
		diags.Append(g.GetAttribute(ctx, p, &v)...)
		if v.IsNull() || v.IsUnknown() {
			return nil, false
		}
		return v.ValueInt64(), true
	case "float":
		var v types.Float64
		diags.Append(g.GetAttribute(ctx, p, &v)...)
		if v.IsNull() || v.IsUnknown() {
			return nil, false
		}
		return v.ValueFloat64(), true
	case "strlist":
		var v types.List
		diags.Append(g.GetAttribute(ctx, p, &v)...)
		if v.IsNull() || v.IsUnknown() {
			return nil, false
		}
		xs := []string{}
		diags.Append(v.ElementsAs(ctx, &xs, false)...)
		return xs, true
	case "json":
		var v types.String
		diags.Append(g.GetAttribute(ctx, p, &v)...)
		if v.IsNull() || v.IsUnknown() || strings.TrimSpace(v.ValueString()) == "" {
			return nil, false
		}
		var out any
		if err := json.Unmarshal([]byte(v.ValueString()), &out); err != nil {
			diags.AddAttributeError(p, "Invalid JSON", fmt.Sprintf("%s must be a JSON document: %s", a.Name, err))
			return nil, false
		}
		return out, true
	default:
		var v types.String
		diags.Append(g.GetAttribute(ctx, p, &v)...)
		if v.IsNull() || v.IsUnknown() {
			return nil, false
		}
		return v.ValueString(), true
	}
}

func isJSONNull(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s == "" || s == "null"
}

// toState writes an API value into the state attribute.
func (a genAttr) toState(ctx context.Context, st *tfsdk.State, raw json.RawMessage, diags *diag.Diagnostics) {
	p := path.Root(a.Name)
	null := isJSONNull(raw)
	switch a.Kind {
	case "bool":
		var b bool
		if null || json.Unmarshal(raw, &b) != nil {
			diags.Append(st.SetAttribute(ctx, p, types.BoolNull())...)
			return
		}
		diags.Append(st.SetAttribute(ctx, p, types.BoolValue(b))...)
	case "int":
		var f float64
		if null || json.Unmarshal(raw, &f) != nil {
			diags.Append(st.SetAttribute(ctx, p, types.Int64Null())...)
			return
		}
		diags.Append(st.SetAttribute(ctx, p, types.Int64Value(int64(f)))...)
	case "float":
		var f float64
		if null || json.Unmarshal(raw, &f) != nil {
			diags.Append(st.SetAttribute(ctx, p, types.Float64Null())...)
			return
		}
		diags.Append(st.SetAttribute(ctx, p, types.Float64Value(f))...)
	case "strlist":
		var xs []string
		if null || json.Unmarshal(raw, &xs) != nil {
			diags.Append(st.SetAttribute(ctx, p, types.ListNull(types.StringType))...)
			return
		}
		l, d := types.ListValueFrom(ctx, types.StringType, xs)
		diags.Append(d...)
		diags.Append(st.SetAttribute(ctx, p, l)...)
	case "json":
		if null {
			diags.Append(st.SetAttribute(ctx, p, types.StringNull())...)
			return
		}
		var server any
		if err := json.Unmarshal(raw, &server); err != nil {
			diags.Append(st.SetAttribute(ctx, p, types.StringValue(string(raw)))...)
			return
		}
		var prior types.String
		diags.Append(st.GetAttribute(ctx, p, &prior)...)
		if !prior.IsNull() && !prior.IsUnknown() {
			var cfg any
			if json.Unmarshal([]byte(prior.ValueString()), &cfg) == nil && jsonCovers(cfg, server) {
				return // the server value still satisfies what is configured
			}
		}
		canon, _ := json.Marshal(server)
		diags.Append(st.SetAttribute(ctx, p, types.StringValue(string(canon)))...)
	default:
		if null {
			diags.Append(st.SetAttribute(ctx, p, types.StringNull())...)
			return
		}
		var s string
		if json.Unmarshal(raw, &s) != nil {
			s = strings.TrimSpace(string(raw))
		}
		diags.Append(st.SetAttribute(ctx, p, types.StringValue(s))...)
	}
}

// jsonCovers reports whether server contains everything in cfg: same scalars,
// every configured object key present and covered, arrays element-wise equal in
// length. Extra server keys (defaults the API fills in) are allowed. A configured
// null is satisfied by an absent or null server value.
func jsonCovers(cfg, server any) bool {
	switch c := cfg.(type) {
	case nil:
		return server == nil
	case map[string]any:
		s, ok := server.(map[string]any)
		if !ok {
			return false
		}
		for k, cv := range c {
			sv, present := s[k]
			if !present {
				if cv == nil {
					continue
				}
				return false
			}
			if !jsonCovers(cv, sv) {
				return false
			}
		}
		return true
	case []any:
		s, ok := server.([]any)
		if !ok || len(s) != len(c) {
			return false
		}
		for i := range c {
			if !jsonCovers(c[i], s[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(cfg, server)
	}
}

func nullFor(kind string) any {
	switch kind {
	case "bool":
		return types.BoolNull()
	case "int":
		return types.Int64Null()
	case "float":
		return types.Float64Null()
	case "strlist":
		return types.ListNull(types.StringType)
	}
	return types.StringNull()
}

// ── paths ───────────────────────────────────────────────────────────────────

func fillPath(tpl string, vals map[string]string) string {
	for k, v := range vals {
		tpl = strings.ReplaceAll(tpl, "{"+k+"}", url.PathEscape(v))
	}
	return tpl
}

func (s *genSpec) collPath(vals map[string]string) string {
	return fillPath(strings.TrimSuffix(s.Item, "/{id}"), vals)
}

// paramVals collects the path-key attributes (and id, when non-empty).
func (s *genSpec) paramVals(ctx context.Context, g attrGetter, id string, diags *diag.Diagnostics) map[string]string {
	vals := map[string]string{}
	for _, a := range s.params() {
		var v types.String
		diags.Append(g.GetAttribute(ctx, path.Root(a.Name), &v)...)
		vals[a.Name] = v.ValueString()
	}
	if id != "" {
		vals["id"] = id
	}
	return vals
}

func rawString(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return strings.Trim(strings.TrimSpace(string(raw)), `"`)
}

// ── CRUD ────────────────────────────────────────────────────────────────────

type genObject map[string]json.RawMessage

func (r *genResource) body(ctx context.Context, g attrGetter, diags *diag.Diagnostics) map[string]any {
	body := map[string]any{}
	for _, a := range r.spec.Attrs {
		if a.Comp || a.Param {
			continue
		}
		if v, ok := a.fromModel(ctx, g, diags); ok {
			body[a.API] = v
		}
	}
	return body
}

// apply copies an API object into the state. After a write (fill=true) only the
// attributes the plan left unknown are filled, so a server-normalised value can
// never contradict the plan (the next read reports it as drift instead); a read
// sets every readable attribute. Any value still unknown becomes null.
func (r *genResource) apply(ctx context.Context, st *tfsdk.State, obj genObject, id string, fill bool, diags *diag.Diagnostics) {
	diags.Append(st.SetAttribute(ctx, path.Root("id"), types.StringValue(id))...)
	var before map[string]tftypes.Value
	if err := st.Raw.As(&before); err != nil {
		diags.AddError("Unable to read state", err.Error())
		return
	}
	for _, a := range r.spec.Attrs {
		if a.Param || a.WriteOnly {
			continue
		}
		if v, ok := before[a.Name]; fill && ok && v.IsKnown() {
			continue
		}
		if a.API == "__raw" {
			b, _ := json.Marshal(obj)
			a.toState(ctx, st, b, diags)
			continue
		}
		if raw, ok := obj[a.API]; ok {
			a.toState(ctx, st, raw, diags)
		}
	}
	var all map[string]tftypes.Value
	if err := st.Raw.As(&all); err != nil {
		diags.AddError("Unable to read state", err.Error())
		return
	}
	for _, a := range r.spec.Attrs {
		if v, ok := all[a.Name]; ok && !v.IsKnown() {
			diags.Append(st.SetAttribute(ctx, path.Root(a.Name), nullFor(a.Kind))...)
		}
	}
}

func (r *genResource) decode(out json.RawMessage) genObject {
	obj := genObject{}
	if len(out) == 0 {
		return obj
	}
	_ = json.Unmarshal(out, &obj)
	if r.spec.Pick == "data0" {
		var items []genObject
		if raw, ok := obj["data"]; ok && json.Unmarshal(raw, &items) == nil && len(items) > 0 {
			return items[0]
		}
	}
	return obj
}

func (r *genResource) write(ctx context.Context, g attrGetter, st *tfsdk.State, create bool, diags *diag.Diagnostics) {
	s := r.spec
	body := r.body(ctx, g, diags)
	if diags.HasError() {
		return
	}
	var method, p, id string
	switch s.Kind {
	case "coll":
		if create {
			vals := s.paramVals(ctx, g, "", diags)
			method, p = s.Write, s.collPath(vals)
		} else {
			var cur types.String
			diags.Append(st.GetAttribute(ctx, path.Root("id"), &cur)...)
			id = cur.ValueString()
			method, p = s.Update, fillPath(s.Item, s.paramVals(ctx, g, id, diags))
		}
	default:
		vals := s.paramVals(ctx, g, "", diags)
		method, p = s.Write, fillPath(s.Item, vals)
		if !create && s.Update != "" {
			method = s.Update
		}
		id = s.Type
		if s.Kind == "keyed" {
			parts := []string{}
			for _, a := range s.params() {
				parts = append(parts, vals[a.Name])
			}
			id = strings.Join(parts, "/")
		}
	}
	if diags.HasError() {
		return
	}
	if create && s.Init != "" {
		if err := r.client.Call(ctx, s.Init, p, nil, nil); err != nil {
			if apiErr, ok := err.(*client.APIError); !ok || apiErr.Status != http.StatusConflict {
				diags.AddError("Unable to enable "+s.Type, err.Error())
				return
			}
		}
	}
	var out json.RawMessage
	if err := r.client.Call(ctx, method, p, body, &out); err != nil {
		diags.AddError("Unable to write "+s.Type, err.Error())
		return
	}
	obj := r.decode(out)
	if s.Kind == "coll" && create {
		id = rawString(obj[s.idField()])
		if id == "" {
			diags.AddError("Unable to create "+s.Type, "the API response carried no `"+s.idField()+"`")
			return
		}
	}
	if s.Kind == "single" && s.GetItem {
		var got json.RawMessage
		if err := r.client.Call(ctx, http.MethodGet, p, nil, &got); err == nil {
			obj = r.decode(got)
		}
	}
	r.apply(ctx, st, obj, id, true, diags)
}

func (r *genResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.State.Raw = req.Plan.Raw
	r.write(ctx, req.Plan, &resp.State, true, &resp.Diagnostics)
}

func (r *genResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.spec.Update == "" {
		resp.Diagnostics.AddError("Update not supported", r.spec.Type+" cannot be updated in place; change forces replacement.")
		return
	}
	resp.State.Raw = req.Plan.Raw
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	r.write(ctx, req.Plan, &resp.State, false, &resp.Diagnostics)
}

// fetch loads the object: GET on its path, else a paged scan of the collection.
func (r *genResource) fetch(ctx context.Context, vals map[string]string, id string) (genObject, bool, error) {
	s := r.spec
	if s.Kind == "keyed" && !s.GetItem {
		if ps := s.params(); len(ps) > 0 {
			id = vals[ps[len(ps)-1].Name]
		}
	}
	if s.GetItem {
		var out json.RawMessage
		if err := r.client.Call(ctx, http.MethodGet, fillPath(s.Item, vals), nil, &out); err != nil {
			if client.IsNotFound(err) {
				return nil, false, nil
			}
			return nil, false, err
		}
		return r.decode(out), true, nil
	}
	after := ""
	for page := 0; page < 200; page++ {
		q := "?limit=100"
		if after != "" {
			q += "&starting_after=" + url.QueryEscape(after)
		}
		var list struct {
			Data    []genObject `json:"data"`
			HasMore bool        `json:"has_more"`
		}
		if err := r.client.Call(ctx, http.MethodGet, fillPath(s.List, vals)+q, nil, &list); err != nil {
			if client.IsNotFound(err) {
				return nil, false, nil
			}
			return nil, false, err
		}
		for _, o := range list.Data {
			if rawString(o[s.Match]) == id {
				return o, true, nil
			}
		}
		if !list.HasMore || len(list.Data) == 0 {
			return nil, false, nil
		}
		after = rawString(list.Data[len(list.Data)-1]["id"])
		if after == "" {
			return nil, false, nil
		}
	}
	return nil, false, nil
}

func (r *genResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if r.spec.NoRead {
		return
	}
	vals := r.spec.paramVals(ctx, req.State, id.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	obj, found, err := r.fetch(ctx, vals, id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read "+r.spec.Type, err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	r.apply(ctx, &resp.State, obj, id.ValueString(), false, &resp.Diagnostics)
}

func (r *genResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.spec.Delete && r.spec.Revoke == "" {
		return
	}
	var id types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	vals := r.spec.paramVals(ctx, req.State, id.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	method, target := http.MethodDelete, fillPath(r.spec.Item, vals)
	if r.spec.Revoke != "" {
		method, target = http.MethodPost, target+r.spec.Revoke
	}
	if err := r.client.Call(ctx, method, target, nil, nil); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete "+r.spec.Type, err.Error())
	}
}

// ImportState: coll = "<parent keys…>/<id>", keyed = "<keys…>", single = any value.
func (r *genResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	s := r.spec
	params := s.params()
	if s.Kind == "single" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(s.Type))...)
		return
	}
	want := len(params)
	if s.Kind == "coll" {
		want++
	}
	parts := strings.Split(req.ID, "/")
	if len(parts) != want {
		names := []string{}
		for _, a := range params {
			names = append(names, "<"+a.Name+">")
		}
		if s.Kind == "coll" {
			names = append(names, "<id>")
		}
		resp.Diagnostics.AddError("Invalid import id", "expected "+strings.Join(names, "/")+", got "+strconv.Quote(req.ID))
		return
	}
	for i, a := range params {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(a.Name), types.StringValue(parts[i]))...)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(parts[want-1]))...)
	if s.Kind == "keyed" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	}
}
