package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/mod/semver"
)

const (
	typeOIDC = "OIDC"
	// Value of client_secret on read when the Bastion holds no secret; it masks a set one.
	oidcNoClientSecret = ""
)

type jsonExternalAuthOIDC struct {
	ID                     string                        `json:"id,omitempty"`
	AuthenticationName     string                        `json:"authentication_name"`
	Type                   string                        `json:"type"`
	Description            string                        `json:"description"`
	Timeout                float64                       `json:"timeout"`
	ClientID               string                        `json:"client_id"`
	ClientSecret           *string                       `json:"client_secret,omitempty"`
	DiscoveryURL           string                        `json:"discovery_url"`
	RedirectURI            string                        `json:"redirect_uri,omitempty"`
	ClaimCustomization     jsonExternalAuthOIDCClaims    `json:"claim_customization"`
	OIDCAdvancedAttributes *jsonExternalAuthOIDCAdvanced `json:"oidc_advanced_attributes,omitempty"`
}

// The API requires username and group, and replaces the whole object on update.
type jsonExternalAuthOIDCClaims struct {
	Username    string `json:"username"`
	Group       string `json:"group"`
	Email       string `json:"email"`
	Displayname string `json:"displayname"`
	Language    string `json:"language"`
}

// The API requires every field but userinfo_url_endpoint, and refuses to remove the object.
type jsonExternalAuthOIDCAdvanced struct {
	Issuer                           string   `json:"issuer"`
	Scope                            []string `json:"scope"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	UserinfoURLEndpoint              string   `json:"userinfo_url_endpoint,omitempty"`
	GrantType                        string   `json:"grant_type"`
	ResponseType                     string   `json:"response_type"`
	JwksURI                          string   `json:"jwks_uri"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

var (
	_ resource.ResourceWithConfigure   = (*externalAuthOIDCResource)(nil)
	_ resource.ResourceWithImportState = (*externalAuthOIDCResource)(nil)
)

type externalAuthOIDCResource struct {
	client *Client
}

type externalAuthOIDCModel struct {
	ID                     types.String       `tfsdk:"id"`
	AuthenticationName     types.String       `tfsdk:"authentication_name"`
	Description            types.String       `tfsdk:"description"`
	Timeout                types.Float64      `tfsdk:"timeout"`
	ClientID               types.String       `tfsdk:"client_id"`
	ClientSecretWO         types.String       `tfsdk:"client_secret_wo"`
	ClientSecretWOVersion  types.Int64        `tfsdk:"client_secret_wo_version"`
	DiscoveryURL           types.String       `tfsdk:"discovery_url"`
	RedirectURI            types.String       `tfsdk:"redirect_uri"`
	ClaimCustomization     *oidcClaimsModel   `tfsdk:"claim_customization"`
	OIDCAdvancedAttributes *oidcAdvancedModel `tfsdk:"oidc_advanced_attributes"`
}

type oidcClaimsModel struct {
	Username    types.String `tfsdk:"username"`
	Group       types.String `tfsdk:"group"`
	Email       types.String `tfsdk:"email"`
	Displayname types.String `tfsdk:"displayname"`
	Language    types.String `tfsdk:"language"`
}

type oidcAdvancedModel struct {
	Issuer                           types.String `tfsdk:"issuer"`
	Scope                            []string     `tfsdk:"scope"`
	AuthorizationEndpoint            types.String `tfsdk:"authorization_endpoint"`
	TokenEndpoint                    types.String `tfsdk:"token_endpoint"`
	UserinfoURLEndpoint              types.String `tfsdk:"userinfo_url_endpoint"`
	GrantType                        types.String `tfsdk:"grant_type"`
	ResponseType                     types.String `tfsdk:"response_type"`
	JwksURI                          types.String `tfsdk:"jwks_uri"`
	IDTokenSigningAlgValuesSupported []string     `tfsdk:"id_token_signing_alg_values_supported"`
}

// oidcVersionCheck: OIDC external auths and auth domains came with API v3.12.
func oidcVersionCheck(kind, name, version string) error {
	if slices.Contains(defaultVersionsValid(), version) && semver.Compare(version, VersionWallixAPI312) >= 0 {
		return nil
	}

	return fmt.Errorf("%s %s not available with api version %s", kind, name, version)
}

func (r *externalAuthOIDCResource) Metadata(
	_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_externalauth_oidc"
}

func (r *externalAuthOIDCResource) Schema(
	_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	maxLen := func(n int) []validator.String { return []validator.String{stringvalidator.LengthAtMost(n)} }
	optional := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Description: desc, Validators: maxLen(255)}
	}
	required := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{
			Required: true, Description: desc, Validators: []validator.String{stringvalidator.LengthBetween(1, 255)},
		}
	}
	// The Framework requires a Required attribute even when its block is left out.
	requiredInBlock := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{
			Optional: true, Description: desc + " Required with the block.",
			Validators: []validator.String{stringvalidator.LengthBetween(1, 255)},
		}
	}
	advancedRequired := []string{
		"issuer", "scope", "authorization_endpoint", "token_endpoint", "jwks_uri", "id_token_signing_alg_values_supported",
	}
	advancedPaths := make([]path.Expression, len(advancedRequired))
	for i, name := range advancedRequired {
		advancedPaths[i] = path.MatchRelative().AtName(name)
	}
	resp.Schema = schema.Schema{
		Description: "OIDC external authentication (API v3.12 and later).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "ID of the external authentication on the Bastion.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			skAuthenticationName: schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(1, 512)},
			},
			skDescription: schema.StringAttribute{Optional: true, Validators: maxLen(1024)},
			skTimeout: schema.Float64Attribute{
				Required:    true,
				Description: "Timeout of the requests to the identity provider, in seconds (1 to 900).",
				Validators:  []validator.Float64{float64validator.Between(1, 900)},
			},
			"client_id": schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(1, 1024)},
			},
			"client_secret_wo": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
				Description: "Client secret, write-only (Terraform 1.11 and later): sent to the Bastion, never " +
					"stored in the state. Change client_secret_wo_version to send a new value.",
				Validators: []validator.String{stringvalidator.LengthBetween(1, 1024)},
			},
			"client_secret_wo_version": schema.Int64Attribute{
				Optional: true,
				Description: "Version of client_secret_wo: a change sends the secret again. Removing both " +
					"removes the secret from the Bastion.",
				Validators: []validator.Int64{int64validator.AlsoRequires(path.MatchRoot("client_secret_wo"))},
			},
			"discovery_url": optional("URL of the OpenID Provider metadata (.well-known/openid-configuration)."),
			"redirect_uri": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Redirection URI of the Bastion, its /api/oidc URL when not set. " +
					"The Bastion does not update it: a change replaces the external authentication.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				// Not empty: "" is not sent, and the Bastion would compute another value than the plan.
				Validators: []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
		},
		Blocks: map[string]schema.Block{
			"claim_customization": schema.SingleNestedBlock{
				Description: "Names of the OIDC claims that hold the user attributes. Required.",
				Validators:  []validator.Object{objectvalidator.IsRequired()},
				Attributes: map[string]schema.Attribute{
					skUsername:    required("Claim of the user name."),
					skGroup:       optional("Claim of the user groups."),
					skEmail:       optional("Claim of the email."),
					skDisplayname: optional("Claim of the display name."),
					skLanguage:    optional("Claim of the language."),
				},
			},
			"oidc_advanced_attributes": schema.SingleNestedBlock{
				Description: "OpenID Provider metadata, as published at discovery_url. The Bastion cannot " +
					"remove it: removing the block replaces the external authentication.",
				PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplaceIf(
					func(_ context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
						resp.RequiresReplace = !req.StateValue.IsNull() && req.PlanValue.IsNull()
					},
					"Removing the block replaces the external authentication.",
					"Removing the block replaces the external authentication.",
				)},
				Validators: []validator.Object{objectvalidator.AlsoRequires(advancedPaths...)},
				Attributes: map[string]schema.Attribute{
					"issuer": requiredInBlock("Issuer identifier of the OpenID Provider."),
					"scope": schema.ListAttribute{
						Optional: true, ElementType: types.StringType, Description: "Scopes requested. Required with the block.",
					},
					"authorization_endpoint": requiredInBlock("Authorization endpoint."),
					"token_endpoint":         requiredInBlock("Token endpoint."),
					"userinfo_url_endpoint":  optional("UserInfo endpoint."),
					"grant_type": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "authorization_code, the only grant type of the Bastion.",
						Default:     stringdefault.StaticString("authorization_code"),
						Validators:  []validator.String{stringvalidator.OneOf("authorization_code")},
					},
					"response_type": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "code, the only response type of the Bastion.",
						Default:     stringdefault.StaticString("code"),
						Validators:  []validator.String{stringvalidator.OneOf("code")},
					},
					"jwks_uri": requiredInBlock("URL of the JSON Web Key Set."),
					"id_token_signing_alg_values_supported": schema.ListAttribute{
						Optional: true, ElementType: types.StringType,
						Description: "JWS signing algorithms of the ID token. Required with the block.",
					},
				},
			},
		},
	}
}

