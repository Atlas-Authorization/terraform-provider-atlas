// Command terraform-provider-atlas is the Terraform provider for Atlas, the
// Clerk-class authentication platform. It lets customers manage their Atlas
// instance configuration — OAuth clients, SSO connections, resource servers,
// JWT templates, webhook endpoints, roles, organizations and custom domains —
// as declarative infrastructure-as-code, wrapping the secret-key Backend API.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/atlas/terraform-provider-atlas/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// Matches the registry address customers put in their required_providers
		// block: `source = "Atlas-Authorization/atlas"`.
		Address: "registry.terraform.io/Atlas-Authorization/atlas",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err.Error())
	}
}
