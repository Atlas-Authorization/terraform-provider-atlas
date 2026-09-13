package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &jwtTemplateResource{}
	_ resource.ResourceWithImportState = &jwtTemplateResource{}
	_ resource.ResourceWithConfigure   = &jwtTemplateResource{}
)

// NewJwtTemplateResource is the factory registered with the provider.
func NewJwtTemplateResource() resource.Resource { return &jwtTemplateResource{} }

type jwtTemplateResource struct {
	client *client.Client
}

type jwtTemplateModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Claims types.Map    `tfsdk:"claims"`
}

func (r *jwtTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jwt_template"
}

func (r *jwtTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *jwtTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A JWT template: a named map of custom claims injected into issued tokens. The name is " +
			"the key and is immutable; only the claims map is updatable.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Equal to the template name.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Template name — the key. Immutable; changing it forces replacement.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"claims": schema.MapAttribute{
				MarkdownDescription: "Claim name → value-template map. Reserved system claims are rejected by the API.",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *jwtTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan jwtTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	claims := mapToStringMap(ctx, plan.Claims, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateJwtTemplate(ctx, plan.Name.ValueString(), claims)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create JWT template", err.Error())
		return
	}
	r.mapToState(ctx, created, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *jwtTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state jwtTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetJwtTemplate(ctx, state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read JWT template", err.Error())
		return
	}
	r.mapToState(ctx, fetched, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *jwtTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan jwtTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	claims := mapToStringMap(ctx, plan.Claims, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateJwtTemplate(ctx, plan.Name.ValueString(), claims)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update JWT template", err.Error())
		return
	}
	r.mapToState(ctx, updated, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *jwtTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state jwtTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteJwtTemplate(ctx, state.Name.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete JWT template", err.Error())
	}
}

// ImportState takes the template name (which is also the id).
func (r *jwtTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *jwtTemplateResource) mapToState(ctx context.Context, t *client.JwtTemplate, m *jwtTemplateModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue(t.Name)
	m.Name = types.StringValue(t.Name)
	m.Claims = stringMapToMap(ctx, t.Claims, diags)
}