func (r *externalAuthOIDCResource) Configure(
	_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse,
) {
	r.client = providerClient(req.ProviderData, &resp.Diagnostics)
}

func (r *externalAuthOIDCResource) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan externalAuthOIDCModel
	var secret types.String
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("client_secret_wo"), &secret)...)
	if resp.Diagnostics.HasError() || !r.versionCheck(&resp.Diagnostics) {
		return
	}
	name := plan.AuthenticationName.ValueString()
	id, err := searchExternalAuthOIDC(ctx, r.client, name)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if id != "" {
		resp.Diagnostics.AddError(fmt.Sprintf("authentication_name %s already exists", name), "")

		return
	}
	jsonData := plan.toJSON()
	jsonData.ClientSecret = secret.ValueStringPointer()
	// Only on creation: the Bastion refuses redirect_uri on update.
	jsonData.RedirectURI = plan.RedirectURI.ValueString()
	body, headers, code, err := r.client.newRequestWithHeaders(ctx, "/externalauths/", http.MethodPost, jsonData)
	if err == nil && code != http.StatusOK && code != http.StatusNoContent {
		err = fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if id = headers.Get("X-Object-Id"); id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		if id, err = searchExternalAuthOIDC(ctx, r.client, name); err == nil && id == "" {
			err = fmt.Errorf("authentication_name %s not found after POST", name)
		}
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), "")

			return
		}
	}
	plan.ID = types.StringValue(id)
	r.readBack(ctx, plan, &resp.State, &resp.Diagnostics)
}

