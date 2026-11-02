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
  { t: 'mcp_gateway', item: '/v1/mcp_gateways/{id}', ds: true, patchAll: true, echo: true, secrets: ['upstream_authorization'], extra: [
    { api: 'has_upstream_authorization', kind: 'bool' }, { api: 'resource' }, { api: 'resource_metadata_url' }, { api: 'url' }, { api: 'created_at', kind: 'int' },
  ] },
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
  { t: 'iga_entitlement', item: '/v1/iga/entitlements/{id}', ds: true },
  { t: 'iga_campaign', item: '/v1/iga/campaigns/{id}', update: '' },
  { t: 'iga_sod_rule', item: '/v1/iga/sod_rules/{id}' },
  { t: 'iga_birthright_rule', item: '/v1/iga/birthright_rules/{id}' },
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
  { t: 'vc_credential_type', item: '/v1/vc/credential_types/{id}', update: '' },

  // ── D-24: Backend API collections without a resource ────────────────────────
  { t: 'group', item: '/v1/groups/{id}', ds: true },
  { t: 'group_rule', item: '/v1/group_rules/{id}' },
  { t: 'machine', item: '/v1/machines/{id}', ds: true },
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

  // ── Data sources ─────────────────────────────────────────────────────────────
  { t: 'user', kind: 'dataonly', item: '/v1/users/{id}', desc: 'A user in the instance, looked up by id.' },
  { t: 'organization', kind: 'dataonly', item: '/v1/organizations/{id}', desc: 'An organization in the instance, looked up by id.' },
  { t: 'sso_connection', kind: 'dataonly', item: '/v1/sso_connections/{id}', desc: 'An enterprise SSO connection, looked up by id.' },
];
