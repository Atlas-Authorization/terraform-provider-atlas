// Package shim exposes the Atlas Terraform provider's plugin-framework
// constructor to code outside this module — in particular the Pulumi provider
// (pulumi-atlas/), which bridges this provider via pulumi-terraform-bridge.
//
// The real implementation lives in internal/provider, which Go forbids any
// other module from importing. This package sits at the module root, so it is
// allowed to import internal/provider and re-export a stable, public entry
// point. Keep it tiny: just the constructor the bridge needs.
package shim

import (
	"github.com/atlas/terraform-provider-atlas/internal/provider"
	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
)

// NewProvider returns the Atlas terraform-plugin-framework provider for the
// given build version. The Pulumi bridge wraps the returned value with
// pf/tfbridge.ShimProvider.
func NewProvider(version string) fwprovider.Provider {
	return provider.New(version)()
}
