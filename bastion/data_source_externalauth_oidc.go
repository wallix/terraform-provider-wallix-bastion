package bastion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = (*externalAuthOIDCDataSource)(nil)

type externalAuthOIDCDataSource struct {
	client *Client
}

type externalAuthOIDCDataSourceModel struct {
	ID                     types.String  `tfsdk:"id"`
	AuthenticationName     types.String  `tfsdk:"authentication_name"`
	Description            types.String  `tfsdk:"description"`
	Timeout                types.Float64 `tfsdk:"timeout"`
	ClientID               types.String  `tfsdk:"client_id"`
	DiscoveryURL           types.String  `tfsdk:"discovery_url"`
	RedirectURI            types.String  `tfsdk:"redirect_uri"`
	ClaimCustomization     types.Object  `tfsdk:"claim_customization"`
	OIDCAdvancedAttributes types.Object  `tfsdk:"oidc_advanced_attributes"`
}

var (
	oidcClaimsAttrTypes = map[string]attr.Type{ //nolint:gochecknoglobals
		skUsername: types.StringType, skGroup: types.StringType, skEmail: types.StringType,
		skDisplayname: types.StringType, skLanguage: types.StringType,
	}
	oidcAdvancedAttrTypes = map[string]attr.Type{ //nolint:gochecknoglobals
		"issuer": types.StringType, "scope": types.ListType{ElemType: types.StringType},
		"authorization_endpoint": types.StringType, "token_endpoint": types.StringType,
		"userinfo_url_endpoint": types.StringType, "grant_type": types.StringType, "response_type": types.StringType,
		"jwks_uri": types.StringType, "id_token_signing_alg_values_supported": types.ListType{ElemType: types.StringType},
	}
)

func (d *externalAuthOIDCDataSource) Metadata(
	_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_externalauth_oidc"
}

func (d *externalAuthOIDCDataSource) Schema(
	_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "OIDC external authentication (API v3.12 and later). The client secret is never read.",
		Attributes: map[string]schema.Attribute{
			"id":                 schema.StringAttribute{Computed: true},
			skAuthenticationName: schema.StringAttribute{Required: true},
			skDescription:        schema.StringAttribute{Computed: true},
			skTimeout:            schema.Float64Attribute{Computed: true},
			"client_id":          schema.StringAttribute{Computed: true},
			"discovery_url":      schema.StringAttribute{Computed: true},
			"redirect_uri":       schema.StringAttribute{Computed: true},
			"claim_customization": schema.ObjectAttribute{
				Computed: true, AttributeTypes: oidcClaimsAttrTypes,
			},
			"oidc_advanced_attributes": schema.ObjectAttribute{
				Computed: true, AttributeTypes: oidcAdvancedAttrTypes, Description: "Null when not set on the Bastion.",
			},
		},
	}
}

func (d *externalAuthOIDCDataSource) Configure(
	_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse,
) {
	d.client = providerClient(req.ProviderData, &resp.Diagnostics)
}

func (d *externalAuthOIDCDataSource) Read(
	ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse,
) {
	var data externalAuthOIDCDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := oidcVersionCheck("data source", "wallix-bastion_externalauth_oidc", d.client.bastionAPIVersion); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	name := data.AuthenticationName.ValueString()
	id, err := searchExternalAuthOIDC(ctx, d.client, name)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	var cfg jsonExternalAuthOIDC
	if id != "" {
		if cfg, err = readExternalAuthOIDC(ctx, d.client, id); err != nil {
			resp.Diagnostics.AddError(err.Error(), "")

			return
		}
	}
	if cfg.ID == "" || cfg.Type != typeOIDC {
		resp.Diagnostics.AddError(fmt.Sprintf("OIDC authentication_name %s doesn't exists", name), "")

		return
	}
	c := cfg.ClaimCustomization
	claims, diags := types.ObjectValue(oidcClaimsAttrTypes, map[string]attr.Value{
		skUsername: types.StringValue(c.Username), skGroup: types.StringValue(c.Group),
		skEmail: types.StringValue(c.Email), skDisplayname: types.StringValue(c.Displayname),
		skLanguage: types.StringValue(c.Language),
	})
	resp.Diagnostics.Append(diags...)
	advanced := types.ObjectNull(oidcAdvancedAttrTypes)
	if a := cfg.OIDCAdvancedAttributes; a != nil {
		scope, diags := types.ListValueFrom(ctx, types.StringType, nonNil(a.Scope))
		resp.Diagnostics.Append(diags...)
		algs, diags := types.ListValueFrom(ctx, types.StringType, nonNil(a.IDTokenSigningAlgValuesSupported))
		resp.Diagnostics.Append(diags...)
		advanced, diags = types.ObjectValue(oidcAdvancedAttrTypes, map[string]attr.Value{
			"issuer": types.StringValue(a.Issuer), "scope": scope,
			"authorization_endpoint": types.StringValue(a.AuthorizationEndpoint),
			"token_endpoint":         types.StringValue(a.TokenEndpoint),
			"userinfo_url_endpoint":  types.StringValue(a.UserinfoURLEndpoint),
			"grant_type":             types.StringValue(a.GrantType), "response_type": types.StringValue(a.ResponseType),
			"jwks_uri": types.StringValue(a.JwksURI), "id_token_signing_alg_values_supported": algs,
		})
		resp.Diagnostics.Append(diags...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(cfg.ID)
	data.AuthenticationName = types.StringValue(cfg.AuthenticationName)
	data.Description = types.StringValue(cfg.Description)
	data.Timeout = types.Float64Value(cfg.Timeout)
	data.ClientID = types.StringValue(cfg.ClientID)
	data.DiscoveryURL = types.StringValue(cfg.DiscoveryURL)
	data.RedirectURI = types.StringValue(cfg.RedirectURI)
	data.ClaimCustomization = claims
	data.OIDCAdvancedAttributes = advanced
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
