package bastion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = (*authDomainOIDCDataSource)(nil)

type authDomainOIDCDataSource struct {
	client *Client
}

func (d *authDomainOIDCDataSource) Metadata(
	_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_authdomain_oidc"
}

func (d *authDomainOIDCDataSource) Schema(
	_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse,
) {
	stringList := schema.ListAttribute{Computed: true, ElementType: types.StringType}
	resp.Schema = schema.Schema{
		Description: "OIDC authentication domain (API v3.12 and later).",
		Attributes: map[string]schema.Attribute{
			"id":                 schema.StringAttribute{Computed: true},
			skDomainName:         schema.StringAttribute{Required: true},
			skAuthDomainName:     schema.StringAttribute{Computed: true},
			skDefaultEmailDomain: schema.StringAttribute{Computed: true},
			skDefaultLanguage:    schema.StringAttribute{Computed: true},
			skExternalAuths:      stringList,
			skLabel:              schema.StringAttribute{Computed: true},
			skDescription:        schema.StringAttribute{Computed: true},
			skIsDefault:          schema.BoolAttribute{Computed: true},
			skSecondaryAuth:      stringList,
		},
	}
}

func (d *authDomainOIDCDataSource) Configure(
	_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse,
) {
	d.client = providerClient(req.ProviderData, &resp.Diagnostics)
}

func (d *authDomainOIDCDataSource) Read(
	ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse,
) {
	var data authDomainOIDCModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := oidcVersionCheck("data source", "wallix-bastion_authdomain_oidc", d.client.bastionAPIVersion); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	name := data.DomainName.ValueString()
	id, err := searchAuthDomainOIDC(ctx, d.client, name)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	var cfg jsonAuthDomainOIDC
	if id != "" {
		if cfg, err = readAuthDomainOIDC(ctx, d.client, id); err != nil {
			resp.Diagnostics.AddError(err.Error(), "")

			return
		}
	}
	if cfg.ID == "" || cfg.Type != typeOIDC {
		resp.Diagnostics.AddError(fmt.Sprintf("OIDC domain_name %s doesn't exists", name), "")

		return
	}
	// Computed values: "", false and [] are stored as the API returns them.
	resp.Diagnostics.Append(resp.State.Set(ctx, authDomainOIDCModel{
		ID:                 types.StringValue(cfg.ID),
		DomainName:         types.StringValue(cfg.DomainName),
		AuthDomainName:     types.StringValue(cfg.AuthDomainName),
		DefaultEmailDomain: types.StringValue(cfg.DefaultEmailDomain),
		DefaultLanguage:    types.StringValue(cfg.DefaultLanguage),
		ExternalAuths:      nonNil(cfg.ExternalAuths),
		Label:              types.StringValue(cfg.Label),
		Description:        types.StringValue(cfg.Description),
		IsDefault:          types.BoolValue(cfg.IsDefault),
		SecondaryAuth:      nonNil(cfg.SecondaryAuth),
	})...)
}
