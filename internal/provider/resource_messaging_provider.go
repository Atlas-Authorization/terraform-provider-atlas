package provider

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// messagingFieldKind is how a provider setting is typed in Terraform. The API
// stores every setting as a string; ports and booleans are converted.
type messagingFieldKind int

const (
	kindString messagingFieldKind = iota
	kindInt
	kindBool
)

// messagingField is one per-transport provider setting. Secret settings are
// write-only on the API (never returned): they are sensitive here and kept
// from configuration/state across a read.
type messagingField struct {
	key         string
	description string
	required    bool
	secret      bool
	kind        messagingFieldKind
}

type messagingTransport struct {
	name   string
	label  string
	fields []messagingField
}

var smsFromField = messagingField{key: "from", description: "Sender number or ID for this provider. Optional; `from_address` takes precedence."}

// emailTransports / smsTransports mirror the server's per-provider config
// schema (the keys, which are required, and which are secret).
var emailTransports = []messagingTransport{
	{"resend", "Resend", []messagingField{{key: "api_key", required: true, secret: true, description: "Resend API key (re_…)."}}},
	{"ses", "Amazon SES", []messagingField{
		{key: "region", required: true, description: "AWS region, e.g. us-east-1."},
		{key: "access_key_id", required: true, secret: true, description: "IAM access key ID."},
		{key: "secret_access_key", required: true, secret: true, description: "IAM secret access key."},
		{key: "session_token", secret: true, description: "STS session token, only for temporary credentials."},
	}},
	{"smtp", "SMTP", []messagingField{
		{key: "host", required: true, description: "SMTP server hostname, e.g. smtp.example.com."},
		{key: "port", kind: kindInt, description: "SMTP port (1-65535). Default 587."},
		{key: "secure", kind: kindBool, description: "Use implicit TLS. Default: true on port 465, otherwise STARTTLS."},
		{key: "username", description: "SMTP username, if the server requires authentication."},
		{key: "password", secret: true, description: "SMTP password or app password."},
	}},
	{"brevo", "Brevo", []messagingField{{key: "api_key", required: true, secret: true, description: "Brevo API key (xkeysib-…)."}}},
	{"mailgun", "Mailgun", []messagingField{
		{key: "api_key", required: true, secret: true, description: "Mailgun private API key."},
		{key: "domain", required: true, description: "Your Mailgun sending domain, e.g. mg.example.com."},
		{key: "region", description: "Mailgun region: `us` (default) or `eu`."},
	}},
}

