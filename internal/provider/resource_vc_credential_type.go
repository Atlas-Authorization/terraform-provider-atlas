package provider

import (
	"context"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &vcCredentialTypeResource{}
	_ resource.ResourceWithImportState = &vcCredentialTypeResource{}
	_ resource.ResourceWithConfigure   = &vcCredentialTypeResource{}
)

// NewVcCredentialTypeResource is the factory registered with the provider.
func NewVcCredentialTypeResource() resource.Resource { return &vcCredentialTypeResource{} }

type vcCredentialTypeResource struct {
	client *client.Client
}

type vcClaimModel struct {
	Name        types.String `tfsdk:"name"`
	Source      types.String `tfsdk:"source"`
	Disclosable types.Bool   `tfsdk:"disclosable"`
	ValueType   types.String `tfsdk:"value_type"`
}

type vcCredentialTypeModel struct {
	ID                       types.String   `tfsdk:"id"`
	Vct                      types.String   `tfsdk:"vct"`
	Name                     types.String   `tfsdk:"name"`
	Format                   types.String   `tfsdk:"format"`
	Namespace                types.String   `tfsdk:"namespace"`
	IacaRoots                types.List     `tfsdk:"iaca_roots"`
	SignerChain              types.List     `tfsdk:"signer_chain"`
	Claims                   []vcClaimModel `tfsdk:"claims"`
	LifetimeSeconds          types.Int64    `tfsdk:"lifetime_seconds"`
	RequireWalletAttestation types.Bool     `tfsdk:"require_wallet_attestation"`
}

func (r *vcCredentialTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vc_credential_type"
}

