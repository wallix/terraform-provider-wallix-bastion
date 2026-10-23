package bastion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
)

// NewProviderServer returns the protocol 5 server of the provider, which muxes its SDKv2 and
// Framework servers.
func NewProviderServer(ctx context.Context) (func() tfprotov5.ProviderServer, error) {
	mux, err := tf5muxserver.NewMuxServer(ctx,
		Provider().GRPCProvider,
	)
	if err != nil {
		return nil, fmt.Errorf("muxing the provider servers: %w", err)
	}

	return mux.ProviderServer, nil
}
