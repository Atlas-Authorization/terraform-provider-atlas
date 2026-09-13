package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &resourceServerResource{}
	_ resource.ResourceWithImportState = &resourceServerResource{}
	_ resource.ResourceWithConfigure   = &resourceServerResource{}
)

// NewResourceServerResource is the factory registered with the provider.
func NewResourceServerResource() resource.Resource { return &resourceServerResource{} }

type resourceServerResource struct {
	client *client.Client
}

// resourceServerScopeModel is one {value, description} nested object.
type resourceServerScopeModel struct {
	Value       types.String `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
}

type resourceServerModel struct {
	ID              types.String               `tfsdk:"id"`
	Identifier      types.String               `tfsdk:"identifier"`
	Name            types.String               `tfsdk:"name"`
	Scopes          []resourceServerScopeModel `tfsdk:"scopes"`
	TokenTTLSeconds types.Int64                `tfsdk:"token_ttl_seconds"`
	SigningAlg      types.String               `tfsdk:"signing_alg"`
	CreatedAt       types.Int64                `tfsdk:"created_at"`
	UpdatedAt       types.Int64                `tfsdk:"updated_at"`
}

func (r *resourceServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_server"
}

func (r *resourceServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *resourceServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An API (resource server / audience) that accepts machine tokens from the OAuth2 " +
			"client_credentials grant. Defines an audience identifier and the scopes it grants.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Atlas object id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The audience (aud) stamped into issued tokens. Immutable — changing it forces replacement.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Human-readable name.",
				Required:            true,
			},
			"scopes": schema.ListNestedAttribute{
				MarkdownDescription: "Scopes this API defines.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							MarkdownDescription: "The scope string (e.g. read:widgets).",
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Optional human-readable description.",
							Optional:            true,
						},
					},
				},
			},
			"token_ttl_seconds": schema.Int64Attribute{
				MarkdownDescription: "Lifetime of machine tokens for this API, in seconds.",
				Optional:            true,
				Computed:            true,
			},
			"signing_alg": schema.StringAttribute{
				MarkdownDescription: "JWT signing algorithm for issued tokens (e.g. RS256).",
				Optional:            true,
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Creation time (epoch milliseconds).",
				Computed:            true,
			},
			"updated_at": schema.Int64Attribute{
				MarkdownDescription: "Last update time (epoch milliseconds).",
				Computed:            true,
			},
		},
	}
}

func (r *resourceServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := client.ResourceServerWrite{
		Identifier: client.Ptr(plan.Identifier.ValueString()),
		Name:       client.Ptr(plan.Name.ValueString()),
		Scopes:     scopesToAPI(plan.Scopes),
	}
	if !plan.TokenTTLSeconds.IsNull() && !plan.TokenTTLSeconds.IsUnknown() {
		body.TokenTTLSeconds = client.Ptr(plan.TokenTTLSeconds.ValueInt64())
	}
	if s := optionalString(plan.SigningAlg); s != nil {
		body.SigningAlg = s
	}

	created, err := r.client.CreateResourceServer(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create resource server", err.Error())
		return
	}
	r.mapToState(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetResourceServer(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read resource server", err.Error())
		return
	}
	r.mapToState(fetched, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.ResourceServerWrite{
		Name:   client.Ptr(plan.Name.ValueString()),
		Scopes: scopesToAPI(plan.Scopes),
	}
	if !plan.TokenTTLSeconds.IsNull() && !plan.TokenTTLSeconds.IsUnknown() {
		body.TokenTTLSeconds = client.Ptr(plan.TokenTTLSeconds.ValueInt64())
	}
	if s := optionalString(plan.SigningAlg); s != nil {
		body.SigningAlg = s
	}
	updated, err := r.client.UpdateResourceServer(ctx, plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update resource server", err.Error())
		return
	}
	r.mapToState(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteResourceServer(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete resource server", err.Error())
	}
}

func (r *resourceServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceServerResource) mapToState(rs *client.ResourceServer, m *resourceServerModel) {
	m.ID = types.StringValue(rs.ID)
	m.Identifier = types.StringValue(rs.Identifier)
	m.Name = types.StringValue(rs.Name)
	m.Scopes = scopesFromAPI(rs.Scopes)
	m.TokenTTLSeconds = types.Int64Value(rs.TokenTTLSeconds)
	m.SigningAlg = types.StringValue(rs.SigningAlg)
	m.CreatedAt = types.Int64Value(rs.CreatedAt)
	m.UpdatedAt = types.Int64Value(rs.UpdatedAt)
}

func scopesToAPI(in []resourceServerScopeModel) []client.ResourceServerScope {
	if in == nil {
		return nil
	}
	out := make([]client.ResourceServerScope, 0, len(in))
	for _, s := range in {
		scope := client.ResourceServerScope{Value: s.Value.ValueString()}
		if !s.Description.IsNull() && !s.Description.IsUnknown() {
			scope.Description = client.Ptr(s.Description.ValueString())
		}
		out = append(out, scope)
	}
	return out
}

func scopesFromAPI(in []client.ResourceServerScope) []resourceServerScopeModel {
	out := make([]resourceServerScopeModel, 0, len(in))
	for _, s := range in {
		out = append(out, resourceServerScopeModel{
			Value:       types.StringValue(s.Value),
			Description: stringPtrToValue(s.Description),
		})
	}
	return out
}
