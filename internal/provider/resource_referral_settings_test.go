package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestReferralSettings_partnerPortalAndLinkedSignups(t *testing.T) {
	ctx := context.Background()
	c, api := newFakeAPI(t, func(call apiCall) (int, interface{}) {
		return 200, map[string]interface{}{
			"object": "referral_settings", "enabled": true, "provider": "webhook",
			"partner_portal_url": "https://partners.example.com", "report_linked_atlas_id_signups": true,
			"security": map[string]interface{}{"challenge_referred_signups": "always", "on_high_risk": "flag", "max_referred_signups_per_ip_hour": 5},
		}
	})
	r := &referralSettingsResource{client: c}
	cfg := referralSettingsModel{
		Enabled:                    types.BoolValue(true),
		PartnerPortalURL:           types.StringValue("https://partners.example.com"),
		ReportLinkedAtlasIDSignups: types.BoolValue(true),
	}
	plan := cfg
	var diags diag.Diagnostics
	r.apply(ctx, &plan, &cfg, nil, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	patch := api.last(http.MethodPatch, "/v1/instance/referrals")
	if patch.Body["partner_portal_url"] != "https://partners.example.com" || patch.Body["report_linked_atlas_id_signups"] != true {
		t.Errorf("PATCH body = %v", patch.Body)
	}
	if plan.PartnerPortalURL.ValueString() != "https://partners.example.com" || !plan.ReportLinkedAtlasIDSignups.ValueBool() {
		t.Errorf("read back = %+v", plan)
	}

	// Left unset: neither key is sent (Atlas keeps its values).
	empty := referralSettingsModel{Enabled: types.BoolValue(true)}
	plan = empty
	r.apply(ctx, &plan, &empty, nil, &diags)
	patch = api.last(http.MethodPatch, "/v1/instance/referrals")
	for _, k := range []string{"partner_portal_url", "report_linked_atlas_id_signups"} {
		if _, sent := patch.Body[k]; sent {
			t.Errorf("unset %s must not be sent", k)
		}
	}
}
