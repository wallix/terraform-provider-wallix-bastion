package bastion_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
)

const (
	envWallixHost               = "WALLIX_BASTION_HOST"
	envWallixToken              = "WALLIX_BASTION_TOKEN"
	envWallixUser               = "WALLIX_BASTION_USER"
	envWallixInsecureSkipVerify = "WALLIX_INSECURE_SKIP_VERIFY"
)

func init() { //nolint:gochecknoinits
	// Ensure TLS verification bypass is set for acceptance tests with self-signed certs
	if os.Getenv(envWallixInsecureSkipVerify) == "" {
		os.Setenv(envWallixInsecureSkipVerify, "true")
	}
}

var (
	testAccProvider          = bastion.Provider()                           //nolint: gochecknoglobals
	testAccProviderFactories = map[string]func() (*schema.Provider, error){ //nolint: gochecknoglobals
		"wallix-bastion": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
)

func TestProvider(t *testing.T) {
	t.Parallel()

	if err := bastion.Provider().InternalValidate(); err != nil {
		t.Fatalf("provider validation failed: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	t.Parallel()

	// Explicit type is the point of this compile-time assertion.
	var _ *schema.Provider = bastion.Provider() //nolint:staticcheck
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	requiredEnvVars := []string{
		envWallixHost,
		envWallixToken,
		envWallixUser,
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}

	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"insecure_skip_verify": true,
	})

	if err := testAccProvider.Configure(context.Background(), config); err != nil {
		t.Fatalf("failed to configure provider: %v", err)
	}
}
