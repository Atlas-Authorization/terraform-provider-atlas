package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

func unkS() types.String           { return types.StringUnknown() }
func nullS() types.String          { return types.StringNull() }
func strVal(s string) types.String { return types.StringValue(s) }
func unkI() types.Int64            { return types.Int64Unknown() }