var smsTransports = []messagingTransport{
	{"twilio", "Twilio", []messagingField{
		{key: "account_sid", required: true, description: "Twilio Account SID (AC…)."},
		{key: "auth_token", required: true, secret: true, description: "Twilio auth token."},
		smsFromField,
		{key: "whatsapp_from", description: "WhatsApp-enabled sender number, for the WhatsApp channel."},
		{key: "voice_from", description: "Voice-capable number, for the voice channel."},
	}},
	{"vonage", "Vonage", []messagingField{
		{key: "api_key", required: true, description: "Vonage API key."},
		{key: "api_secret", required: true, secret: true, description: "Vonage API secret."},
		smsFromField,
	}},
	{"messagebird", "MessageBird (Bird)", []messagingField{
		{key: "access_key", required: true, secret: true, description: "MessageBird (Bird) access key."},
		smsFromField,
	}},
	{"plivo", "Plivo", []messagingField{
		{key: "auth_id", required: true, description: "Plivo Auth ID."},
		{key: "auth_token", required: true, secret: true, description: "Plivo auth token."},
		smsFromField,
	}},
	{"telnyx", "Telnyx", []messagingField{
		{key: "api_key", required: true, secret: true, description: "Telnyx API key."},
		smsFromField,
	}},
	{"sinch", "Sinch", []messagingField{
		{key: "service_plan_id", required: true, description: "Sinch service plan ID."},
		{key: "api_token", required: true, secret: true, description: "Sinch API token."},
		{key: "region", description: "Sinch API region, e.g. us or eu. Default us."},
		smsFromField,
	}},
	{"infobip", "Infobip", []messagingField{
		{key: "base_url", required: true, description: "Your Infobip base URL, e.g. https://xxxxx.api.infobip.com."},
		{key: "api_key", required: true, secret: true, description: "Infobip API key."},
		smsFromField,
	}},
	{"clicksend", "ClickSend", []messagingField{
		{key: "username", required: true, description: "ClickSend username (your login email)."},
		{key: "api_key", required: true, secret: true, description: "ClickSend API key."},
		smsFromField,
	}},
	{"bandwidth", "Bandwidth", []messagingField{
		{key: "account_id", required: true, description: "Bandwidth account ID."},
		{key: "application_id", required: true, description: "Bandwidth messaging application ID."},
		{key: "username", required: true, description: "Bandwidth API username."},
		{key: "password", required: true, secret: true, description: "Bandwidth API password."},
		smsFromField,
	}},
	{"aws_sns", "Amazon SNS", []messagingField{
		{key: "region", required: true, description: "AWS region, e.g. us-east-1."},
		{key: "access_key_id", required: true, description: "IAM access key ID."},
		{key: "secret_access_key", required: true, secret: true, description: "IAM secret access key."},
		{key: "session_token", secret: true, description: "STS session token, only for temporary credentials."},
		smsFromField,
	}},
}

var (
	_ resource.Resource                   = &messagingProviderResource{}
	_ resource.ResourceWithImportState    = &messagingProviderResource{}
	_ resource.ResourceWithConfigure      = &messagingProviderResource{}
	_ resource.ResourceWithValidateConfig = &messagingProviderResource{}
)

// NewEmailProviderResource is the factory for atlas_email_provider.
func NewEmailProviderResource() resource.Resource {
	return &messagingProviderResource{channel: "email", transports: emailTransports}
}

// NewSMSProviderResource is the factory for atlas_sms_provider.
func NewSMSProviderResource() resource.Resource {
	return &messagingProviderResource{channel: "sms", transports: smsTransports}
}

// messagingProviderResource manages the instance's own email or SMS provider
// (bring your own Resend / SES / SMTP / Twilio / … account). Exactly one
// transport block is set.
type messagingProviderResource struct {
	client     *client.Client
	channel    string
	transports []messagingTransport
}

func (r *messagingProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.channel + "_provider"
}

