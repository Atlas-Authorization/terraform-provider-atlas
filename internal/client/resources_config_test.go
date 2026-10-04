package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetInstanceDecodes(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":               "instance",
			"id":                   "ins_123",
			"environment":          "production",
			"publishable_key":      "pk_live_abc",
			"frontend_api_host":    "acme.atlas.dev",
			"allowed_origins":      []string{"https://app.acme.com"},
			"auth_config":          map[string]any{"password": map[string]any{"minLength": 10}},
			"auth_config_resolved": map[string]any{"password": map[string]any{"minLength": 10, "enabled": true}},
			"created_at":           1700000000000,
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	in, err := c.GetInstance(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/v1/instance" {
		t.Errorf("request = %s %s, want GET /v1/instance", gotMethod, gotPath)
	}
	if in.ID != "ins_123" || in.PublishableKey != "pk_live_abc" || in.Environment != "production" {
		t.Errorf("decoded instance mismatch: %+v", in)
	}
	if len(in.AllowedOrigins) != 1 || in.AllowedOrigins[0] != "https://app.acme.com" {
		t.Errorf("allowed_origins = %v", in.AllowedOrigins)
	}
	var ac map[string]any
	if err := json.Unmarshal(in.AuthConfig, &ac); err != nil || ac["password"] == nil {
		t.Errorf("auth_config did not round-trip: %s (%v)", in.AuthConfig, err)
	}
}

