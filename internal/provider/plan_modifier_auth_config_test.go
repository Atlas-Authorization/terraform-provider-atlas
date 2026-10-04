package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// runAuthConfigModifier drives the modifier the way the framework does: the
// response's PlanValue is pre-seeded with the planned value, then the modifier
// may overwrite it to suppress a diff.
func runAuthConfigModifier(config, state, plan types.String) types.String {
	req := planmodifier.StringRequest{
		ConfigValue: config,
		StateValue:  state,
		PlanValue:   plan,
	}
	resp := &planmodifier.StringResponse{PlanValue: plan}
	authConfigSubsetModifier{}.PlanModifyString(context.Background(), req, resp)
	return resp.PlanValue
}

func TestAuthConfigSubsetModifier(t *testing.T) {
	tests := []struct {
		name   string
		config types.String
		state  types.String
		plan   types.String
		want   types.String // expected resulting plan value
	}{
		{
			name:   "subset suppressed",
			config: types.StringValue(`{"a":1}`),
			state:  types.StringValue(`{"a":1,"b":2}`),
			plan:   types.StringValue(`{"a":1}`),
			// config ⊆ state → plan copies the server-merged state value.
			want: types.StringValue(`{"a":1,"b":2}`),
		},
		{
			name:   "superset with changed value shows diff",
			config: types.StringValue(`{"a":2}`),
			state:  types.StringValue(`{"a":1,"b":2}`),
			plan:   types.StringValue(`{"a":2}`),
			// a changed → not a subset → plan left as config.
			want: types.StringValue(`{"a":2}`),
		},
		{
			name:   "added key shows diff",
			config: types.StringValue(`{"a":1,"c":3}`),
			state:  types.StringValue(`{"a":1,"b":2}`),
			plan:   types.StringValue(`{"a":1,"c":3}`),
			// c not in state → not a subset → plan left as config.
			want: types.StringValue(`{"a":1,"c":3}`),
		},
		{
			name:   "null config skipped",
			config: types.StringNull(),
			state:  types.StringValue(`{"a":1,"b":2}`),
			plan:   types.StringNull(),
			// nothing to do → plan untouched.
			want: types.StringNull(),
		},
		{
			name:   "invalid json exact compare equal",
			config: types.StringValue(`not json`),
			state:  types.StringValue(`not json`),
			plan:   types.StringValue(`not json`),
			// non-object → exact compare; identical → suppressed.
			want: types.StringValue(`not json`),
		},
		{
			name:   "invalid json exact compare differ",
			config: types.StringValue(`not json`),
			state:  types.StringValue(`other junk`),
			plan:   types.StringValue(`not json`),
			// non-object → exact compare; differ → plan left as config.
			want: types.StringValue(`not json`),
		},
		{
			name:   "no prior state skipped",
			config: types.StringValue(`{"a":1}`),
			state:  types.StringNull(),
			plan:   types.StringValue(`{"a":1}`),
			// nothing to suppress against → plan untouched.
			want: types.StringValue(`{"a":1}`),
		},
		{
			name:   "nested value difference shows diff",
			config: types.StringValue(`{"session":{"timeout":10}}`),
			state:  types.StringValue(`{"session":{"timeout":20},"b":2}`),
			plan:   types.StringValue(`{"session":{"timeout":10}}`),
			// session value differs deeply → not a subset → diff.
			want: types.StringValue(`{"session":{"timeout":10}}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runAuthConfigModifier(tt.config, tt.state, tt.plan)
			if !got.Equal(tt.want) {
				t.Errorf("plan value = %v, want %v", got, tt.want)
			}
		})
	}
}
