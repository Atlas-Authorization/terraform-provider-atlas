// The hand-written part of the generated resources: which Backend API object
// each one manages and how it is addressed. Everything else (fields, types,
// requiredness, descriptions) is read from the OpenAPI document.
//
//   t        Terraform type name without the atlas_ prefix
//   kind     coll (default: create in a collection, id from the response) |
//            single (one object at a fixed path) | keyed (addressed by path keys) |
//            dataonly (a data source only)
//   item     the OpenAPI object path; for coll the last {var} is the object id
//   params   path variable -> attribute name, for the parent / key variables
//   ds       also publish a data source
//   others   see scripts/gen-resources.mjs
export const manifest = [
  // ── Re-audit track resources ────────────────────────────────────────────────
  // Track 5: agents, MCP gateways, actions, triggers
  // mcp_gateway: hand-written (internal/provider/resource_*.go); generated as a data source only.
  { t: 'mcp_gateway', kind: 'dataonly', item: '/v1/mcp_gateways/{id}', desc: 'An MCP tool gateway, looked up by id.' },
  { t: 'mcp_gateway_upstream', item: '/v1/mcp_gateways/{id}/upstreams/{upstreamId}', params: { id: 'mcp_gateway_id' } },
  { t: 'mcp_gateway_pack', item: '/v1/mcp_gateways/{id}/packs/{packId}', params: { id: 'mcp_gateway_id' } },
  { t: 'mcp_gateway_budget', item: '/v1/mcp_gateways/{id}/budgets/{budgetId}', params: { id: 'mcp_gateway_id' } },
  { t: 'mcp_registry_entry', item: '/v1/mcp_registry/{id}', update: '' },
  { t: 'connection_action', kind: 'keyed', item: '/v1/connection_actions/{provider}/{name}', params: { provider: 'provider_name', name: 'action_name' }, skip: ['input_schema', 'method', 'source'], rename: {} },
  { t: 'connection_trigger', item: '/v1/connections/{id}/triggers/{triggerId}', params: { id: 'connection_id' } },
  { t: 'agent_lifecycle', kind: 'keyed', item: '/v1/agents/{id}/lifecycle', params: { id: 'agent_id' }, noDelete: true },
  { t: 'telemetry_export', kind: 'single', item: '/v1/telemetry_export', noDelete: true },
  // Track 3: conditional access, device assurance, risk
  { t: 'access_policy', item: '/v1/access_policies/{id}', ds: true },
  { t: 'device_assurance_policy', item: '/v1/device_assurance_policies/{id}', ds: true },
  { t: 'device_integration', item: '/v1/device_integrations/{id}' },
  { t: 'authenticator_policy', item: '/v1/authenticator_policies/{id}' },
  { t: 'email_risk_settings', kind: 'single', item: '/v1/fraud/email_risk/settings', noDelete: true },
  { t: 'email_suppression', item: '/v1/messaging/suppressions/{email}', update: '', pick: 'data0', idField: 'email', skip: ['emails'], required: ['email'] },
  // Track 4: workflows, governance, delegated admin, chat approvals
  { t: 'workflow', item: '/v1/workflows/{id}', ds: true },
  { t: 'chat_integration', item: '/v1/chat_integrations/{id}' },
  { t: 'iga_bundle', item: '/v1/iga/bundles/{id}' },
  { t: 'iga_eligibility', item: '/v1/iga/eligibilities/{id}', update: '' },
  // iga_entitlement: hand-written (internal/provider/resource_*.go); generated as a data source only.
  { t: 'iga_entitlement', kind: 'dataonly', item: '/v1/iga/entitlements/{id}', desc: 'An IGA entitlement, looked up by id.' },
  { t: 'iga_campaign', item: '/v1/iga/campaigns/{id}', update: '' },
  // iga_sod_rule: hand-written (internal/provider/resource_*.go).
  // iga_birthright_rule: hand-written (internal/provider/resource_*.go).
  { t: 'administrative_unit', item: '/v1/administrative_units/{id}' },
  { t: 'organization_application', kind: 'keyed', item: '/v1/organizations/{id}/applications/{clientId}', params: { id: 'organization_id', clientId: 'client_id' }, list: '/v1/organizations/{id}/applications', match: 'client_id' },
  { t: 'oauth_client_organization_access', kind: 'keyed', item: '/v1/oauth_clients/{id}/organization_access', params: { id: 'client_id' }, noRead: true, noDelete: true },
  { t: 'organization_connection', item: '/v1/organization_connections/{id}', update: '' },
  // Track 6: authorization, billing, customer-managed keys
  { t: 'fga_assertions', kind: 'keyed', item: '/v1/openfga/stores/{store}/assertions/{model}', params: { store: 'store_id', model: 'model_id' }, noDelete: true },
  { t: 'fga_materialization', item: '/v1/fga/stores/{id}/materializations/{mid}', params: { id: 'store_id' }, update: '', extra: [
    { api: 'store_id' }, { api: 'status' }, { api: 'built_revision', kind: 'int' }, { api: 'built_at', kind: 'int' }, { api: 'valid_until', kind: 'int' },
    { api: 'users', kind: 'int' }, { api: 'rows', kind: 'int' }, { api: 'error' },
  ], skip: ['store_id'], noGet: true },
  { t: 'abac_token_mapping', kind: 'single', item: '/v1/abac/token_mapping', noDelete: true, echo: true },
  { t: 'billing_coupon', item: '/v1/billing/coupons/{id}' },
  { t: 'audit_retention', kind: 'single', item: '/v1/audit_retention', noDelete: true, extra: [
    { api: 'configured', kind: 'bool' }, { api: 'default_days', kind: 'int' }, { api: 'min_days', kind: 'int' }, { api: 'max_days', kind: 'int' },
  ], echo: true, required: ['retention_days'], writeOnly: ['confirm_deletion'], kinds: { retention_days: 'int' } },
  { t: 'data_key_provider', kind: 'single', item: '/v1/instance/data_key_provider', init: 'POST', noDelete: true, raw: 'status', echo: true,
    desc: 'Customer-managed encryption for the instance: enables it on first apply (the KMS key must already be set with the key provider) and sets the data-key cache window. Destroying the resource leaves encryption on: removing it crypto-shreds everything encrypted under the data keys, so it is never done by terraform destroy.' },
  // Track 1: SAML IdP, SCIM, provisioning, catalog
  { t: 'scim_provisioning_target', item: '/v1/scim_provisioning_targets/{id}' },
  { t: 'scim_resource_type', item: '/v1/scim_resource_types/{id}', update: '' },
  { t: 'app_catalog_preset', item: '/v1/app_catalog/custom_presets/{id}', update: '' },
  { t: 'cloud_access_target', item: '/v1/cloud_access/targets/{id}' },
  { t: 'wsfed_connection', item: '/v1/wsfed_connections/{id}' },
  // Track 2: OIDC / FAPI / OAuth
  { t: 'oauth_client_profile', kind: 'keyed', item: '/v1/oauth_client_profiles/{name}', params: { name: 'name' }, skip: ['builtin'], list: '/v1/oauth_client_profiles', match: 'name' },
  // vc_credential_type: hand-written (internal/provider/resource_*.go).

  // ── D-24: Backend API collections without a resource ────────────────────────
  { t: 'group', item: '/v1/groups/{id}', ds: true },
  { t: 'group_rule', item: '/v1/group_rules/{id}' },
  { t: 'machine', item: '/v1/machines/{id}', ds: true },
  { t: 'machine_grant', item: '/v1/machines/{id}/grants/{grantId}', params: { id: 'machine_id' }, update: '' },
  { t: 'organization_machine_membership', item: '/v1/organizations/{id}/machine_memberships/{machineId}', params: { id: 'organization_id' }, idField: 'machine_id' },
  { t: 'action', item: '/v1/actions/{id}', ds: true },
  { t: 'action_template', item: '/v1/action_templates/{id}' },
  { t: 'flow', item: '/v1/flows/{id}', ds: true },
  { t: 'feature_flag', item: '/v1/feature_flags/{key}', idField: 'key', ds: true },
  { t: 'network_zone', item: '/v1/network_zones/{id}', ds: true },
  { t: 'network_acl', item: '/v1/network_acls/{id}' },
  { t: 'abac_policy', item: '/v1/abac_policies/{id}', ds: true },
  { t: 'abac_template', item: '/v1/abac_templates/{id}' },
  { t: 'abac_policy_link', item: '/v1/abac_policy_links/{id}' },
  { t: 'abac_schema', kind: 'single', item: '/v1/abac_schema' },
  { t: 'resource_type', item: '/v1/resource_types/{key}', idField: 'key' },
  { t: 'org_role_grant', item: '/v1/org_role_grants/{id}' },
  { t: 'log_stream', item: '/v1/log_streams/{id}' },
  { t: 'custom_oauth_provider', item: '/v1/custom_oauth_providers/{id}', ds: true },
  { t: 'connector', item: '/v1/connectors/{id}' },
  { t: 'service_connector', item: '/v1/service_connectors/{id}' },
  { t: 'org_ldap_connection', item: '/v1/org_ldap_connections/{id}' },
  { t: 'ldap_server_client', item: '/v1/ldap_server_clients/{id}' },
  { t: 'radius_client', item: '/v1/radius_clients/{id}' },
  { t: 'directory_pull_source', item: '/v1/directory_pull_sources/{id}' },
  { t: 'relay_agent', item: '/v1/relay_agents/{id}' },
  { t: 'scim_token', item: '/v1/scim_tokens/{id}', revoke: '/revoke', noGet: true },
  { t: 'sso_onboarding_profile', item: '/v1/sso_onboarding_profiles/{id}', update: '' },
  { t: 'proxy_app', item: '/v1/proxy_apps/{id}' },
  { t: 'ssf_stream', item: '/v1/ssf/streams/{id}', echo: true },
  { t: 'ssf_receiver', item: '/v1/ssf/receivers/{id}', echo: true },
  { t: 'fga_store', item: '/v1/fga/stores/{id}', update: '', ds: true },
  { t: 'localization', item: '/v1/localizations/{id}' },
  { t: 'sms_template', item: '/v1/sms_templates/{name}', idField: 'name' },
  { t: 'app_role', item: '/v1/app_roles/{id}' },
  { t: 'audit_event_schema', item: '/v1/audit_event_schemas/{id}' },
  { t: 'analytics_alert_rule', item: '/v1/analytics/alert_rules/{id}' },
  { t: 'session_exchange_app', item: '/v1/session_exchange_apps/{id}' },
  { t: 'logout_notify_endpoint', item: '/v1/logout_notify_endpoints/{id}' },
  { t: 'lti_platform', item: '/v1/lti_platforms/{id}' },
  { t: 'oauth_client_token_exchange_profile', item: '/v1/oauth_clients/{id}/token_exchange_profiles/{profileId}', params: { id: 'client_id' } },
  { t: 'oauth_client_trusted_issuer', item: '/v1/oauth_clients/{id}/trusted_issuers/{issuerId}', params: { id: 'client_id' }, update: '' },

  // ── Surface wave: governance, privacy, keys, region, referrals ──────────────
  { t: 'erasure_connector', item: '/v1/erasure_connectors/{id}' },
  { t: 'oauth_publisher', item: '/v1/oauth_publishers/{id}' },
  { t: 'oauth_publisher_client', kind: 'keyed', item: '/v1/oauth_publishers/{id}/clients/{client_id}', params: { id: 'publisher_id', client_id: 'client_id' }, noRead: true, echo: true },
  { t: 'connection_access_rules', kind: 'keyed', item: '/v1/connections/{id}/access_rules', params: { id: 'connection_id' } },
  { t: 'signing_key_schedule', kind: 'single', item: '/v1/instance/signing_key_schedule' },
  { t: 'audit_chain', kind: 'single', item: '/v1/audit/chain', noDelete: true, desc: 'The tamper-evident audit hash chain: turns it on and sets the checkpoint interval. Destroying the resource leaves the chain on, because the chain cannot be switched off without breaking verification of the rows already chained.' },
  { t: 'policy_audience', kind: 'keyed', item: '/v1/policies/{id}/audience', params: { id: 'policy_id' }, noDelete: true },
  { t: 'policy_translation', kind: 'keyed', item: '/v1/policies/{id}/versions/{versionId}/translations/{locale}', params: { id: 'policy_id', versionId: 'version_id', locale: 'locale' }, list: '/v1/policies/{id}/versions/{versionId}/translations', match: 'locale' },
  { t: 'referral_reward_rules', kind: 'single', item: '/v1/referral_reward_rules', noDelete: true },
  { t: 'region', kind: 'single', item: '/v1/region', noDelete: true, echo: true, example: '  pin = true', desc: 'Data residency for the instance: the pinned serving region, the failover region and FAPI isolation. Destroying the resource leaves the setting in place.' },
  { t: 'config_snapshot', item: '/v1/instance/config_snapshots/{id}', update: '' },
  { t: 'billing_connect_account', item: '/v1/billing/connect/accounts/{id}', update: '', noDelete: true },
  { t: 'agent_delegation', item: '/v1/agent_delegations/{id}', update: '', revoke: '/revoke' },

  // ── Surface-gap batch: native apps, hooks, aliases, PCA ────────────────────
  { t: 'native_app', item: '/v1/native_apps/{id}', example: '  platform      = "ios"\n  bundle_id     = "com.example.app"\n  team_id       = "ABCDE12345"\n  callback_urls = ["com.example.app://auth/callback"]' },
  { t: 'native_app_publishable_key', kind: 'single', item: '/v1/native_apps/publishable_key', noDelete: true },
  { t: 'before_first_session_hook', kind: 'single', item: '/v1/instance/hooks/before_first_session', write: 'patch', secrets: ['signing_secret'] },
  { t: 'permission_alias', kind: 'keyed', item: '/v1/permission_aliases/{alias}', params: { alias: 'alias' }, list: '/v1/permission_aliases', match: 'alias' },
  { t: 'pca_approvals_settings', kind: 'single', item: '/v1/pca/approvals/settings', noDelete: true },
  // PCA action registry and approvers. Passkey ceremonies happen outside Terraform: an approver is adopted once enrolled,
  // a reviewer is created by its enrolment link (create: '/enrolments') and revoked on destroy.
  { t: 'pca_action_source', item: '/v1/pca/action-registry/sources/{id}', required: ['ns', 'from', 'url'], writeOnly: ['from', 'url'] },
  { t: 'pca_action_publisher_key', item: '/v1/pca/action-registry/publisher-keys/{kid}', idField: 'kid', update: '', noRead: true, echo: true },
  { t: 'pca_action_namespace_pin', kind: 'keyed', item: '/v1/pca/action-registry/namespaces/{ns}/pin', params: { ns: 'ns' }, noRead: true, echo: true },
  { t: 'pca_action_reviewer', item: '/v1/pca/action-registry/reviewers/{id}', create: '/enrolments', idField: 'reviewer_id', update: '', list: '/v1/pca/action-registry/reviewers', match: 'id' },
  { t: 'pca_approver', kind: 'keyed', item: '/v1/pca/approvers/{id}', params: { id: 'approver_id' }, noDelete: true },
  { t: 'pca_approver_member', kind: 'keyed', item: '/v1/pca/approvers/{id}/members/{memberId}', params: { id: 'approver_id', memberId: 'member_id' }, noRead: true, noDelete: true, echo: true, skip: ['display_name', 'status'] },
  { t: 'pca_principal', item: '/v1/pca/principals/{principalPub}', update: '', idField: 'principal_pub' },

  // ── Surface-gap batch 2: older config surfaces ─────────────────────────────
  { t: 'machine_trust_policy', item: '/v1/machine_trust_policies/{id}' },
  { t: 'mfa_campaign', item: '/v1/mfa_campaigns/{id}' },
  { t: 'org_role_set', item: '/v1/org_role_sets/{id}' },
  { t: 'org_template', item: '/v1/org_templates/{id}' },
  { t: 'rate_limit_policy', item: '/v1/rate_limit_policies/{id}' },
  { t: 'rate_limit_override', kind: 'keyed', item: '/v1/rate_limit_overrides/{subject_type}/{subject_id}', params: { subject_type: 'subject_type', subject_id: 'subject_id' } },
  { t: 'step_up_policy', item: '/v1/step_up/policies/{id}' },
  { t: 'suppression', item: '/v1/suppressions/{id}', update: '', pick: 'data0', required: ['address'], skip: ['addresses', 'data', 'has_more', 'next_cursor', 'created'] },
  { t: 'instance_write_budget', kind: 'single', item: '/v1/instance/write_budget' },
  { t: 'connection_reconnect_settings', kind: 'single', item: '/v1/connection_reconnect_settings', noDelete: true },
  { t: 'fga_sync', kind: 'single', item: '/v1/fga/sync', noDelete: true },
  { t: 'user_traits_schema', kind: 'single', item: '/v1/instance/user_traits_schema', noDelete: true },
  { t: 'message_receiver', kind: 'keyed', item: '/v1/message_receivers/{provider}', params: { provider: 'provider_name' }, list: '/v1/message_receivers', match: 'provider' },
  { t: 'messaging_settings', kind: 'single', item: '/v1/messaging/settings', noDelete: true },
  { t: 'messaging_country_route', item: '/v1/messaging/country_routes/{country}', write: 'put', idField: 'country', update: '' },
  { t: 'messaging_provider_fallback', item: '/v1/messaging_providers/{channel}/fallbacks/{id}', params: { channel: 'channel' } },
  { t: 'organization_role_policy', kind: 'keyed', item: '/v1/organizations/{id}/role_policy', params: { id: 'organization_id' }, noDelete: true },
  { t: 'organization_seats', kind: 'keyed', item: '/v1/organizations/{id}/seats', params: { id: 'organization_id' }, noDelete: true },
  { t: 'web_bot_auth', kind: 'single', item: '/v1/web_bot_auth', noDelete: true },
  { t: 'web_bot_auth_directory', item: '/v1/web_bot_auth/directories/{id}', update: '', echo: true },
  { t: 'user_legal_hold', kind: 'keyed', item: '/v1/users/{id}/legal_hold', params: { id: 'user_id' }, revoke: '/release' },
  { t: 'id_app_adoption_code', item: '/v1/platform/atlas_id_apps/{id}/adoption_codes/{codeId}', params: { id: 'atlas_id_app_id' }, update: '', kinds: { status: 'string' } },

  // messaging_push_provider: hand-written spec (the push body is a oneOf the generator cannot flatten).

  // ── Data sources ─────────────────────────────────────────────────────────────
  { t: 'user', kind: 'dataonly', item: '/v1/users/{id}', desc: 'A user in the instance, looked up by id.' },
  { t: 'organization', kind: 'dataonly', item: '/v1/organizations/{id}', desc: 'An organization in the instance, looked up by id.' },
  { t: 'sso_connection', kind: 'dataonly', item: '/v1/sso_connections/{id}', desc: 'An enterprise SSO connection, looked up by id.' },
];
