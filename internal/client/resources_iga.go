package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// ───────────────────────────── Identity governance ─────────────────────────────

// IgaEntitlement mirrors /v1/iga/entitlements.
type IgaEntitlement struct {
	ID                   string          `json:"id"`
	Key                  string          `json:"key"`
	Name                 string          `json:"name"`
	Description          *string         `json:"description"`
	Kind                 string          `json:"kind"`
	Target               json.RawMessage `json:"target"`
	OwnerUserID          *string         `json:"owner_user_id"`
	RiskLevel            string          `json:"risk_level"`
	ApprovalRules        json.RawMessage `json:"approval_rules"`
	MaxDurationSeconds   *int64          `json:"max_duration_seconds"`
	Requestable          bool            `json:"requestable"`
	RequireJustification bool            `json:"require_justification"`
}

// IgaEntitlementWrite is the create / update payload (key and kind are create-only).
type IgaEntitlementWrite struct {
	Key                  string                 `json:"key,omitempty"`
	Kind                 string                 `json:"kind,omitempty"`
	Name                 string                 `json:"name"`
	Description          *string                `json:"description"`
	Target               map[string]interface{} `json:"target"`
	OwnerUserID          *string                `json:"owner_user_id"`
	RiskLevel            string                 `json:"risk_level"`
	ApprovalRules        map[string]interface{} `json:"approval_rules"`
	MaxDurationSeconds   *int64                 `json:"max_duration_seconds"`
	Requestable          bool                   `json:"requestable"`
	RequireJustification bool                   `json:"require_justification"`
}

// IgaSodRule mirrors /v1/iga/sod_rules.
type IgaSodRule struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  *string         `json:"description"`
	EntitlementA string          `json:"entitlement_a"`
	EntitlementB string          `json:"entitlement_b"`
	Mode         string          `json:"mode"`
	Approver     json.RawMessage `json:"approver"`
	Enabled      bool            `json:"enabled"`
}

// IgaSodRuleWrite is the create payload; an update sends the mutable fields.
type IgaSodRuleWrite struct {
	Name         string                 `json:"name"`
	Description  *string                `json:"description,omitempty"`
	EntitlementA string                 `json:"entitlement_a,omitempty"`
	EntitlementB string                 `json:"entitlement_b,omitempty"`
	Mode         string                 `json:"mode"`
	Approver     map[string]interface{} `json:"approver,omitempty"`
	Enabled      bool                   `json:"enabled"`
}

// IgaBirthrightRule mirrors /v1/iga/birthright_rules.
type IgaBirthrightRule struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Match           json.RawMessage `json:"match"`
	EntitlementIDs  []string        `json:"entitlement_ids"`
	RevokeOnUnmatch bool            `json:"revoke_on_unmatch"`
	Enabled         bool            `json:"enabled"`
}

// IgaBirthrightRuleWrite is the create / update payload.
type IgaBirthrightRuleWrite struct {
	Name            string                 `json:"name"`
	Match           map[string]interface{} `json:"match"`
	EntitlementIDs  []string               `json:"entitlement_ids"`
	RevokeOnUnmatch bool                   `json:"revoke_on_unmatch"`
	Enabled         bool                   `json:"enabled"`
}

func igaPath(kind, id string) string {
	if id == "" {
		return "/v1/iga/" + kind
	}
	return "/v1/iga/" + kind + "/" + url.PathEscape(id)
}

func (c *Client) CreateIgaEntitlement(ctx context.Context, body IgaEntitlementWrite) (*IgaEntitlement, error) {
	var out IgaEntitlement
	if err := c.do(ctx, http.MethodPost, igaPath("entitlements", ""), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetIgaEntitlement(ctx context.Context, id string) (*IgaEntitlement, error) {
	var out IgaEntitlement
	if err := c.do(ctx, http.MethodGet, igaPath("entitlements", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateIgaEntitlement(ctx context.Context, id string, body IgaEntitlementWrite) (*IgaEntitlement, error) {
	body.Key, body.Kind = "", ""
	var out IgaEntitlement
	if err := c.do(ctx, http.MethodPatch, igaPath("entitlements", id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteIgaEntitlement(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, igaPath("entitlements", id), nil, nil)
}

func (c *Client) CreateIgaSodRule(ctx context.Context, body IgaSodRuleWrite) (*IgaSodRule, error) {
	var out IgaSodRule
	if err := c.do(ctx, http.MethodPost, igaPath("sod_rules", ""), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIgaSodRule reads one rule from the list: the API has no single-rule read.
func (c *Client) GetIgaSodRule(ctx context.Context, id string) (*IgaSodRule, error) {
	var list struct {
		Data []IgaSodRule `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet, igaPath("sod_rules", ""), nil, &list); err != nil {
		return nil, err
	}
	for i := range list.Data {
		if list.Data[i].ID == id {
			return &list.Data[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Errors: []ErrorItem{{Message: "Unknown separation-of-duties rule."}}}
}

func (c *Client) UpdateIgaSodRule(ctx context.Context, id string, body IgaSodRuleWrite) (*IgaSodRule, error) {
	body.EntitlementA, body.EntitlementB = "", ""
	var out IgaSodRule
	if err := c.do(ctx, http.MethodPatch, igaPath("sod_rules", id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteIgaSodRule(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, igaPath("sod_rules", id), nil, nil)
}

func (c *Client) CreateIgaBirthrightRule(ctx context.Context, body IgaBirthrightRuleWrite) (*IgaBirthrightRule, error) {
	var out IgaBirthrightRule
	if err := c.do(ctx, http.MethodPost, igaPath("birthright_rules", ""), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIgaBirthrightRule reads one rule from the list: the API has no single-rule read.
func (c *Client) GetIgaBirthrightRule(ctx context.Context, id string) (*IgaBirthrightRule, error) {
	var list struct {
		Data []IgaBirthrightRule `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet, igaPath("birthright_rules", ""), nil, &list); err != nil {
		return nil, err
	}
	for i := range list.Data {
		if list.Data[i].ID == id {
			return &list.Data[i], nil
		}
	}
	return nil, &APIError{Status: http.StatusNotFound, Errors: []ErrorItem{{Message: "Unknown birthright rule."}}}
}

func (c *Client) UpdateIgaBirthrightRule(ctx context.Context, id string, body IgaBirthrightRuleWrite) (*IgaBirthrightRule, error) {
	var out IgaBirthrightRule
	if err := c.do(ctx, http.MethodPatch, igaPath("birthright_rules", id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteIgaBirthrightRule(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, igaPath("birthright_rules", id), nil, nil)
}
