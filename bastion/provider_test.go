package bastion_test

import (
	"context"
	"os"
	"testing"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
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

func TestProvider_impl(_ *testing.T) {
	var _ *schema.Provider = bastion.Provider()
}

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
