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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &instanceConfigResource{}
	_ resource.ResourceWithImportState = &instanceConfigResource{}
	_ resource.ResourceWithConfigure   = &instanceConfigResource{}
)

// NewInstanceConfigResource is the factory registered with the provider.
func NewInstanceConfigResource() resource.Resource { return &instanceConfigResource{} }

type instanceConfigResource struct {
	client *client.Client
}

type instanceConfigModel struct {
	ID                 types.String `tfsdk:"id"`
	AllowedOrigins     types.Set    `tfsdk:"allowed_origins"`
	AuthConfig         types.String `tfsdk:"auth_config"`
	AuthConfigResolved types.String `tfsdk:"auth_config_resolved"`
	NativeApps         types.Object `tfsdk:"native_apps"`
	Environment        types.String `tfsdk:"environment"`
	PublishableKey     types.String `tfsdk:"publishable_key"`
	FrontendAPIHost    types.String `tfsdk:"frontend_api_host"`
	CreatedAt          types.Int64  `tfsdk:"created_at"`
}

// nativeAppsModel is the typed `native_apps` block. It is folded into the
// `auth_config` patch under the camelCase `nativeApps` key before it is sent.
type nativeAppsModel struct {
	AppleAppIds types.List `tfsdk:"apple_app_ids"`
	AndroidApps types.List `tfsdk:"android_apps"`
}

// androidAppModel is one entry of `native_apps.android_apps`.
type androidAppModel struct {
	PackageName            types.String `tfsdk:"package_name"`
	Sha256CertFingerprints types.List   `tfsdk:"sha256_cert_fingerprints"`
}

func (r *instanceConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_config"
}

func (r *instanceConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *instanceConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The singleton configuration of the Atlas instance this secret key belongs to: its " +
			"`allowed_origins` CORS list and its `auth_config`. There is exactly one per instance — manage a single " +
			"`atlas_instance_config` resource. `auth_config` is a PARTIAL patch merged server-side (never a wholesale " +
			"replacement); state keeps the patch you wrote verbatim, while the full merged result is read back into the " +
			"computed `auth_config_resolved`. Native-app passkey association can be configured with the typed " +
			"`native_apps` block instead of raw `auth_config` JSON — it folds into the same `auth_config` patch under " +
			"`nativeApps`. Destroying this resource only stops Terraform managing the config; it does " +
			"not reset the instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The Atlas instance id (singleton key).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"allowed_origins": schema.SetAttribute{
				MarkdownDescription: "Exact origins (`scheme://host[:port]`, no wildcards, no path) permitted for the " +
					"widget / JS SDK and OAuth redirect validation. Native-app webview origins (`tauri://`, " +
					"`capacitor://`, `ionic://`) are accepted.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"auth_config": schema.StringAttribute{
				MarkdownDescription: "The instance auth configuration as a JSON object string (use `jsonencode(...)`). " +
					"A PARTIAL patch merged into the current config server-side. State holds exactly the patch you " +
					"wrote (round-tripped verbatim), NOT the server-merged object, so the plan is clean. The full " +
					"effective configuration is exposed separately as `auth_config_resolved`.",
				Optional: true,
				Computed: true,
			},
			"auth_config_resolved": schema.StringAttribute{
				MarkdownDescription: "The full effective configuration Atlas resolved from your `auth_config` plus " +
					"defaults (read-only).",
				Computed: true,
			},
			"native_apps": schema.SingleNestedAttribute{
				MarkdownDescription: "Typed native-app passkey association, folded into the `auth_config` patch under " +
					"`nativeApps` before it is sent (so you configure it with typed HCL instead of raw JSON). A value " +
					"here TAKES PRECEDENCE over any `nativeApps` embedded in the raw `auth_config` JSON. Atlas serves the " +
					"matching `/.well-known/apple-app-site-association` and `/.well-known/assetlinks.json` on the instance " +
					"Frontend API host (see `frontend_api_host`), and the WebAuthn ceremony accepts the derived native " +
					"app origins. Like `auth_config`, this is an Optional partial patch kept verbatim in state and is " +
					"never refreshed from the server; leave it unset to manage `nativeApps` via raw `auth_config` (or not " +
					"at all).",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"apple_app_ids": schema.ListAttribute{
						MarkdownDescription: "iOS/macOS app ids in `<TeamID>.<bundleId>` form (e.g. " +
							"`LB4397Q8XJ.com.acme.app`). Emitted as `appleAppIds` in the auth_config patch.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"android_apps": schema.ListNestedAttribute{
						MarkdownDescription: "Android apps allowed to assert passkeys. Emitted as `androidApps` in the " +
							"auth_config patch.",
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"package_name": schema.StringAttribute{
									MarkdownDescription: "The Android application id / package name (e.g. `com.acme.app`). " +
										"Emitted as `packageName`.",
									Required: true,
								},
								"sha256_cert_fingerprints": schema.ListAttribute{
									MarkdownDescription: "SHA-256 signing-certificate fingerprints (colon-separated hex, " +
										"from keytool or the Play signing key). Emitted as `sha256CertFingerprints`.",
									Required:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
				},
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The instance environment (e.g. `production`, `development`).",
				Computed:            true,
			},
			"publishable_key": schema.StringAttribute{
				MarkdownDescription: "The instance publishable key (`pk_...`).",
				Computed:            true,
			},
			// Read-only/informational: auto-assigned by Atlas and populated by
			// Read. It is not a diff source — never set from config.
			"frontend_api_host": schema.StringAttribute{
				MarkdownDescription: "The instance Frontend API host (read-only; auto-assigned by Atlas).",
				Computed:            true,
			},
			"created_at": schema.Int64Attribute{
				MarkdownDescription: "Instance creation time (epoch ms).",
				Computed:            true,
			},
		},
	}
}

