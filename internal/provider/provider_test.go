package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/atlas/terraform-provider-atlas/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories wires the in-process provider into the
// terraform-plugin-testing harness under the name "atlas". Every acceptance
// test references it via ProtoV6ProviderFactories so no separately-built or
// registry-published binary is needed — the framework serves the provider over
// an in-memory gRPC channel.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"atlas": providerserver.NewProtocol6WithError(New("test")()),
}

// providerConfig is prepended to every acceptance-test HCL config. It leaves the
// api_url / secret_key attributes unset on purpose so the provider falls back to
// the ATLAS_API_URL / ATLAS_SECRET_KEY environment variables that
// testAccPreCheck asserts are present. Keeping credentials in the environment
// (never in HCL) mirrors how the provider is meant to be configured.
const providerConfig = `
provider "atlas" {}
`

// testAccPreCheck runs once, before the steps of each resource.Test, but ONLY
// when TF_ACC is set (resource.Test itself skips the entire test otherwise, so
// `go test ./...` with no TF_ACC is a no-op and CI stays green without a live
// server). When acceptance mode IS on, a missing credential is a hard failure
// rather than a silent skip: the test author asked for a real run, so surface
// the misconfiguration loudly instead of pretending to pass.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("ATLAS_SECRET_KEY"); v == "" {
		t.Fatal("ATLAS_SECRET_KEY must be set for acceptance tests (a real sk_ instance secret key)")
	}
	if v := os.Getenv("ATLAS_API_URL"); v == "" {
		t.Fatal("ATLAS_API_URL must be set for acceptance tests (the Backend API origin of the disposable test instance)")
	}
}

// testAccClient builds a Backend API client from the same environment the
// provider uses. Out-of-band helpers (the "disappears" checks) call the API
// directly to delete an object behind Terraform's back and assert the next plan
// notices. Only invoked inside a running acceptance test, after testAccPreCheck.
func testAccClient() *client.Client {
	return client.New(os.Getenv("ATLAS_API_URL"), os.Getenv("ATLAS_SECRET_KEY"))
}

// testAccCaptureAttr records the current value of a state attribute into dst, so
// a later step can assert whether it changed (e.g. an id after a replacement).
func testAccCaptureAttr(resourceName, attr string, dst *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		*dst = rs.Primary.Attributes[attr]
		return nil
	}
}

// testAccCheckAttrChanged asserts the captured attribute differs from its
// current value — the signature of a resource that was replaced, not updated in
// place.
func testAccCheckAttrChanged(resourceName, attr string, old *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		if now := rs.Primary.Attributes[attr]; now == *old {
			return fmt.Errorf("expected %s.%s to change (resource replaced), but it stayed %q", resourceName, attr, now)
		}
		return nil
	}
}

// testAccCheckAttrUnchanged asserts the captured attribute matches its current
// value — the signature of an in-place update (no replacement, value preserved).
func testAccCheckAttrUnchanged(resourceName, attr string, old *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		if now := rs.Primary.Attributes[attr]; now != *old {
			return fmt.Errorf("expected %s.%s to be preserved (in-place update), but it changed from %q to %q", resourceName, attr, *old, now)
		}
		return nil
	}
}

// deleteOutOfBand returns a check that deletes the resource through the Backend
// API directly, simulating drift. Pairing it with ExpectNonEmptyPlan: true
// proves the provider's Read drops a vanished object from state and re-plans it.
func deleteOutOfBand(resourceName string, del func(ctx context.Context, c *client.Client, id string) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id := rs.Primary.Attributes["id"]
		if id == "" {
			return fmt.Errorf("resource %s has no id in state", resourceName)
		}
		return del(context.Background(), testAccClient(), id)
	}
}
