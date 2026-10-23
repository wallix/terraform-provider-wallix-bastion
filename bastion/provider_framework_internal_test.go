package bastion

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// clearProviderEnv unsets the env variables of the provider arguments for the test.
func clearProviderEnv(t *testing.T) {
	t.Helper()
	for _, env := range []string{
		envBastionHost, envBastionUser, envBastionPort, envBastionToken, envBastionPassword,
		envBastionAPIVersion, envSessionTimeout, envCSRFEnabled, envInsecureSkipVerify,
	} {
		t.Setenv(env, "")
	}
}

// objectValue returns a value of typ with the given attributes, the others null.
func objectValue(typ tftypes.Object, values map[string]interface{}) tftypes.Value {
	attrs := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for name, attrType := range typ.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, values[name])
	}

	return tftypes.NewValue(typ, attrs)
}

// tf5muxserver rejects provider schemas that differ, and SDKv2 exports ip and user as Optional when
// their env variable is set: GetProviderSchema must succeed in both cases.
func TestProviderServerSchemasMatch(t *testing.T) {
	for name, envSet := range map[string]bool{"env unset": false, "env set": true} {
		t.Run(name, func(t *testing.T) {
			clearProviderEnv(t)
			if envSet {
				t.Setenv(envBastionHost, "192.0.2.1")
				t.Setenv(envBastionUser, "tf")
			}
			server, err := NewProviderServer(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			resp, err := server().GetProviderSchema(t.Context(), &tfprotov5.GetProviderSchemaRequest{})
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range resp.Diagnostics {
				t.Errorf("%s: %s: %s", d.Severity, d.Summary, d.Detail)
			}
			for _, attr := range resp.Provider.Block.Attributes {
				if (attr.Name == "ip" || attr.Name == skUser) && attr.Required == envSet {
					t.Errorf("%s: required = %t, want %t", attr.Name, attr.Required, !envSet)
				}
			}
		})
	}
}

// Both servers are configured with the same request: they must get the same client, one session.
func TestProviderServersShareTheClient(t *testing.T) {
	explicit := map[string]interface{}{
		"ip": "192.0.2.10", skUser: "tf", "token": "secret-a", "port": 8443,
		"api_version": VersionWallixAPI38, "csrf_enabled": false, "insecure_skip_verify": true,
	}
	for name, tc := range map[string]struct {
		config map[string]interface{}
		env    map[string]string
	}{
		"explicit": {config: explicit},
		"defaults": {config: map[string]interface{}{"ip": "192.0.2.12", skUser: "tf", "password": "secret-c"}},
		"env": {env: map[string]string{
			envBastionHost: "192.0.2.11", envBastionUser: "tf", envBastionToken: "secret-b",
			envBastionPort: "8443", envCSRFEnabled: "false", envSessionTimeout: "300",
			envBastionAPIVersion: VersionWallixAPI38, envInsecureSkipVerify: "1", envBastionPassword: "secret-d",
		}},
	} {
		t.Run(name, func(t *testing.T) {
			clearProviderEnv(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			sdk := Provider()
			if diags := sdk.Configure(t.Context(), terraform.NewResourceConfigRaw(tc.config)); diags.HasError() {
				t.Fatalf("SDKv2 configure: %v", diags)
			}

			fw := &frameworkProvider{}
			var schemaResp provider.SchemaResponse
			fw.Schema(t.Context(), provider.SchemaRequest{}, &schemaResp)
			typ, ok := schemaResp.Schema.Type().TerraformType(t.Context()).(tftypes.Object)
			if !ok {
				t.Fatal("provider schema type is not an object")
			}
			var resp provider.ConfigureResponse
			fw.Configure(t.Context(), provider.ConfigureRequest{
				Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: objectValue(typ, tc.config)},
			}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Framework configure: %v", resp.Diagnostics)
			}

			if resp.ResourceData != sdk.Meta() || resp.DataSourceData != sdk.Meta() {
				t.Errorf("Framework client %p, SDKv2 client %p: want the same", resp.ResourceData, sdk.Meta())
			}
		})
	}
}

func TestSharedClient(t *testing.T) {
	config := Config{BastionIP: "192.0.2.20", BastionUser: "tf", BastionToken: "secret", BastionPort: 443}

	clients := make([]*Client, 20)
	var wg sync.WaitGroup
	for i := range clients {
		wg.Go(func() {
			client, diags := sharedClient(config)
			if diags.HasError() {
				t.Errorf("sharedClient: %v", diags)
			}
			clients[i] = client
		})
	}
	wg.Wait()
	for _, c := range clients {
		if c == nil || c != clients[0] {
			t.Fatalf("got clients %p and %p for the same configuration", c, clients[0])
		}
	}

	// Two aliases with different credentials keep their own session.
	other := config
	other.BastionToken = "other-secret"
	if c, _ := sharedClient(other); c == clients[0] {
		t.Error("got the same client for different tokens")
	}

	for name, invalid := range map[string]Config{
		"no ip":    {BastionUser: "tf"},
		"no user":  {BastionIP: "192.0.2.20"},
		"bad port": {BastionIP: "192.0.2.20", BastionUser: "tf", BastionPort: 70000},
	} {
		if _, diags := sharedClient(invalid); !diags.HasError() {
			t.Errorf("%s: want an error", name)
		}
	}
}

// The Framework server would repeat the deprecation warning of session_timeout.
func TestProviderServerWarnsOnce(t *testing.T) {
	clearProviderEnv(t)
	server, err := NewProviderServer(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	schemaResp, err := server().GetProviderSchema(t.Context(), &tfprotov5.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	typ, ok := schemaResp.Provider.ValueType().(tftypes.Object)
	if !ok {
		t.Fatal("provider schema type is not an object")
	}
	config, err := tfprotov5.NewDynamicValue(typ, objectValue(typ, map[string]interface{}{
		"ip": "192.0.2.1", skUser: "tf", "session_timeout": 300,
	}))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := server().PrepareProviderConfig(t.Context(), &tfprotov5.PrepareProviderConfigRequest{Config: &config})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Diagnostics) != 1 || resp.Diagnostics[0].Severity != tfprotov5.DiagnosticSeverityWarning {
		for _, d := range resp.Diagnostics {
			t.Logf("%s: %s", d.Severity, d.Summary)
		}
		t.Errorf("got %d diagnostics, want the deprecation warning once", len(resp.Diagnostics))
	}
}
