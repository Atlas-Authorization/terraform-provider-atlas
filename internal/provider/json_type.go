package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// jsonObjectType is a string attribute type holding a JSON object, with
// SEMANTIC equality: two values are equal when they decode to the same document
// after dropping null members and empty objects (null means "not set" on every
// Atlas design / branding surface). Whitespace and key order never cause a diff,
// so `jsonencode({...})` and a heredoc of the same object plan identically, and
// the API echoing defaults as nulls does not thrash the plan.
type jsonObjectType struct {
	basetypes.StringType
}

var (
	_ basetypes.StringTypable                    = jsonObjectType{}
	_ basetypes.StringValuableWithSemanticEquals = jsonObjectValue{}
	_ xattr.ValidateableAttribute                = jsonObjectValue{}
)

func (t jsonObjectType) String() string { return "jsonObjectType" }

func (t jsonObjectType) ValueType(context.Context) attr.Value { return jsonObjectValue{} }

func (t jsonObjectType) Equal(o attr.Type) bool {
	_, ok := o.(jsonObjectType)
	return ok
}

func (t jsonObjectType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return jsonObjectValue{StringValue: in}, nil
}

func (t jsonObjectType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	v, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	s, ok := v.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T", v)
	}
	return jsonObjectValue{StringValue: s}, nil
}

// jsonObjectValue is the value of a jsonObjectType attribute.
type jsonObjectValue struct {
	basetypes.StringValue
}

func newJSONObjectNull() jsonObjectValue {
	return jsonObjectValue{StringValue: basetypes.NewStringNull()}
}

func newJSONObjectValue(s string) jsonObjectValue {
	return jsonObjectValue{StringValue: basetypes.NewStringValue(s)}
}

func (v jsonObjectValue) Type(context.Context) attr.Type { return jsonObjectType{} }

func (v jsonObjectValue) Equal(o attr.Value) bool {
	other, ok := o.(jsonObjectValue)
	return ok && v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals reports whether the two documents are the same object
// once nulls and empty objects are dropped.
func (v jsonObjectValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	nv, ok := newValuable.(jsonObjectValue)
	if !ok {
		diags.AddError("Semantic equality check error", fmt.Sprintf("unexpected value type %T", newValuable))
		return false, diags
	}
	a, errA := normalizeJSONDocument(v.ValueString())
	b, errB := normalizeJSONDocument(nv.ValueString())
	if errA != nil || errB != nil {
		return false, nil
	}
	return a == b, nil
}

// ValidateAttribute rejects a value that is not a JSON object.
func (v jsonObjectValue) ValidateAttribute(_ context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(v.ValueString()), &obj); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON object",
			"The value must be a JSON object (use jsonencode({...})): "+err.Error())
	}
}

// normalizeJSONDocument canonicalises a JSON document: compact, key-sorted,
// with null members and (recursively) empty objects removed.
func normalizeJSONDocument(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "{}", nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "", err
	}
	v = pruneJSON(v)
	if v == nil {
		v = map[string]interface{}{}
	}
	out, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// pruneJSON drops null members and empty objects, recursively. Arrays keep
// their elements (order and nulls inside an array are meaningful).
func pruneJSON(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		out := map[string]interface{}{}
		for k, child := range t {
			pc := pruneJSON(child)
			if pc == nil {
				continue
			}
			if m, ok := pc.(map[string]interface{}); ok && len(m) == 0 {
				continue
			}
			out[k] = pc
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, child := range t {
			pc := pruneJSON(child)
			if m, ok := pc.(map[string]interface{}); ok && len(m) == 0 {
				pc = map[string]interface{}{}
			}
			out[i] = pc
		}
		return out
	default:
		return v
	}
}

// jsonObjectFromAPI maps an API JSON object onto the attribute value, pruned
// of nulls/empties. An absent, null or empty object becomes a null attribute.
func jsonObjectFromAPI(raw json.RawMessage) jsonObjectValue {
	if len(raw) == 0 || string(raw) == "null" {
		return newJSONObjectNull()
	}
	canon, err := normalizeJSONDocument(string(raw))
	if err != nil {
		return newJSONObjectValue(string(raw))
	}
	if canon == "{}" {
		return newJSONObjectNull()
	}
	return newJSONObjectValue(canon)
}

// jsonObjectToMap decodes a non-null attribute into a map; null/unknown yields nil.
func jsonObjectToMap(v jsonObjectValue, p path.Path, diags *diag.Diagnostics) map[string]interface{} {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(v.ValueString()), &out); err != nil {
		diags.AddAttributeError(p, "Invalid JSON object", err.Error())
		return nil
	}
	return out
}

// mergePatchFor builds a merge-patch that turns `prior` into `desired` on a
// server that merges objects per field and treats null as "clear": every key
// of desired is sent, every key present in prior but absent from desired is
// sent as null, and nested objects present on both sides are diffed the same
// way. This gives replace semantics on top of a merging PATCH.
func mergePatchFor(prior, desired map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, dv := range desired {
		pm, pok := prior[k].(map[string]interface{})
		dm, dok := dv.(map[string]interface{})
		if pok && dok {
			out[k] = mergePatchFor(pm, dm)
			continue
		}
		out[k] = dv
	}
	for k := range prior {
		if _, ok := desired[k]; !ok {
			out[k] = nil
		}
	}
	return out
}
