package bastion_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	testAccProvider = bastion.Provider() //nolint: gochecknoglobals
	// Every acceptance test runs the muxed provider, as Terraform does.
	testAccProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){ //nolint: gochecknoglobals
		"wallix-bastion": func() (tfprotov5.ProviderServer, error) {
			server, err := bastion.NewProviderServer(context.Background())
			if err != nil {
				return nil, fmt.Errorf("creating the provider server: %w", err)
			}

			return server(), nil
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

// A default value would make Terraform warn about the deprecated argument in every configuration.
func TestProviderSessionTimeoutDeprecated(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		env      string
		wantWarn bool
	}{
		{name: "unset", config: map[string]interface{}{}},
		{name: "argument", config: map[string]interface{}{"session_timeout": 300}, wantWarn: true},
		{name: "env", config: map[string]interface{}{}, env: "300", wantWarn: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envWallixHost, "192.0.2.1")
			t.Setenv(envWallixUser, "tf")
			t.Setenv("WALLIX_SESSION_TIMEOUT", tt.env)

			diags := bastion.Provider().Validate(terraform.NewResourceConfigRaw(tt.config))
			if diags.HasError() {
				t.Fatalf("unexpected errors: %v", diags)
			}
			warned := false
			for _, d := range diags {
				warned = warned || (d.Severity == diag.Warning && d.Summary == "Argument is deprecated")
			}
			if warned != tt.wantWarn {
				t.Errorf("deprecation warning: got %t, want %t", warned, tt.wantWarn)
			}
		})
	}
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
