package bastion

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProviderTLSSchemaOption(t *testing.T) {
	provider := Provider()

	if provider.Schema["insecure_skip_verify"] == nil {
		t.Fatal("Expected 'insecure_skip_verify' option in provider schema")
	}

	tlsOption := provider.Schema["insecure_skip_verify"]
	if tlsOption.Type != schema.TypeBool {
		t.Errorf("Expected 'insecure_skip_verify' to be TypeBool, got %v", tlsOption.Type)
	}

	if !tlsOption.Optional {
		t.Error("Expected 'insecure_skip_verify' to be optional")
	}
}

func TestProviderTLSEnvironmentVariable(t *testing.T) {
	t.Setenv("WALLIX_INSECURE_SKIP_VERIFY", "true")

	provider := Provider()
	tlsOption := provider.Schema["insecure_skip_verify"]

	if tlsOption.DefaultFunc == nil {
		t.Fatal("Expected default function for insecure_skip_verify")
	}

	val, _ := tlsOption.DefaultFunc()
	if val != "true" {
		t.Errorf("Expected WALLIX_INSECURE_SKIP_VERIFY env var to be used, got %v", val)
	}
}

func TestProviderTLSDefaultSecure(t *testing.T) {
	// Save current env var state
	originalValue := os.Getenv("WALLIX_INSECURE_SKIP_VERIFY")
	defer func() {
		if originalValue != "" {
			os.Setenv("WALLIX_INSECURE_SKIP_VERIFY", originalValue)
		}
	}()

	// Clear any existing env var
	os.Unsetenv("WALLIX_INSECURE_SKIP_VERIFY")

	provider := Provider()
	tlsOption := provider.Schema["insecure_skip_verify"]

	if tlsOption.DefaultFunc == nil {
		t.Fatal("Expected default function for insecure_skip_verify")
	}

	val, _ := tlsOption.DefaultFunc()
	if val != false {
		t.Errorf("Expected default to be false (secure), got %v", val)
	}
}