func (r *vcCredentialTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *vcCredentialTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A verifiable credential type Atlas issues to wallets (OpenID4VCI) and verifies (OpenID4VP): " +
			"an SD-JWT VC or an ISO 18013-5 mdoc, with each claim filled from a user attribute. A credential type is " +
			"immutable: any change forces replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Atlas object id; also the credential configuration id in the issuer metadata.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"vct": schema.StringAttribute{
				MarkdownDescription: "The credential type identifier (a URN or URL); for `mso_mdoc` the docType, for example `org.iso.18013.5.1.mDL`.",
				Required:            true,
				PlanModifiers:       replace,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name (1 to 120 characters).",
				Required:            true,
				PlanModifiers:       replace,
			},
			"format": schema.StringAttribute{
				MarkdownDescription: "`dc+sd-jwt` (default) or `mso_mdoc`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("dc+sd-jwt"),
				PlanModifiers:       replace,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "`mso_mdoc` only: the namespace the claims live in. Defaults to the docType.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"iaca_roots": schema.ListAttribute{
				MarkdownDescription: "`mso_mdoc` only: PEM IACA root certificates (ISO 18013-5 Annex B). With roots set, a presentation " +
					"must carry an `x5chain` leading to one of them and the issuer signature must verify under its Document Signer key; " +
					"bare trusted keys are then not accepted.",
				ElementType:   types.StringType,
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace(), listplanmodifier.UseStateForUnknown()},
			},
			"signer_chain": schema.ListAttribute{
				MarkdownDescription: "`mso_mdoc` only: PEM Document Signer certificate first, then intermediates, published as `x5chain` in issued " +
					"mdocs. The first certificate must be for the instance's mdoc signing key (its ES256 key).",
				ElementType:   types.StringType,
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace(), listplanmodifier.UseStateForUnknown()},
			},
			"lifetime_seconds": schema.Int64Attribute{
				MarkdownDescription: "How long an issued credential is valid, in seconds (60 to 315360000). Defaults to one year.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace(), int64planmodifier.UseStateForUnknown()},
			},
			"require_wallet_attestation": schema.BoolAttribute{
				MarkdownDescription: "Whether the wallet must present an attestation. Set it with the Backend API; read-only here.",
				Computed:            true,
			},
			"claims": schema.ListNestedAttribute{
				MarkdownDescription: "The claims of the credential, in order.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The claim name (or mdoc element identifier).",
							Required:            true,
						},
						"source": schema.StringAttribute{
							MarkdownDescription: "`user_id`, `email`, `first_name`, `last_name`, `username` or `public_metadata.<key>`.",
							Required:            true,
						},
						"disclosable": schema.BoolAttribute{
							MarkdownDescription: "Selectively disclosable (SD-JWT VC). Defaults to true.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
						},
						"value_type": schema.StringAttribute{
							MarkdownDescription: "`string` (default), `number`, `boolean` or `full-date` (`YYYY-MM-DD`).",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *vcCredentialTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vcCredentialTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := client.VcCredentialTypeWrite{
		Vct:    plan.Vct.ValueString(),
		Name:   plan.Name.ValueString(),
		Format: plan.Format.ValueString(),
	}
	if !plan.Namespace.IsNull() && !plan.Namespace.IsUnknown() {
		body.Namespace = client.Ptr(plan.Namespace.ValueString())
	}
	if !plan.LifetimeSeconds.IsNull() && !plan.LifetimeSeconds.IsUnknown() {
		body.LifetimeSeconds = client.Ptr(plan.LifetimeSeconds.ValueInt64())
	}
	body.IacaRoots = listToStringSlice(ctx, plan.IacaRoots, &resp.Diagnostics)
	body.SignerChain = listToStringSlice(ctx, plan.SignerChain, &resp.Diagnostics)
	for _, c := range plan.Claims {
		claim := client.VcClaim{Name: c.Name.ValueString(), Source: c.Source.ValueString(), ValueType: c.ValueType.ValueString()}
		if !c.Disclosable.IsNull() && !c.Disclosable.IsUnknown() {
			claim.Disclosable = client.Ptr(c.Disclosable.ValueBool())
		}
		body.Claims = append(body.Claims, claim)
	}
	created, err := r.client.CreateVcCredentialType(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create credential type", err.Error())
		return
	}
	mapVcCredentialType(ctx, created, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vcCredentialTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vcCredentialTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetVcCredentialType(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read credential type", err.Error())
		return
	}
	mapVcCredentialType(ctx, fetched, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is never reached: every attribute forces replacement.
func (r *vcCredentialTypeResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Credential types are immutable", "Change the credential type by replacing it.")
}

func (r *vcCredentialTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vcCredentialTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVcCredentialType(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Unable to delete credential type", err.Error())
	}
}

func (r *vcCredentialTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func mapVcCredentialType(ctx context.Context, t *client.VcCredentialType, m *vcCredentialTypeModel, diags *diag.Diagnostics) {
	m.IacaRoots = stringSliceToList(ctx, nonNil(t.IacaRoots), diags)
	m.SignerChain = stringSliceToList(ctx, nonNil(t.SignerChain), diags)
	m.ID = types.StringValue(t.ID)
	m.Vct = types.StringValue(t.Vct)
	m.Name = types.StringValue(t.Name)
	m.Format = types.StringValue(t.Format)
	m.Namespace = stringPtrToValue(t.Namespace)
	m.LifetimeSeconds = types.Int64Value(t.LifetimeSeconds)
	m.RequireWalletAttestation = types.BoolValue(t.RequireWalletAttestation)
	m.Claims = make([]vcClaimModel, 0, len(t.Claims))
	for _, c := range t.Claims {
		cm := vcClaimModel{Name: types.StringValue(c.Name), Source: types.StringValue(c.Source), Disclosable: types.BoolValue(true), ValueType: types.StringNull()}
		if c.Disclosable != nil {
			cm.Disclosable = types.BoolValue(*c.Disclosable)
		}
		if c.ValueType != "" {
			cm.ValueType = types.StringValue(c.ValueType)
		}
		m.Claims = append(m.Claims, cm)
	}
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