func (r *messagingProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (f messagingField) attrType() attr.Type {
	switch f.kind {
	case kindInt:
		return types.Int64Type
	case kindBool:
		return types.BoolType
	default:
		return types.StringType
	}
}

func (t messagingTransport) objectType() types.ObjectType {
	at := map[string]attr.Type{}
	for _, f := range t.fields {
		at[f.key] = f.attrType()
	}
	return types.ObjectType{AttrTypes: at}
}

func (r *messagingProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	names := make([]string, 0, len(r.transports))
	for _, t := range r.transports {
		names = append(names, "`"+t.name+"`")
	}
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Equals the channel (`" + r.channel + "`).",
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"from_address": schema.StringAttribute{
			MarkdownDescription: "The sender " + map[string]string{"email": "address, e.g. `Acme <no-reply@acme.com>`", "sms": "number or ID"}[r.channel] + ".",
			Optional:            true,
		},
		"transport": schema.StringAttribute{
			MarkdownDescription: "The configured transport (the block that is set).",
			Computed:            true,
		},
		"status": schema.StringAttribute{
			MarkdownDescription: "Provider status reported by Atlas.",
			Computed:            true,
		},
		"secrets_set": schema.SetAttribute{
			MarkdownDescription: "Which secret settings Atlas has stored (never their values).",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"updated_at": schema.Int64Attribute{
			MarkdownDescription: "Last update time (epoch ms).",
			Computed:            true,
		},
	}
	if r.channel == "email" {
		attrs["force_destroy"] = schema.BoolAttribute{
			MarkdownDescription: "Remove the provider on destroy even while email-dependent features (email code or " +
				"link sign-in, sign-up email verification, password reset) are on and nothing else can send email. " +
				"Those features then stop being offered until a provider is configured again. Without it, Atlas " +
				"refuses the removal (`409 EMAIL_PROVIDER_IN_USE`). Changing it never calls Atlas. Defaults to false.",
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		}
	}
	for _, t := range r.transports {
		nested := map[string]schema.Attribute{}
		for _, f := range t.fields {
			desc := f.description
			if f.secret {
				desc += " Write-only: never returned by the API; kept from your configuration."
			}
			switch f.kind {
			case kindInt:
				nested[f.key] = schema.Int64Attribute{MarkdownDescription: desc, Required: f.required, Optional: !f.required}
			case kindBool:
				nested[f.key] = schema.BoolAttribute{MarkdownDescription: desc, Required: f.required, Optional: !f.required}
			default:
				nested[f.key] = schema.StringAttribute{MarkdownDescription: desc, Required: f.required, Optional: !f.required, Sensitive: f.secret}
			}
		}
		attrs[t.name] = schema.SingleNestedAttribute{
			MarkdownDescription: t.label + " settings. Set exactly one transport block.",
			Optional:            true,
			Attributes:          nested,
		}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "The instance's own " + r.channel + " provider (bring your own account), so Atlas sends " +
			r.channel + " through it instead of the shared default. Set exactly one transport block (" +
			strings.Join(names, ", ") + "). Secret settings are sensitive and write-only: the API stores them " +
			"encrypted and never returns them. Destroying the resource returns the channel to the Atlas default " +
			"sender. Import with the channel name (`" + r.channel + "`).",
		Attributes: attrs,
	}
}

// ValidateConfig requires exactly one transport block.
func (r *messagingProviderResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var set []string
	for _, t := range r.transports {
		var obj types.Object
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(t.name), &obj)...)
		if !obj.IsNull() {
			set = append(set, t.name)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}
	if len(set) != 1 {
		resp.Diagnostics.AddError("Exactly one transport block is required",
			fmt.Sprintf("Set exactly one transport block for the %s provider; found %d (%s).", r.channel, len(set), strings.Join(set, ", ")))
	}
}

// selected returns the configured transport and its object value from a plan/state.
func (r *messagingProviderResource) selected(ctx context.Context, get func(context.Context, path.Path, interface{}) diag.Diagnostics, diags *diag.Diagnostics) (*messagingTransport, types.Object) {
	for i := range r.transports {
		var obj types.Object
		diags.Append(get(ctx, path.Root(r.transports[i].name), &obj)...)
		if !obj.IsNull() && !obj.IsUnknown() {
			return &r.transports[i], obj
		}
	}
	return nil, types.ObjectNull(nil)
}

// configFromObject flattens a transport block into the API's string config.
func configFromObject(t *messagingTransport, obj types.Object) map[string]string {
	out := map[string]string{}
	values := obj.Attributes()
	for _, f := range t.fields {
		v, ok := values[f.key]
		if !ok || v.IsNull() || v.IsUnknown() {
			continue
		}
		switch tv := v.(type) {
		case types.String:
			out[f.key] = tv.ValueString()
		case types.Int64:
			out[f.key] = strconv.FormatInt(tv.ValueInt64(), 10)
		case types.Bool:
			out[f.key] = strconv.FormatBool(tv.ValueBool())
		}
	}
	return out
}

