package bastion_test

import (
	"context"
	"os"
	"testing"

	"github.com/claranet/terraform-provider-wallix-bastion/bastion"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	testAccProviders = map[string]*schema.Provider{ //nolint: gochecknoglobals
		"wallix-bastion": testAccProvider,
	}
	testAccProvider = bastion.Provider() //nolint: gochecknoglobals
)

func TestProvider(t *testing.T) {
	if err := bastion.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = bastion.Provider()
}

// export TESTACC_SWITCH not empty for test switch options (interface mode trunk, vlan native/members)
// with switch Junos device, else it's test for all others parameters
// (interface inet, 802.3ad, routing instance, security zone/nat/ike/ipsec, etc  ).
// Few resources and parameters works on both devices, but most tested without TESTACC_SWITCH

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("WALLIX_BASTION_HOST") == "" {
		t.Fatal("WALLIX_BASTION_HOST must be set for acceptance tests")
	}
	if os.Getenv("WALLIX_BASTION_TOKEN") == "" {
		t.Fatal("WALLIX_BASTION_TOKEN must be set for acceptance tests")
	}

	if err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil)); err != nil {
		t.Fatal(err)
	}
}