func (r *externalAuthOIDCResource) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state externalAuthOIDCModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !r.versionCheck(&resp.Diagnostics) {
		return
	}
	cfg, err := readExternalAuthOIDC(ctx, r.client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if cfg.ID == "" {
		resp.State.RemoveResource(ctx)

		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state.fromAPI(cfg))...)
}

func (r *externalAuthOIDCResource) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan, state externalAuthOIDCModel
	var secret types.String
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("client_secret_wo"), &secret)...)
	if resp.Diagnostics.HasError() || !r.versionCheck(&resp.Diagnostics) {
		return
	}
	jsonData := plan.toJSON()
	switch {
	case !secret.IsNull():
		jsonData.ClientSecret = secret.ValueStringPointer()
	case !state.ClientSecretWOVersion.IsNull() && plan.ClientSecretWOVersion.IsNull():
		// The secret and its version left the configuration: an empty secret removes it.
		jsonData.ClientSecret = new(string)
	}
	body, code, err := r.client.newRequest(ctx, "/externalauths/"+state.ID.ValueString(), http.MethodPut, jsonData)
	if err == nil && code != http.StatusOK && code != http.StatusNoContent {
		err = fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	plan.ID = state.ID
	r.readBack(ctx, plan, &resp.State, &resp.Diagnostics)
}

