package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
)

// NewProviderServer returns the protocol 5 server of the provider, which muxes its SDKv2 and
// Framework servers.
func NewProviderServer(ctx context.Context) (func() tfprotov5.ProviderServer, error) {
	framework := providerserver.NewProtocol5(&frameworkProvider{})
	mux, err := tf5muxserver.NewMuxServer(ctx,
		Provider().GRPCProvider,
		func() tfprotov5.ProviderServer { return frameworkServer{framework()} },
	)
	if err != nil {
		return nil, fmt.Errorf("muxing the provider servers: %w", err)
	}

	return mux.ProviderServer, nil
}

// frameworkServer drops the Framework's deprecation warnings on the provider config, which repeat
// the SDKv2 ones on the identical schema. Embedding hides the optional ListResourceServer and
// ActionServer (unused).
type frameworkServer struct {
	tfprotov5.ProviderServer
}

func (s frameworkServer) PrepareProviderConfig(
	ctx context.Context, req *tfprotov5.PrepareProviderConfigRequest,
) (
	*tfprotov5.PrepareProviderConfigResponse, error,
) {
	resp, err := s.ProviderServer.PrepareProviderConfig(ctx, req)
	if resp != nil {
		resp.Diagnostics = slices.DeleteFunc(resp.Diagnostics, func(d *tfprotov5.Diagnostic) bool {
			// Summary of the Framework's DeprecationMessage warning; SDKv2 reports the same text.
			return d.Severity == tfprotov5.DiagnosticSeverityWarning && d.Summary == "Attribute Deprecated"
		})
	}

	return resp, err //nolint:wrapcheck
}