func TestUpdateInstanceSendsBody(t *testing.T) {
	var received map[string]any
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "instance", "id": "ins_1"})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.UpdateInstance(context.Background(), InstanceUpdate{
		AllowedOrigins: []string{"https://a.example.com"},
		AuthConfig:     json.RawMessage(`{"session":{"inactivityTimeout":600000}}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", gotMethod)
	}
	origins, ok := received["allowed_origins"].([]any)
	if !ok || len(origins) != 1 || origins[0] != "https://a.example.com" {
		t.Errorf("allowed_origins body = %v", received["allowed_origins"])
	}
	if received["auth_config"] == nil {
		t.Errorf("auth_config missing from body: %v", received)
	}
}

func TestUpdateInstanceOmitsEmptyFields(t *testing.T) {
	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"object":"instance","id":"ins_1"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	if _, err := c.UpdateInstance(context.Background(), InstanceUpdate{AllowedOrigins: []string{"https://x"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, present := received["auth_config"]; present {
		t.Errorf("auth_config should be omitted when nil, body = %v", received)
	}
}

func TestCreateRedirectURLSendsURL(t *testing.T) {
	var received map[string]any
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "redirect_url", "id": "rurl_1", "url": "https://app.acme.com/cb", "created_at": 1,
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	u, err := c.CreateRedirectURL(context.Background(), "https://app.acme.com/cb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v1/redirect_urls" {
		t.Errorf("request = %s %s, want POST /v1/redirect_urls", gotMethod, gotPath)
	}
	if received["url"] != "https://app.acme.com/cb" {
		t.Errorf("body url = %v", received["url"])
	}
	if u.ID != "rurl_1" || u.URL != "https://app.acme.com/cb" {
		t.Errorf("decoded redirect url mismatch: %+v", u)
	}
}

func TestGetRedirectURLPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"redirect_url","id":"rurl_9","url":"https://x","created_at":2}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	if _, err := c.GetRedirectURL(context.Background(), "rurl_9"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v1/redirect_urls/rurl_9" {
		t.Errorf("path = %q, want /v1/redirect_urls/rurl_9", gotPath)
	}
}

func TestBillingPlanCreateAndDecode(t *testing.T) {
	var received map[string]any
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":          "billing_plan",
			"id":              "plan_1",
			"name":            "Pro",
			"slug":            "pro",
			"stripe_price_id": "price_123",
			"free":            false,
			"interval":        "month",
			"amount":          "1500",
			"currency":        "usd",
			"features":        []string{"seats", "sso"},
			"audience":        "org",
			"active":          true,
			"pricing_model":   "per_seat",
			"trial_days":      14,
			"created_at":      10,
			"updated_at":      11,
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	plan, err := c.CreateBillingPlan(context.Background(), BillingPlanWrite{
		Name:          Ptr("Pro"),
		Slug:          Ptr("pro"),
		StripePriceID: Ptr("price_123"),
		Audience:      Ptr("org"),
		Interval:      Ptr("month"),
		Amount:        Ptr("1500"),
		Features:      []string{"seats", "sso"},
		Active:        Ptr(true),
		PricingModel:  Ptr("per_seat"),
		TrialDays:     Ptr(int64(14)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v1/billing/plans" {
		t.Errorf("request = %s %s, want POST /v1/billing/plans", gotMethod, gotPath)
	}
	if received["slug"] != "pro" || received["audience"] != "org" || received["pricing_model"] != "per_seat" {
		t.Errorf("body mismatch: %v", received)
	}
	if plan.ID != "plan_1" || plan.StripePriceID == nil || *plan.StripePriceID != "price_123" {
		t.Errorf("decoded plan mismatch: %+v", plan)
	}
	if plan.Amount == nil || *plan.Amount != "1500" || plan.TrialDays == nil || *plan.TrialDays != 14 {
		t.Errorf("amount/trial decode mismatch: %+v", plan)
	}
	if len(plan.Features) != 2 || plan.Features[0] != "seats" {
		t.Errorf("features = %v", plan.Features)
	}
}

func TestBillingPlanFreeTierOmitsPrice(t *testing.T) {
	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "plan_free", "name": "Free", "slug": "free", "free": true, "currency": "usd", "interval": "month", "audience": "user", "pricing_model": "flat", "active": true, "created_at": 1, "updated_at": 1})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	plan, err := c.CreateBillingPlan(context.Background(), BillingPlanWrite{Name: Ptr("Free"), Slug: Ptr("free")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, present := received["stripe_price_id"]; present {
		t.Errorf("stripe_price_id should be omitted for a free plan, body = %v", received)
	}
	if !plan.Free || plan.StripePriceID != nil {
		t.Errorf("expected free plan with nil price, got %+v", plan)
	}
}

func TestSetOrgPolicySendsPatch(t *testing.T) {
	var received map[string]any
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":          "organization_policy",
			"organization_id": "org_1",
			"policy": map[string]any{
				"requireMfa":           true,
				"allowedSignInMethods": []string{"password", "oauth"},
				"maxSessionAgeSeconds": 3600,
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	policy, err := c.SetOrgPolicy(context.Background(), "org_1", OrgPolicyPatch{
		RequireMfa:           Ptr(true),
		AllowedSignInMethods: []string{"password", "oauth"},
		MaxSessionAgeSeconds: Ptr(int64(3600)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/v1/organizations/org_1/policy" {
		t.Errorf("request = %s %s, want PATCH /v1/organizations/org_1/policy", gotMethod, gotPath)
	}
	if received["requireMfa"] != true {
		t.Errorf("requireMfa body = %v", received["requireMfa"])
	}
	if _, present := received["ssoRequired"]; present {
		t.Errorf("ssoRequired should be omitted when nil, body = %v", received)
	}
	if policy.RequireMfa == nil || !*policy.RequireMfa {
		t.Errorf("decoded requireMfa mismatch: %+v", policy)
	}
	if len(policy.AllowedSignInMethods) != 2 || policy.MaxSessionAgeSeconds == nil || *policy.MaxSessionAgeSeconds != 3600 {
		t.Errorf("decoded policy mismatch: %+v", policy)
	}
}

func TestGetOrgPolicyUsesEmptyPatch(t *testing.T) {
	var received map[string]any
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"organization_id":"org_1","policy":{}}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	policy, err := c.GetOrgPolicy(context.Background(), "org_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("GetOrgPolicy method = %s, want PATCH (empty merge)", gotMethod)
	}
	if len(received) != 0 {
		t.Errorf("GetOrgPolicy should send an empty patch body, got %v", received)
	}
	if policy.RequireMfa != nil || len(policy.AllowedSignInMethods) != 0 {
		t.Errorf("empty policy should decode to all-nil, got %+v", policy)
	}
}