func (r *externalAuthOIDCResource) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state externalAuthOIDCModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !r.versionCheck(&resp.Diagnostics) {
		return
	}
	body, code, err := r.client.newRequest(ctx, "/externalauths/"+state.ID.ValueString(), http.MethodDelete, nil)
	if err == nil && !deleteDone(code, body) {
		err = fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
	}
}

// ImportState takes the authentication name as id, then Read fills the state.
func (r *externalAuthOIDCResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	if !r.versionCheck(&resp.Diagnostics) {
		return
	}
	id, err := searchExternalAuthOIDC(ctx, r.client, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if id == "" {
		resp.Diagnostics.AddError(
			fmt.Sprintf("don't find authentication_name with id %s (id must be <authentication_name>)", req.ID), "")

		return
	}
	cfg, err := readExternalAuthOIDC(ctx, r.client, id)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if cfg.Type != typeOIDC {
		resp.Diagnostics.AddError(fmt.Sprintf("external authentication %s is %s, not OIDC", req.ID, cfg.Type), "")

		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *externalAuthOIDCResource) versionCheck(diags *fwdiag.Diagnostics) bool {
	if err := oidcVersionCheck("resource", "wallix-bastion_externalauth_oidc", r.client.bastionAPIVersion); err != nil {
		diags.AddError(err.Error(), "")

		return false
	}

	return true
}

// readBack stores the external authentication of plan as the API returns it after a write.
func (r *externalAuthOIDCResource) readBack(
	ctx context.Context, plan externalAuthOIDCModel, state *tfsdk.State, diags *fwdiag.Diagnostics,
) {
	// Saved first, so that a failed read keeps the object in state; a state holds no unknown value.
	saved := plan
	if saved.RedirectURI.IsUnknown() {
		saved.RedirectURI = types.StringNull()
	}
	diags.Append(state.Set(ctx, saved)...)
	cfg, err := readExternalAuthOIDC(ctx, r.client, plan.ID.ValueString())
	if err != nil {
		diags.AddError(err.Error(), "")

		return
	}
	if cfg.ID == "" {
		diags.AddError(
			fmt.Sprintf("authentication_name %s not found after the write", plan.AuthenticationName.ValueString()), "")

		return
	}
	diags.Append(state.Set(ctx, plan.fromAPI(cfg))...)
}

// toJSON returns the body of a POST or PUT, without the secret: nil leaves it unchanged.
func (m externalAuthOIDCModel) toJSON() jsonExternalAuthOIDC {
	jsonData := jsonExternalAuthOIDC{
		AuthenticationName: m.AuthenticationName.ValueString(),
		Type:               typeOIDC,
		Description:        m.Description.ValueString(),
		Timeout:            m.Timeout.ValueFloat64(),
		ClientID:           m.ClientID.ValueString(),
		DiscoveryURL:       m.DiscoveryURL.ValueString(),
	}
	if c := m.ClaimCustomization; c != nil {
		jsonData.ClaimCustomization = jsonExternalAuthOIDCClaims{
			Username:    c.Username.ValueString(),
			Group:       c.Group.ValueString(),
			Email:       c.Email.ValueString(),
			Displayname: c.Displayname.ValueString(),
			Language:    c.Language.ValueString(),
		}
	}
	if a := m.OIDCAdvancedAttributes; a != nil {
		jsonData.OIDCAdvancedAttributes = &jsonExternalAuthOIDCAdvanced{
			Issuer:                           a.Issuer.ValueString(),
			Scope:                            nonNil(a.Scope),
			AuthorizationEndpoint:            a.AuthorizationEndpoint.ValueString(),
			TokenEndpoint:                    a.TokenEndpoint.ValueString(),
			UserinfoURLEndpoint:              a.UserinfoURLEndpoint.ValueString(),
			GrantType:                        a.GrantType.ValueString(),
			ResponseType:                     a.ResponseType.ValueString(),
			JwksURI:                          a.JwksURI.ValueString(),
			IDTokenSigningAlgValuesSupported: nonNil(a.IDTokenSigningAlgValuesSupported),
		}
	}

	return jsonData
}

// fromAPI returns the state of the external authentication the API returned, m being the prior
// state or the plan. The secret is never read back: the Bastion masks it.
func (m externalAuthOIDCModel) fromAPI(cfg jsonExternalAuthOIDC) externalAuthOIDCModel {
	claims := m.ClaimCustomization
	if claims == nil {
		claims = &oidcClaimsModel{}
	}
	version := m.ClientSecretWOVersion
	if cfg.ClientSecret != nil && *cfg.ClientSecret == oidcNoClientSecret {
		// Removed outside Terraform: a null version plans the update that sends it again.
		version = types.Int64Null()
	}
	state := externalAuthOIDCModel{
		ID:                    m.ID,
		AuthenticationName:    types.StringValue(cfg.AuthenticationName),
		Description:           optionalString(cfg.Description, m.Description, true),
		Timeout:               types.Float64Value(cfg.Timeout),
		ClientID:              types.StringValue(cfg.ClientID),
		ClientSecretWO:        types.StringNull(),
		ClientSecretWOVersion: version,
		DiscoveryURL:          optionalString(cfg.DiscoveryURL, m.DiscoveryURL, true),
		RedirectURI:           optionalString(cfg.RedirectURI, types.StringNull(), true),
		ClaimCustomization: &oidcClaimsModel{
			Username:    types.StringValue(cfg.ClaimCustomization.Username),
			Group:       optionalString(cfg.ClaimCustomization.Group, claims.Group, true),
			Email:       optionalString(cfg.ClaimCustomization.Email, claims.Email, true),
			Displayname: optionalString(cfg.ClaimCustomization.Displayname, claims.Displayname, true),
			Language:    optionalString(cfg.ClaimCustomization.Language, claims.Language, true),
		},
	}
	if a := cfg.OIDCAdvancedAttributes; a != nil {
		prior := m.OIDCAdvancedAttributes
		if prior == nil {
			prior = &oidcAdvancedModel{}
		}
		state.OIDCAdvancedAttributes = &oidcAdvancedModel{
			Issuer:                           types.StringValue(a.Issuer),
			Scope:                            nonNil(a.Scope),
			AuthorizationEndpoint:            types.StringValue(a.AuthorizationEndpoint),
			TokenEndpoint:                    types.StringValue(a.TokenEndpoint),
			UserinfoURLEndpoint:              optionalString(a.UserinfoURLEndpoint, prior.UserinfoURLEndpoint, true),
			GrantType:                        types.StringValue(a.GrantType),
			ResponseType:                     types.StringValue(a.ResponseType),
			JwksURI:                          types.StringValue(a.JwksURI),
			IDTokenSigningAlgValuesSupported: nonNil(a.IDTokenSigningAlgValuesSupported),
		}
	}

	return state
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}

	return s
}

// searchExternalAuthOIDC returns the id of the external authentication named name, "" if none.
func searchExternalAuthOIDC(ctx context.Context, c *Client, name string) (string, error) {
	// No fields parameter: the Bastion answers 500 to fields= on /externalauths (12.4.3).
	body, code, err := c.newRequest(ctx, "/externalauths/?"+searchQuery(skAuthenticationName, name), http.MethodGet, nil)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK {
		return "", fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonExternalAuthOIDC
	if err := json.Unmarshal([]byte(body), &results); err != nil {
		return "", fmt.Errorf("unmarshaling json: %w", err)
	}
	for _, r := range results {
		if r.AuthenticationName == name {
			return r.ID, nil
		}
	}

	return "", nil
}

// readExternalAuthOIDC returns the external authentication of id, with an empty ID if not found.
func readExternalAuthOIDC(ctx context.Context, c *Client, id string) (jsonExternalAuthOIDC, error) {
	var result jsonExternalAuthOIDC
	body, code, err := c.newRequest(ctx, "/externalauths/"+id, http.MethodGet, nil)
	if err != nil {
		return result, err
	}
	if code == http.StatusNotFound {
		return result, nil
	}
	if code != http.StatusOK {
		return result, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return result, fmt.Errorf("unmarshaling json: %w", err)
	}

	return result, nil
}