func (r *instanceConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan instanceConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *instanceConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fetched, err := r.client.GetInstance(ctx)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read instance config", err.Error())
		return
	}
	r.mapToState(ctx, fetched, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *instanceConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan instanceConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op: an instance's config is a singleton that cannot be deleted.
// Removing the resource simply stops Terraform managing it; the live config is
// left exactly as it is.
func (r *instanceConfigResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *instanceConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The id is the instance id, but Read always fetches the singleton, so any
	// placeholder works; passthrough keeps the convention of the other resources.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// write sends the plan as a PATCH, re-reads the full projection, and maps it back.
func (r *instanceConfigResource) write(ctx context.Context, plan *instanceConfigModel, diags *diag.Diagnostics) {
	body := client.InstanceUpdate{}
	if !plan.AllowedOrigins.IsNull() && !plan.AllowedOrigins.IsUnknown() {
		origins := setToStringSlice(ctx, plan.AllowedOrigins, diags)
		if origins == nil {
			origins = []string{}
		}
		body.AllowedOrigins = origins
	}
	raw, err := stringToJSONRaw(plan.AuthConfig)
	if err != nil {
		diags.AddError("Invalid auth_config", "auth_config must be a JSON object string: "+err.Error())
		return
	}
	// When the typed `native_apps` block is set, fold it into the auth_config
	// patch under `nativeApps` (overwriting any nativeApps the raw JSON carried).
	// When it is null/unknown we inject nothing, leaving the raw patch as-is so a
	// user may still manage nativeApps via raw JSON or leave it unmanaged.
	if !plan.NativeApps.IsNull() && !plan.NativeApps.IsUnknown() {
		merged, err := mergeNativeAppsIntoAuthConfig(ctx, raw, plan.NativeApps, diags)
		if err != nil {
			diags.AddError("Invalid native_apps", "could not fold native_apps into auth_config: "+err.Error())
			return
		}
		if diags.HasError() {
			return
		}
		raw = merged
	}
	body.AuthConfig = raw

	if _, err := r.client.UpdateInstance(ctx, body); err != nil {
		diags.AddError("Unable to update instance config", err.Error())
		return
	}
	// The PATCH response omits the read-only fields (environment, keys, …); re-read
	// the full projection so every computed attribute is populated.
	fetched, err := r.client.GetInstance(ctx)
	if err != nil {
		diags.AddError("Unable to read instance config", err.Error())
		return
	}
	r.mapToState(ctx, fetched, plan, diags)
	// `auth_config` in state must equal what the user WROTE (the partial patch,
	// round-tripped verbatim), NOT the server-merged object mapToState leaves
	// alone — this keeps plan == apply and avoids a phantom diff. Only when the
	// user configured nothing (an unset Optional+Computed value, which plans as
	// unknown) do we surface the API's stored patch so the computed attribute
	// resolves to a concrete value.
	//
	// `native_apps` lives SOLELY in its own typed attribute. If the user manages
	// native_apps but left auth_config unset, we folded a `nativeApps` key into
	// the patch we sent, so the stored patch we now surface contains it — strip
	// it back out so it is not duplicated into auth_config state. (When the user
	// did not manage native_apps, the stored patch is surfaced byte-for-byte as
	// before.) native_apps itself is kept verbatim from the plan by the
	// resp.State.Set in Create/Update; mapToState never touches it.
	if plan.AuthConfig.IsNull() || plan.AuthConfig.IsUnknown() {
		stored := fetched.AuthConfig
		if !plan.NativeApps.IsNull() && !plan.NativeApps.IsUnknown() {
			stored = stripJSONObjectKey(stored, "nativeApps")
		}
		plan.AuthConfig = jsonRawToValue(stored)
	}
}

// mergeNativeAppsIntoAuthConfig decodes the typed native_apps object into the
// camelCase JSON the server expects and merges it into the auth_config patch
// under `nativeApps`. rawPatch is the user's raw auth_config (nil/empty when
// unset); the typed native_apps value overwrites any nativeApps rawPatch held.
// The returned bytes are what the resource sends as body.AuthConfig.
func mergeNativeAppsIntoAuthConfig(ctx context.Context, rawPatch json.RawMessage, obj types.Object, diags *diag.Diagnostics) (json.RawMessage, error) {
	native := nativeAppsToJSON(ctx, obj, diags)
	if diags.HasError() {
		return rawPatch, nil
	}
	return mergeNativeAppsPatch(rawPatch, native)
}

// mergeNativeAppsPatch is the pure fold: it parses rawPatch into an object (or
// starts from {}), sets the `nativeApps` key to the already-built camelCase
// value, and re-marshals. It preserves every other key in rawPatch and
// overrides any nativeApps rawPatch already had.
func mergeNativeAppsPatch(rawPatch json.RawMessage, native map[string]any) (json.RawMessage, error) {
	patch := map[string]any{}
	if len(rawPatch) > 0 && string(rawPatch) != "null" {
		if err := json.Unmarshal(rawPatch, &patch); err != nil {
			return nil, err
		}
	}
	patch["nativeApps"] = native
	out, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(out), nil
}

// nativeAppsToJSON decodes the typed native_apps object into the camelCase shape
// the server reads: {"appleAppIds":[...],"androidApps":[{"packageName":...,
// "sha256CertFingerprints":[...]}]}. Absent (null/unknown) sub-lists are omitted;
// an empty (but set) list is emitted as []. A null/unknown object yields nil.
func nativeAppsToJSON(ctx context.Context, obj types.Object, diags *diag.Diagnostics) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var na nativeAppsModel
	diags.Append(obj.As(ctx, &na, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	out := map[string]any{}
	if !na.AppleAppIds.IsNull() && !na.AppleAppIds.IsUnknown() {
		out["appleAppIds"] = nonNilStrings(listToStringSlice(ctx, na.AppleAppIds, diags))
	}
	if !na.AndroidApps.IsNull() && !na.AndroidApps.IsUnknown() {
		var apps []androidAppModel
		diags.Append(na.AndroidApps.ElementsAs(ctx, &apps, false)...)
		arr := make([]map[string]any, 0, len(apps))
		for _, a := range apps {
			arr = append(arr, map[string]any{
				"packageName":            a.PackageName.ValueString(),
				"sha256CertFingerprints": nonNilStrings(listToStringSlice(ctx, a.Sha256CertFingerprints, diags)),
			})
		}
		out["androidApps"] = arr
	}
	return out
}

// nonNilStrings ensures a slice marshals as [] rather than null.
func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// stripJSONObjectKey returns raw with the given top-level key removed. A raw that
// is not a JSON object, or that lacks the key, is returned unchanged. When the
// removal leaves an empty object, nil is returned so the caller surfaces null.
func stripJSONObjectKey(raw json.RawMessage, key string) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return raw
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return raw // not an object — leave as-is
	}
	if _, ok := m[key]; !ok {
		return raw
	}
	delete(m, key)
	if len(m) == 0 {
		return nil
	}
	out, err := json.Marshal(m)
	if err != nil {
		return raw
	}
	return json.RawMessage(out)
}

// mapToState refreshes the computed fields from the API projection. It
// deliberately does NOT touch `auth_config`: that attribute holds the user's
// last-applied partial patch verbatim (set by Create/Update, preserved by Read),
// NOT the server-merged object. The merged/effective view lives in the computed
// `auth_config_resolved`, which is refreshed here every Read/Create/Update.
func (r *instanceConfigResource) mapToState(ctx context.Context, in *client.Instance, m *instanceConfigModel, diags *diag.Diagnostics) {
	m.ID = types.StringValue(in.ID)
	m.AllowedOrigins = stringSliceToSet(ctx, in.AllowedOrigins, diags)
	m.AuthConfigResolved = jsonRawToValue(in.AuthConfigResolved)
	m.Environment = types.StringValue(in.Environment)
	m.PublishableKey = types.StringValue(in.PublishableKey)
	m.FrontendAPIHost = types.StringValue(in.FrontendAPIHost)
	m.CreatedAt = types.Int64Value(in.CreatedAt)
}
