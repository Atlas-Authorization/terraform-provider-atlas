package provider

import (
	"context"
	"encoding/json"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// designResource manages a sign-in page design: the instance design
// (atlas_instance_design, a singleton) or one organization's override
// (atlas_organization_design). Both are a whole-document PUT of `{ design }`;
// destroy writes `{ design: {} }`, which removes the stored design.
type designResource struct {
	client *client.Client
	org    bool
}

var (
	_ resource.Resource                = &designResource{}
	_ resource.ResourceWithImportState = &designResource{}
	_ resource.ResourceWithConfigure   = &designResource{}
)

// NewInstanceDesignResource is the factory for atlas_instance_design.
func NewInstanceDesignResource() resource.Resource { return &designResource{} }

// NewOrganizationDesignResource is the factory for atlas_organization_design.
func NewOrganizationDesignResource() resource.Resource { return &designResource{org: true} }

type designModel struct {
	ID             types.String    `tfsdk:"id"`
	OrganizationID types.String    `tfsdk:"organization_id"`
	Design         jsonObjectValue `tfsdk:"design"`
	Resolved       types.String    `tfsdk:"resolved"`
	Warnings       types.String    `tfsdk:"warnings"`
}

// instanceDesignModel is designModel without organization_id.
type instanceDesignModel struct {
	ID       types.String    `tfsdk:"id"`
	Design   jsonObjectValue `tfsdk:"design"`
	Resolved types.String    `tfsdk:"resolved"`
	Warnings types.String    `tfsdk:"warnings"`
}

func (r *designResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	if r.org {
		resp.TypeName = req.ProviderTypeName + "_organization_design"
		return
	}
	resp.TypeName = req.ProviderTypeName + "_instance_design"
}

func (r *designResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *designResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"design": schema.StringAttribute{
			CustomType: jsonObjectType{},
			MarkdownDescription: "The design settings as a JSON object (use `jsonencode`): `preset`, `colorScheme`, " +
				"`palette`, `paletteDark`, `typography`, `shape`, `card`, `layout`, `buttons`, `inputs`, `background`, " +
				"`logo`, … — the same document the dashboard design editor saves. Unknown keys are refused. Compared " +
				"semantically, so the server's key order, formatting and null members never cause drift.",
			Required: true,
		},
		"resolved": schema.StringAttribute{
			MarkdownDescription: "The fully-resolved design every page renders (defaults ← preset ← your settings), as JSON.",
			Computed:            true,
		},
		"warnings": schema.StringAttribute{
			MarkdownDescription: "Contrast and legibility warnings for the resolved design, as a JSON array.",
			Computed:            true,
		},
	}
	desc := "The hosted sign-in page design for the instance (a singleton). Replaces the whole stored design on " +
		"every apply; destroying the resource removes the stored design (pages fall back to the default look). " +
		"Import with any id, e.g. `instance`."
	if r.org {
		attrs["organization_id"] = schema.StringAttribute{
			MarkdownDescription: "The organization whose design override this manages. Changing it forces replacement.",
			Required:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
		}
		desc = "An organization's design override, layered over the instance design on that organization's " +
			"pages. Replaces the whole override on every apply; destroying it removes the override. Import by " +
			"organization id."
	}
	resp.Schema = schema.Schema{MarkdownDescription: desc, Attributes: attrs}
}

func (r *designResource) put(ctx context.Context, m *designModel) (*client.Design, error) {
	raw := json.RawMessage(`{}`)
	if !m.Design.IsNull() && !m.Design.IsUnknown() {
		raw = json.RawMessage(m.Design.ValueString())
	}
	if r.org {
		return r.client.PutOrganizationDesign(ctx, m.OrganizationID.ValueString(), raw)
	}
	return r.client.PutInstanceDesign(ctx, raw)
}

func (r *designResource) get(ctx context.Context, m *designModel) (*client.Design, error) {
	if r.org {
		return r.client.GetOrganizationDesign(ctx, m.OrganizationID.ValueString())
	}
	return r.client.GetInstanceDesign(ctx)
}

func (r *designResource) mapToModel(d *client.Design, m *designModel) {
	if r.org {
		m.ID = m.OrganizationID
	} else {
		m.ID = types.StringValue("instance")
	}
	m.Design = jsonObjectFromAPI(d.Design)
	m.Resolved = jsonRawToValue(d.Resolved)
	m.Warnings = jsonRawToValue(d.Warnings)
}

// fromPlan reads the plan/state into the shared model. The instance variant has
// no organization_id attribute, so it is read through its own struct.
func (r *designResource) fromPlan(ctx context.Context, get func(context.Context, interface{}) diag.Diagnostics, diags *diag.Diagnostics) (*designModel, bool) {
	if r.org {
		var m designModel
		diags.Append(get(ctx, &m)...)
		return &m, !diags.HasError()
	}
	var im instanceDesignModel
	diags.Append(get(ctx, &im)...)
	return &designModel{ID: im.ID, OrganizationID: types.StringNull(), Design: im.Design, Resolved: im.Resolved, Warnings: im.Warnings}, !diags.HasError()
}

// setState writes the shared model back in the variant's shape.
func (r *designResource) setState(ctx context.Context, m *designModel, set func(context.Context, interface{}) diag.Diagnostics, diags *diag.Diagnostics) {
	if r.org {
		diags.Append(set(ctx, m)...)
		return
	}
	diags.Append(set(ctx, &instanceDesignModel{ID: m.ID, Design: m.Design, Resolved: m.Resolved, Warnings: m.Warnings})...)
}

func (r *designResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	m, ok := r.fromPlan(ctx, req.Plan.Get, &resp.Diagnostics)
	if !ok {
		return
	}
	d, err := r.put(ctx, m)
	if err != nil {
		resp.Diagnostics.AddError("Unable to set design", err.Error())
		return
	}
	r.mapToModel(d, m)
	r.setState(ctx, m, resp.State.Set, &resp.Diagnostics)
}

func (r *designResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	m, ok := r.fromPlan(ctx, req.State.Get, &resp.Diagnostics)
	if !ok {
		return
	}
	d, err := r.get(ctx, m)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read design", err.Error())
		return
	}
	r.mapToModel(d, m)
	r.setState(ctx, m, resp.State.Set, &resp.Diagnostics)
}

func (r *designResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	m, ok := r.fromPlan(ctx, req.Plan.Get, &resp.Diagnostics)
	if !ok {
		return
	}
	d, err := r.put(ctx, m)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update design", err.Error())
		return
	}
	r.mapToModel(d, m)
	r.setState(ctx, m, resp.State.Set, &resp.Diagnostics)
}

func (r *designResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	m, ok := r.fromPlan(ctx, req.State.Get, &resp.Diagnostics)
	if !ok {
		return
	}
	m.Design = newJSONObjectNull()
	if _, err := r.put(ctx, m); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to remove design", err.Error())
	}
}

func (r *designResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if r.org {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), req.ID)...)
	}
}
