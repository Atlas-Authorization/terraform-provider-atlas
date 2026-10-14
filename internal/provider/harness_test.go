package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	pschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// This file is an offline harness that drives a resource's CRUD hooks the way
// the framework does (plan in, state out) against an httptest fake of the
// Atlas API, recording every request. It complements the TF_ACC acceptance
// tests, which need a live instance and a terraform binary.

// apiCall is one request the fake API received.
type apiCall struct {
	Method string
	Path   string
	Query  string
	Auth   string
	Body   map[string]interface{}
}

// fakeAPI answers requests with a handler and records them.
type fakeAPI struct {
	mu    sync.Mutex
	calls []apiCall
}

func (f *fakeAPI) all() []apiCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]apiCall(nil), f.calls...)
}

// last returns the most recent call matching method+path, or nil.
func (f *fakeAPI) last(method, path string) *apiCall {
	calls := f.all()
	for i := len(calls) - 1; i >= 0; i-- {
		if calls[i].Method == method && calls[i].Path == path {
			return &calls[i]
		}
	}
	return nil
}

func (f *fakeAPI) count(method, path string) int {
	n := 0
	for _, c := range f.all() {
		if c.Method == method && c.Path == path {
			n++
		}
	}
	return n
}

// newFakeAPI starts a fake server; handle returns (status, JSON body).
func newFakeAPI(t *testing.T, handle func(c apiCall) (int, interface{})) (*client.Client, *fakeAPI) {
	t.Helper()
	f := &fakeAPI{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		call := apiCall{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Auth: r.Header.Get("Authorization")}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &call.Body)
		}
		f.mu.Lock()
		f.calls = append(f.calls, call)
		f.mu.Unlock()
		status, body := handle(call)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
	t.Cleanup(srv.Close)
	c := client.New(srv.URL, "sk_test_key")
	c.PlatformKey = "psk_test_key"
	c.HTTP = srv.Client()
	return c, f
}

func configureResource(t *testing.T, r resource.Resource, c *client.Client) {
	t.Helper()
	rc, ok := r.(resource.ResourceWithConfigure)
	if !ok {
		t.Fatal("resource does not implement Configure")
	}
	var resp resource.ConfigureResponse
	rc.Configure(context.Background(), resource.ConfigureRequest{ProviderData: c}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("configure: %v", resp.Diagnostics)
	}
}

func nullState(t *testing.T, r resource.Resource) tfsdk.State {
	s := schemaOf(t, r)
	return tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)}
}

// planFrom builds a plan from a typed model (the model must cover every attribute).
func planFrom(t *testing.T, r resource.Resource, model interface{}) tfsdk.Plan {
	t.Helper()
	s := schemaOf(t, r)
	p := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)}
	if d := p.Set(context.Background(), model); d.HasError() {
		t.Fatalf("plan set: %v", d)
	}
	return p
}

func stateFrom(t *testing.T, r resource.Resource, model interface{}) tfsdk.State {
	t.Helper()
	st := nullState(t, r)
	if d := st.Set(context.Background(), model); d.HasError() {
		t.Fatalf("state set: %v", d)
	}
	return st
}

func doCreate(t *testing.T, r resource.Resource, plan tfsdk.Plan) (tfsdk.State, resource.CreateResponse) {
	t.Helper()
	resp := resource.CreateResponse{State: nullState(t, r)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan, Config: tfsdk.Config{Schema: plan.Schema, Raw: plan.Raw}}, &resp)
	return resp.State, resp
}

func doRead(t *testing.T, r resource.Resource, state tfsdk.State) (tfsdk.State, resource.ReadResponse) {
	t.Helper()
	resp := resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
	return resp.State, resp
}

func doUpdate(t *testing.T, r resource.Resource, plan tfsdk.Plan, prior tfsdk.State) (tfsdk.State, resource.UpdateResponse) {
	t.Helper()
	resp := resource.UpdateResponse{State: prior}
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan, State: prior, Config: tfsdk.Config{Schema: plan.Schema, Raw: plan.Raw}}, &resp)
	return resp.State, resp
}

func doDelete(t *testing.T, r resource.Resource, state tfsdk.State) resource.DeleteResponse {
	t.Helper()
	resp := resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)
	return resp
}

func mustOK(t *testing.T, what string, hasError bool, diags interface{}) {
	t.Helper()
	if hasError {
		t.Fatalf("%s: %v", what, diags)
	}
}

func readDataSource(t *testing.T, d datasource.DataSource, c *client.Client) (tfsdk.State, datasource.ReadResponse) {
	t.Helper()
	ctx := context.Background()
	if dc, ok := d.(datasource.DataSourceWithConfigure); ok {
		var cr datasource.ConfigureResponse
		dc.Configure(ctx, datasource.ConfigureRequest{ProviderData: c}, &cr)
	}
	var sr datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &sr)
	null := tftypes.NewValue(sr.Schema.Type().TerraformType(ctx), nil)
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: sr.Schema, Raw: null}}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: sr.Schema, Raw: null}}, &resp)
	return resp.State, resp
}

func providerSchema(t *testing.T, p provider.Provider) pschema.Schema {
	t.Helper()
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)
	return resp.Schema
}

func dataSourceTypeName(d datasource.DataSource) string {
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "atlas"}, &resp)
	return resp.TypeName
}