// objectFromAPI builds the transport block from the API's non-secret config,
// carrying each secret over from prior (when prior is the same transport).
func objectFromAPI(t *messagingTransport, cfg map[string]string, prior types.Object, diags *diag.Diagnostics) types.Object {
	var priorVals map[string]attr.Value
	if !prior.IsNull() && !prior.IsUnknown() {
		priorVals = prior.Attributes()
	}
	vals := map[string]attr.Value{}
	for _, f := range t.fields {
		raw, has := cfg[f.key]
		switch {
		case f.secret:
			if pv, ok := priorVals[f.key]; ok {
				vals[f.key] = pv
			} else {
				vals[f.key] = types.StringNull()
			}
		case f.kind == kindInt:
			if n, err := strconv.ParseInt(raw, 10, 64); has && err == nil {
				vals[f.key] = types.Int64Value(n)
			} else {
				vals[f.key] = types.Int64Null()
			}
		case f.kind == kindBool:
			if b, err := strconv.ParseBool(raw); has && err == nil {
				vals[f.key] = types.BoolValue(b)
			} else {
				vals[f.key] = types.BoolNull()
			}
		default:
			if has && raw != "" {
				vals[f.key] = types.StringValue(raw)
			} else {
				vals[f.key] = types.StringNull()
			}
		}
	}
	obj, d := types.ObjectValue(t.objectType().AttrTypes, vals)
	diags.Append(d...)
	return obj
}

// writeState sets the computed attributes and every transport block on state.
func (r *messagingProviderResource) writeState(ctx context.Context, state *tfsdk.State, p *client.MessagingProvider, prior func(name string) types.Object, diags *diag.Diagnostics) {
	diags.Append(state.SetAttribute(ctx, path.Root("id"), r.channel)...)
	diags.Append(state.SetAttribute(ctx, path.Root("transport"), p.Transport)...)
	diags.Append(state.SetAttribute(ctx, path.Root("status"), p.Status)...)
	diags.Append(state.SetAttribute(ctx, path.Root("from_address"), stringPtrToValue(nonEmpty(p.FromAddress)))...)
	diags.Append(state.SetAttribute(ctx, path.Root("updated_at"), p.UpdatedAt)...)
	keys := []string{}
	for k, set := range p.SecretsSet {
		if set {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	diags.Append(state.SetAttribute(ctx, path.Root("secrets_set"), stringSliceToSet(ctx, keys, diags))...)
	for i := range r.transports {
		t := &r.transports[i]
		if t.name == p.Transport {
			diags.Append(state.SetAttribute(ctx, path.Root(t.name), objectFromAPI(t, p.Config, prior(t.name), diags))...)
		} else {
			diags.Append(state.SetAttribute(ctx, path.Root(t.name), types.ObjectNull(t.objectType().AttrTypes))...)
		}
	}
}

func (r *messagingProviderResource) save(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diags *diag.Diagnostics) {
	t, obj := r.selected(ctx, plan.GetAttribute, diags)
	if diags.HasError() {
		return
	}
	if t == nil {
		diags.AddError("No transport configured", "Set exactly one transport block.")
		return
	}
	var from types.String
	diags.Append(plan.GetAttribute(ctx, path.Root("from_address"), &from)...)
	saved, err := r.client.PutMessagingProvider(ctx, r.channel, client.MessagingProviderWrite{
		Transport:   t.name,
		FromAddress: optionalString(from),
		Config:      configFromObject(t, obj),
	})
	if err != nil {
		diags.AddError("Unable to save "+r.channel+" provider", err.Error())
		return
	}
	state.Raw = plan.Raw
	r.writeState(ctx, state, saved, func(name string) types.Object {
		var o types.Object
		diags.Append(plan.GetAttribute(ctx, path.Root(name), &o)...)
		return o
	}, diags)
	// The API reports from_address as stored; keep the configured form when unset.
	if from.IsNull() {
		diags.Append(state.SetAttribute(ctx, path.Root("from_address"), types.StringNull())...)
	}
}

func (r *messagingProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.save(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

func (r *messagingProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.onlyForceDestroyChanged(ctx, req) {
		// force_destroy is provider-side only: record it, keep everything else.
		var force types.Bool
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("force_destroy"), &force)...)
		resp.State.Raw = req.State.Raw
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("force_destroy"), force)...)
		return
	}
	r.save(ctx, req.Plan, &resp.State, &resp.Diagnostics)
}

