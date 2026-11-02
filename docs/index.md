# Atlas Provider

Manage an [Atlas](https://atlasauth.net) authentication instance as declarative
infrastructure-as-code. The provider wraps Atlas's secret-key **Backend API**
(BAPI) — the instance-scoped `/v1/*` endpoints an `sk_` key can call — so your
OAuth clients, SSO connections, resource servers, JWT templates, webhook
endpoints, roles, organizations (and their policies), custom domains, instance
settings, redirect-URL allowlist and billing plans live in version control and
roll out through the same plan/apply flow as the rest of your infrastructure.

Terraform manages **declarative configuration**, not per-user runtime data.
There are deliberately no `user` or `session` resources — those belong to the
application and the Backend API, not to a plan/apply lifecycle.

> **Official provider.** This is the first-party Atlas Terraform provider,
> published and maintained by Atlas. It talks only to the documented public
> Backend API and authenticates with your instance secret key (`sk_...`).

## Example Usage

```hcl
terraform {
  required_providers {
    atlas = {
      source  = "Atlas-Authorization/atlas"
      version = "~> 0.1"
    }
  }
}

provider "atlas" {
  # api_url    = "https://api.atlasauth.net"  # or ATLAS_API_URL (this is the default)
  # secret_key = "sk_live_..."            # or ATLAS_SECRET_KEY (preferred)
}
```

## Authentication

Every request is sent as `Authorization: Bearer <secret_key>` and is scoped to
the single Atlas instance that key belongs to. Provide credentials via the
environment (preferred, keeps the key out of state and VCS):

```sh
export ATLAS_SECRET_KEY="sk_live_..."
export ATLAS_API_URL="https://api.atlasauth.net"   # optional; this is the default
```

or inline in the provider block. Inline values take precedence over the
environment.

## Schema

### Optional

- `api_url` (String) — Base URL of the Atlas Backend API. Defaults to
  `https://api.atlasauth.net`, or `ATLAS_API_URL` when set.
- `secret_key` (String, Sensitive) — Atlas instance secret key (`sk_...`).
  Prefer `ATLAS_SECRET_KEY` so the key never lands in state or config.
- `platform_key` (String, Sensitive) — Atlas platform key (`psk_...`), needed
  only by platform resources such as `atlas_id_app`. Prefer
  `ATLAS_PLATFORM_KEY`. A configuration that manages only platform resources
  may omit `secret_key`.

## Resources

| Resource | Description |
|----------|-------------|
| [`atlas_oauth_client`](resources/oauth_client.md) | "Sign in with Atlas" relying party (OAuth/OIDC client). |
| [`atlas_oauth_provider`](resources/oauth_provider.md) | Social sign-in provider config (Google, GitHub, ...). |
| [`atlas_sso_connection`](resources/sso_connection.md) | Enterprise SSO connection (OIDC / SAML / Discourse). |
| [`atlas_resource_server`](resources/resource_server.md) | Machine-token API (audience + scopes). |
| [`atlas_jwt_template`](resources/jwt_template.md) | Named custom-claims map injected into tokens. |
| [`atlas_webhook_endpoint`](resources/webhook_endpoint.md) | Signed-event delivery endpoint. |
| [`atlas_role`](resources/role.md) | Custom organization role and its permissions. |
| [`atlas_organization`](resources/organization.md) | Organization (tenant) profile. |
| [`atlas_organization_policy`](resources/organization_policy.md) | Per-org security policy (MFA, SSO, sign-in methods, IP allow-list, session age). |
| [`atlas_domain`](resources/domain.md) | Custom instance domain (`fapi` / `accounts`). |
| [`atlas_instance_config`](resources/instance_config.md) | Singleton instance config: allowed origins + auth config. |
| [`atlas_redirect_url`](resources/redirect_url.md) | Allowlisted OAuth/SSO redirect URL. |
| [`atlas_billing_plan`](resources/billing_plan.md) | End-user billing plan (features, Stripe price, pricing model). |
| [`atlas_org_settings_schema`](resources/org_settings_schema.md) | JSON Schema for organization settings. |
| [`atlas_notification_template`](resources/notification_template.md) | Tenant-defined notification email template. |
| [`atlas_notification_category`](resources/notification_category.md) | Tenant-defined notification category (preference toggle). |
| [`atlas_email_template`](resources/email_template.md) | Override of a built-in transactional email template. |
| [`atlas_entitlement_override`](resources/entitlement_override.md) | Sales-led plan / features / limits for a user or organization, without Stripe. |
| [`atlas_instance_design`](resources/instance_design.md) | Hosted sign-in page design for the instance. |
| [`atlas_organization_design`](resources/organization_design.md) | An organization's design override. |
| [`atlas_cookie_consent`](resources/cookie_consent.md) | Cookie-consent banner (any CMP) on the hosted pages. |
| [`atlas_email_provider`](resources/email_provider.md) | Your own email provider (Resend, SES, SMTP, Brevo, Mailgun). |
| [`atlas_sms_provider`](resources/sms_provider.md) | Your own SMS provider (Twilio, Vonage, …). |
| [`atlas_id_app`](resources/id_app.md) | Sign in with Atlas app (needs `platform_key`). |
| [`atlas_better_auth_bridge`](resources/better_auth_bridge.md) | Better Auth session bridge: migration window, cookie prefix, live fallback (write-only secret). |
| [`atlas_policy`](resources/policy.md) | A policy users accept (terms, privacy): key, title, display, applies_to, reaccept, grace_days. |
| [`atlas_policy_source`](resources/policy_source.md) | External policy source: endpoint, sync mode, write-only signing secret. |
| [`atlas_policy_settings`](resources/policy_settings.md) | Policy settings: the `pol` session claim. |

## Data Sources

| Data source | Description |
|-------------|-------------|
| [`atlas_oauth_client`](data-sources/oauth_client.md) | Look up an OAuth client by id. |
| [`atlas_oauth_provider`](data-sources/oauth_provider.md) | Look up a social sign-in provider by catalog key. |
| [`atlas_resource_server`](data-sources/resource_server.md) | Look up a resource server by id. |
| [`atlas_instance_preflight`](data-sources/instance_preflight.md) | Setup check: everything Atlas can see is broken or worth knowing. |

<!-- generated:start -->

### More resources

| Resource | Description |
|----------|-------------|
| [`atlas_mcp_gateway`](resources/mcp_gateway.md) | Create an MCP gateway: one URL (`url` in the response) that fronts your MCP server. |
| [`atlas_mcp_gateway_upstream`](resources/mcp_gateway_upstream.md) | Add an upstream. |
| [`atlas_mcp_gateway_pack`](resources/mcp_gateway_pack.md) | Create a tool pack. |
| [`atlas_mcp_gateway_budget`](resources/mcp_gateway_budget.md) | Add a budget. |
| [`atlas_mcp_registry_entry`](resources/mcp_registry_entry.md) | Add a server to the registry as approved (default), or file it as pending for review. |
| [`atlas_connection_action`](resources/connection_action.md) | Define or replace a custom action: data only (an input schema, a request template, a response mapping), never code. |
| [`atlas_connection_trigger`](resources/connection_trigger.md) | Subscribe to provider events. |
| [`atlas_agent_lifecycle`](resources/agent_lifecycle.md) | Set an agent's sponsor, backup sponsor and expiry (omitted fields keep their value; null clears). |
| [`atlas_telemetry_export`](resources/telemetry_export.md) | Configure OTLP/HTTP export of gateway tool calls and relay calls as spans and metrics (atlas.calls, atlas.call.duration) to your collector. |
| [`atlas_access_policy`](resources/access_policy.md) | Create a conditional-access policy. |
| [`atlas_device_assurance_policy`](resources/device_assurance_policy.md) | Create a device assurance policy: per-platform minimum OS version (or "newest major minus N"), disk encryption, screen lock, secure hardware, jailbreak/root, firewall and EDR checks, and a grace period. |
| [`atlas_device_integration`](resources/device_integration.md) | Connect a device source that Atlas pulls on a schedule: Microsoft Intune (Graph, `tenantId` + app client credentials with DeviceManagementManagedDevices.Read.All), Jamf Pro (`baseUrl` + API client), CrowdStrike Falcon (`cloud` + API client with Zero Trust Assessment and Hosts read; `minZtaScore`, `userEmailDomain`), Fleet (`baseUrl` + API token) or Google Chrome / ChromeOS (service account with domain-wide delegation). |
| [`atlas_authenticator_policy`](resources/authenticator_policy.md) | Create an authenticator policy for a group: `required_factors` (members must pass a second factor, or enrol one), `grace_until` (members with no factor may still sign in until then), and a `passkey` profile (user verification, attestation, attachment, AAGUID allow / block lists) applied when members register passkeys. |
| [`atlas_email_risk_settings`](resources/email_risk_settings.md) | Change email risk settings. |
| [`atlas_email_suppression`](resources/email_suppression.md) | Add one or many addresses (a bounce or complaint webhook forwarder can post here). |
| [`atlas_workflow`](resources/workflow.md) | Create a lifecycle workflow. |
| [`atlas_chat_integration`](resources/chat_integration.md) | Connect a Slack workspace or a Teams bot using your own app. |
| [`atlas_iga_bundle`](resources/iga_bundle.md) | Create a bundle: several catalog entitlements, request form fields, and a policy (who may ask, guest expiry). |
| [`atlas_iga_eligibility`](resources/iga_eligibility.md) | Make a person eligible for an entitlement: they hold nothing until they activate it. |
| [`atlas_iga_entitlement`](resources/iga_entitlement.md) | Add an entitlement to the catalog: a role, organization membership, group, resource role, installed-app assignment or custom item, with an owner, a risk level and approval rules. |
| [`atlas_iga_campaign`](resources/iga_campaign.md) | Create a draft campaign scoped to everything, specific entitlements, or the entitlements for an app, role, group or organization, with reviewers (the person's manager, the entitlement owner, or named users), a due date and reminder days. |
| [`atlas_iga_sod_rule`](resources/iga_sod_rule.md) | Create a rule. |
| [`atlas_iga_birthright_rule`](resources/iga_birthright_rule.md) | Create a birthright rule: everyone matching `match` (a group, organization and role, app role or verified email domain; all conditions must hold) holds the listed entitlements. |
| [`atlas_administrative_unit`](resources/administrative_unit.md) | Create an administrative unit. |
| [`atlas_organization_application`](resources/organization_application.md) | Give an organization access to an application (by client id or id), optionally limited to some organization roles. |
| [`atlas_oauth_client_organization_access`](resources/oauth_client_organization_access.md) | Set who may use an application. |
| [`atlas_organization_connection`](resources/organization_connection.md) | Invite a partner organization to use an application the supplier organization holds. |
| [`atlas_fga_assertions`](resources/fga_assertions.md) | Replace the stored assertions for a model version. |
| [`atlas_fga_materialization`](resources/fga_materialization.md) | Precompute list-objects for every user of `user_type` on `object_type#relation`, so a huge list question is one indexed read. |
| [`atlas_abac_token_mapping`](resources/abac_token_mapping.md) | Set the claim mapping used by `authorize_token`: the principal type and id claim, claims that become parent entities (org, role, groups), claims copied as attributes, and trusted issuers with their JWKS URLs. |
| [`atlas_billing_coupon`](resources/billing_coupon.md) | Create a discount code. |
| [`atlas_audit_retention`](resources/audit_retention.md) | Set audit log retention in days (`null` returns to the default). |
| [`atlas_data_key_provider`](resources/data_key_provider.md) | Customer-managed encryption for the instance: enables it on first apply (the KMS key must already be set with the key provider) and sets the data-key cache window. |
| [`atlas_scim_provisioning_target`](resources/scim_provisioning_target.md) | Create an outbound SCIM provisioning target, optionally from a preset (preset + preset_params instead of base_url). |
| [`atlas_scim_resource_type`](resources/scim_resource_type.md) | Define a custom SCIM resource type: a name, an endpoint served at /scim/v2/<endpoint>, a schema URN and attribute definitions (type, multiValued, required, mutability). |
| [`atlas_app_catalog_preset`](resources/app_catalog_preset.md) | Save a private SAML preset for apps that are not in the catalog (an id, name, entityIdTemplate and acsUrlTemplate using {param} placeholders, params with a kind of slug, host, https_url or entity, and an attribute map). |
| [`atlas_cloud_access_target`](resources/cloud_access_target.md) | Add a cloud access target. |
| [`atlas_wsfed_connection`](resources/wsfed_connection.md) | Add a WS-Federation connection. |
| [`atlas_oauth_client_profile`](resources/oauth_client_profile.md) | Create or replace a custom client policy profile (rules: PKCE, PAR, sender constraint, signed request object, allowed auth methods and signing algorithms, https-only redirects, allowed claims). |
| [`atlas_vc_credential_type`](resources/vc_credential_type.md) | Define a credential type: its vct, and how each claim is filled from a user’s attributes. |
| [`atlas_group`](resources/group.md) | Create an instance group. |
| [`atlas_group_rule`](resources/group_rule.md) | Create a rule: when a user matches condition, run the actions (add to a group, grant an app role, set a trait). |
| [`atlas_machine`](resources/machine.md) | Create a named machine (M2M) identity; returns the machine secret (mk_) ONCE. |
| [`atlas_action`](resources/action.md) | Create an Action for a trigger (see the Actions guide for the 22 triggers: sign-in, registration, tokens, credentials, MFA, profile, organizations, billing). |
| [`atlas_action_template`](resources/action_template.md) | Create a gallery template (slug, name, author_name, summary, trigger, licence…), optionally with its first version (version, code, secret names). |
| [`atlas_flow`](resources/flow.md) | Create a sign-in journey (version 1, active unless activate is false). |
| [`atlas_feature_flag`](resources/feature_flag.md) | Create a feature flag: a typed value (`boolean`, `string`, `int` or `json`) with a `default_value`, an `off_value` (served while `enabled` is false; defaults to false for a boolean, else the default) and ordered targeting `rules` by org, user, plan and percentage. |
| [`atlas_network_zone`](resources/network_zone.md) | Create a network zone. |
| [`atlas_network_acl`](resources/network_acl.md) | Create a network ACL rule (allow/deny, a CIDR or single address or a `zone_id`, a priority). |
| [`atlas_abac_policy`](resources/abac_policy.md) | Create an ABAC policy from `source` written in the Cedar-subset policy language (permit/forbid with when/unless over principal, action, resource and context attributes; `fga.check(...)` consults the policy’s FGA store). |
| [`atlas_abac_template`](resources/abac_template.md) | Create a policy template: policy source whose scope uses the slots `?principal` and/or `?resource` (`principal == ?principal`, `resource in ?resource`, `resource is Doc in ?resource`). |
| [`atlas_abac_policy_link`](resources/abac_policy_link.md) | Link a template to entities. |
| [`atlas_abac_schema`](resources/abac_schema.md) | Set the ABAC entity schema: `entity_types` (attributes with `boolean`, `long`, `string`, `set`, `entity` and `record` types, `required: false` for optional ones, and `member_of`) and `actions` (`applies_to` principal and resource types and a `context`). |
| [`atlas_resource_type`](resources/resource_type.md) | Define a resource type: a `key`, its `roles` (each a set of permissions, optionally `includes` other roles), an optional `parent_type` resources nest under, and `inherit` (default true) controlling whether assignments on a parent flow down. |
| [`atlas_org_role_grant`](resources/org_role_grant.md) | Let a grantee organization hand out a bounded set of the grantor organization’s roles (role_keys) to its own members. |
| [`atlas_log_stream`](resources/log_stream.md) | Create a log stream. |
| [`atlas_custom_oauth_provider`](resources/custom_oauth_provider.md) | Create a custom OAuth 2.0 / OIDC provider (Atlas is the relying party). |
| [`atlas_connector`](resources/connector.md) | Create a community connector profile (SEO profile + OAuth config + share modes). |
| [`atlas_service_connector`](resources/service_connector.md) | Connect a service. |
| [`atlas_org_ldap_connection`](resources/org_ldap_connection.md) | Create an organization-owned LDAP / AD connection (url, bind_dn, base_dn, user_filter, attributes). |
| [`atlas_ldap_server_client`](resources/ldap_server_client.md) | Create a service account for the LDAPS interface: base_dn (dc=/o=/ou= parts, unique across Atlas), bind_cn, allowed_cidrs (empty = nobody may connect), allow_user_bind and size_limit. |
| [`atlas_radius_client`](resources/radius_client.md) | Register a RADIUS NAS client (name, NAS-Identifier, optional source-IP bind, write-only shared secret). |
| [`atlas_directory_pull_source`](resources/directory_pull_source.md) | Create a pull source for an organization: provider (entra, google_workspace, bamboohr, hibob, personio, workday, successfactors, gusto, csv, rest_json), config, write-only credentials and settings (schedule, role mappings, disable on removal). |
| [`atlas_relay_agent`](resources/relay_agent.md) | Register an outbound-only relay agent. |
| [`atlas_scim_token`](resources/scim_token.md) | Mint a SCIM provisioning token for an organization. |
| [`atlas_sso_onboarding_profile`](resources/sso_onboarding_profile.md) | Create a self-service SSO onboarding profile: allowed connection types and an optional bound organization. |
| [`atlas_proxy_app`](resources/proxy_app.md) | Create a proxy app: a hostname, optional upstream URL, ordered path/method rules requiring a role, permission or entitlement, public paths, and header/token injection. |
| [`atlas_ssf_stream`](resources/ssf_stream.md) | Create a stream. |
| [`atlas_ssf_receiver`](resources/ssf_receiver.md) | Register a transmitter. |
| [`atlas_fga_store`](resources/fga_store.md) | Create a fine-grained authorization store — an isolated namespace for authorization models and relationship tuples. |
| [`atlas_localization`](resources/localization.md) | Create or replace a localization override for one resource + locale. |
| [`atlas_sms_template`](resources/sms_template.md) | Set an SMS body template. |
| [`atlas_app_role`](resources/app_role.md) | Create an app-level user role: a role not tied to any organization, granting permission keys. |
| [`atlas_audit_event_schema`](resources/audit_event_schema.md) | Register a schema for an action: the target types and actor types it allows and a JSON Schema its metadata must satisfy. |
| [`atlas_analytics_alert_rule`](resources/analytics_alert_rule.md) | Create an alert rule: tell me when a sign-in health metric, or a key or organization's API usage against its baseline, crosses a threshold over a trailing window. |
| [`atlas_session_exchange_app`](resources/session_exchange_app.md) | Register a session-exchange app and its backchannel_logout_uri. |
| [`atlas_logout_notify_endpoint`](resources/logout_notify_endpoint.md) | Register a logout fan-out endpoint. |
| [`atlas_lti_platform`](resources/lti_platform.md) | Register an LTI 1.3 platform (issuer, client_id, auth_login_url, jwks_uri, deployment_ids). |
| [`atlas_oauth_client_token_exchange_profile`](resources/oauth_client_token_exchange_profile.md) | Register a custom subject_token_type (a urn: or https URI that is not a standard Atlas or IETF token type) for a confidential client. |
| [`atlas_oauth_client_trusted_issuer`](resources/oauth_client_trusted_issuer.md) | Trust an external issuer for this client. |

### More data sources

| Data source | Description |
|-------------|-------------|
| [`atlas_mcp_gateway`](data-sources/mcp_gateway.md) | Look up mcp gateway by id. |
| [`atlas_access_policy`](data-sources/access_policy.md) | Look up access policy by id. |
| [`atlas_device_assurance_policy`](data-sources/device_assurance_policy.md) | Look up device assurance policy by id. |
| [`atlas_workflow`](data-sources/workflow.md) | Look up workflow by id. |
| [`atlas_iga_entitlement`](data-sources/iga_entitlement.md) | Look up iga entitlement by id. |
| [`atlas_group`](data-sources/group.md) | Look up group by id. |
| [`atlas_machine`](data-sources/machine.md) | Look up machine by id. |
| [`atlas_action`](data-sources/action.md) | Look up action by id. |
| [`atlas_flow`](data-sources/flow.md) | Look up flow by id. |
| [`atlas_feature_flag`](data-sources/feature_flag.md) | Look up feature flag by id. |
| [`atlas_network_zone`](data-sources/network_zone.md) | Look up network zone by id. |
| [`atlas_abac_policy`](data-sources/abac_policy.md) | Look up abac policy by id. |
| [`atlas_custom_oauth_provider`](data-sources/custom_oauth_provider.md) | Look up custom oauth provider by id. |
| [`atlas_fga_store`](data-sources/fga_store.md) | Look up fga store by id. |
| [`atlas_user`](data-sources/user.md) | Look up user by id. |
| [`atlas_organization`](data-sources/organization.md) | Look up organization by id. |
| [`atlas_sso_connection`](data-sources/sso_connection.md) | Look up sso connection by id. |

<!-- generated:end -->
