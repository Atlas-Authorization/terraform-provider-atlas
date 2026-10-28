package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccGeneric_examples applies each generated resource's example config,
// imports it and re-plans. It needs a live instance (TF_ACC) and is further
// opt-in with ATLAS_ACC_GENERIC=1, because the examples carry placeholder values
// that some resources validate against real objects (a parent id, a live KMS key).
// Resources with a parent / path key are skipped: they need the parent created
// first.
func TestAccGeneric_examples(t *testing.T) {
	if os.Getenv("ATLAS_ACC_GENERIC") == "" {
		t.Skip("set ATLAS_ACC_GENERIC=1 (with TF_ACC) to run the generated examples against a live instance")
	}
	for _, spec := range genResourceSpecs() {
		spec := spec
		cfg, ok := genExampleConfigs[spec.Type]
		if !ok || spec.Kind == "dataonly" || len(spec.params()) > 0 {
			continue
		}
		t.Run(spec.Type, func(t *testing.T) {
			name := "atlas_" + spec.Type + ".example"
			ignore := []string{}
			for _, a := range spec.Attrs {
				if a.WriteOnly {
					ignore = append(ignore, a.Name)
				}
			}
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: providerConfig + cfg,
						Check:  resource.TestCheckResourceAttrSet(name, "id"),
					},
					{
						ResourceName:            name,
						ImportState:             true,
						ImportStateVerify:       spec.Kind == "coll" && !spec.NoRead,
						ImportStateVerifyIgnore: ignore,
					},
				},
			})
		})
	}
}