// onlyForceDestroyChanged reports whether an update changes force_destroy and
// nothing else. The configuration is compared with state (the computed
// attributes are unknown in the plan).
func (r *messagingProviderResource) onlyForceDestroyChanged(ctx context.Context, req resource.UpdateRequest) bool {
	if r.channel != "email" {
		return false
	}
	var diags diag.Diagnostics
	var planned, prior types.Bool
	diags.Append(req.Plan.GetAttribute(ctx, path.Root("force_destroy"), &planned)...)
	diags.Append(req.State.GetAttribute(ctx, path.Root("force_destroy"), &prior)...)
	if diags.HasError() || planned.Equal(prior) {
		return false
	}
	names := []string{"from_address"}
	for _, t := range r.transports {
		names = append(names, t.name)
	}
	for _, name := range names {
		var configured, stored attr.Value
		diags.Append(req.Config.GetAttribute(ctx, path.Root(name), &configured)...)
		diags.Append(req.State.GetAttribute(ctx, path.Root(name), &stored)...)
		if diags.HasError() || !configured.Equal(stored) {
			return false
		}
	}
	return true
}

func (r *messagingProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	p, err := r.client.GetMessagingProvider(ctx, r.channel)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read "+r.channel+" provider", err.Error())
		return
	}
	prior := req.State
	r.writeState(ctx, &resp.State, p, func(name string) types.Object {
		var o types.Object
		resp.Diagnostics.Append(prior.GetAttribute(ctx, path.Root(name), &o)...)
		return o
	}, &resp.Diagnostics)
}

func (r *messagingProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	force := false
	if r.channel == "email" {
		var f types.Bool
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("force_destroy"), &f)...)
		force = f.ValueBool()
	}
	err := r.client.DeleteMessagingProvider(ctx, r.channel, force)
	if err == nil || client.IsNotFound(err) {
		return
	}
	if apiErr, ok := err.(*client.APIError); ok && apiErr.HasCode("EMAIL_PROVIDER_IN_USE") {
		resp.Diagnostics.AddError("The email provider is still in use", emailProviderInUseDetail(apiErr))
		return
	}
	resp.Diagnostics.AddError("Unable to remove "+r.channel+" provider", err.Error())
}

// emailProviderInUseDetail turns a 409 EMAIL_PROVIDER_IN_USE into the steps
// that get the destroy through, naming the features from meta.features.
func emailProviderInUseDetail(apiErr *client.APIError) string {
	var features []string
	for _, item := range apiErr.Errors {
		if item.Code != "EMAIL_PROVIDER_IN_USE" {
			continue
		}
		if list, ok := item.Meta["features"].([]interface{}); ok {
			for _, f := range list {
				if name, ok := f.(string); ok && name != "" {
					features = append(features, name)
				}
			}
		}
	}
	which := "Email-dependent features are on"
	if len(features) > 0 {
		which = "These features are on and need it: " + strings.Join(features, ", ")
	}
	return "Atlas refused to remove the email provider: it is the only way this instance can send email. " + which + ".\n\n" +
		"To destroy it, either:\n" +
		"  - turn those features off first (the email code / email link strategies, sign-up email verification and " +
		"password reset in the instance auth configuration, e.g. atlas_instance_config.auth_config), apply, then destroy; or\n" +
		"  - set force_destroy = true on this atlas_email_provider, apply, then destroy. Those features stop being " +
		"offered until an email provider is configured again."
}

func (r *messagingProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if r.channel == "email" {
		// Imported providers start with the default, so the next plan is clean.
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("force_destroy"), false)...)
	}
}
