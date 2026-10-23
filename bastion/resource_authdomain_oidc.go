package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type jsonAuthDomainOIDC struct {
	ID                 string   `json:"id,omitempty"`
	DomainName         string   `json:"domain_name"`
	Type               string   `json:"type"`
	Description        string   `json:"description"`
	IsDefault          bool     `json:"is_default"`
	AuthDomainName     string   `json:"auth_domain_name"`
	ExternalAuths      []string `json:"external_auths"`
	SecondaryAuth      []string `json:"secondary_auth"`
	DefaultLanguage    string   `json:"default_language"`
	DefaultEmailDomain string   `json:"default_email_domain"`
	Label              string   `json:"label"`
}

var (
	_ resource.ResourceWithConfigure   = (*authDomainOIDCResource)(nil)
	_ resource.ResourceWithImportState = (*authDomainOIDCResource)(nil)
)

type authDomainOIDCResource struct {
	client *Client
}

type authDomainOIDCModel struct {
	ID                 types.String `tfsdk:"id"`
	DomainName         types.String `tfsdk:"domain_name"`
	AuthDomainName     types.String `tfsdk:"auth_domain_name"`
	DefaultEmailDomain types.String `tfsdk:"default_email_domain"`
	DefaultLanguage    types.String `tfsdk:"default_language"`
	ExternalAuths      []string     `tfsdk:"external_auths"`
	Label              types.String `tfsdk:"label"`
	Description        types.String `tfsdk:"description"`
	IsDefault          types.Bool   `tfsdk:"is_default"`
	SecondaryAuth      []string     `tfsdk:"secondary_auth"`
}

func (r *authDomainOIDCResource) Metadata(
	_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_authdomain_oidc"
}

func (r *authDomainOIDCResource) Schema(
	_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "OIDC authentication domain (API v3.12 and later).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "ID of the authentication domain on the Bastion.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			skDomainName: schema.StringAttribute{
				Required:    true,
				Description: "Domain name: letters, digits, dots and hyphens.",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 63)},
			},
			skAuthDomainName: schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			skDefaultEmailDomain: schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			skDefaultLanguage: schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{stringvalidator.OneOf("de", "en", "es", "fr", "ru")},
			},
			skExternalAuths: schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Names of the OIDC external authentications of the domain.",
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
			},
			skLabel: schema.StringAttribute{
				Required:    true,
				Description: "Label shown on the login page.",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			skDescription: schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{stringvalidator.LengthAtMost(1024)},
			},
			skIsDefault: schema.BoolAttribute{Optional: true},
			skSecondaryAuth: schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Names of the secondary authentication methods.",
			},
		},
	}
}

func (r *authDomainOIDCResource) Configure(
	_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse,
) {
	r.client = providerClient(req.ProviderData, &resp.Diagnostics)
}

func (r *authDomainOIDCResource) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan authDomainOIDCModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !r.versionCheck(&resp.Diagnostics) {
		return
	}
	name := plan.DomainName.ValueString()
	id, err := searchAuthDomainOIDC(ctx, r.client, name)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if id != "" {
		resp.Diagnostics.AddError(fmt.Sprintf("domain_name %s already exists", name), "")

		return
	}
	body, headers, code, err := r.client.newRequestWithHeaders(ctx, "/authdomains/", http.MethodPost, plan.toJSON())
	if err == nil && code != http.StatusOK && code != http.StatusNoContent {
		err = fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if id = headers.Get("X-Object-Id"); id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		if id, err = searchAuthDomainOIDC(ctx, r.client, name); err == nil && id == "" {
			err = fmt.Errorf("domain_name %s not found after POST", name)
		}
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), "")

			return
		}
	}
	plan.ID = types.StringValue(id)
	r.readBack(ctx, plan, &resp.State, &resp.Diagnostics)
}

func (r *authDomainOIDCResource) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state authDomainOIDCModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !r.versionCheck(&resp.Diagnostics) {
		return
	}
	cfg, err := readAuthDomainOIDC(ctx, r.client, state.ID.ValueString())
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

func (r *authDomainOIDCResource) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan, state authDomainOIDCModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !r.versionCheck(&resp.Diagnostics) {
		return
	}
	body, code, err := r.client.newRequest(ctx,
		"/authdomains/"+state.ID.ValueString()+"?force=true", http.MethodPut, plan.toJSON())
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

func (r *authDomainOIDCResource) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state authDomainOIDCModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !r.versionCheck(&resp.Diagnostics) {
		return
	}
	body, code, err := r.client.newRequest(ctx, "/authdomains/"+state.ID.ValueString(), http.MethodDelete, nil)
	if err == nil && !deleteDone(code, body) {
		err = fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
	}
}

// ImportState takes the domain name as id, then Read fills the state.
func (r *authDomainOIDCResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	if !r.versionCheck(&resp.Diagnostics) {
		return
	}
	id, err := searchAuthDomainOIDC(ctx, r.client, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if id == "" {
		resp.Diagnostics.AddError(
			fmt.Sprintf("don't find domain_name with id %s (id must be <domain_name>)", req.ID), "")

		return
	}
	cfg, err := readAuthDomainOIDC(ctx, r.client, id)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if cfg.Type != typeOIDC {
		resp.Diagnostics.AddError(fmt.Sprintf("authentication domain %s is %s, not OIDC", req.ID, cfg.Type), "")

		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *authDomainOIDCResource) versionCheck(diags *fwdiag.Diagnostics) bool {
	if err := oidcVersionCheck("resource", "wallix-bastion_authdomain_oidc", r.client.bastionAPIVersion); err != nil {
		diags.AddError(err.Error(), "")

		return false
	}

	return true
}

// readBack stores the domain of plan as the API returns it after a write.
func (r *authDomainOIDCResource) readBack(
	ctx context.Context, plan authDomainOIDCModel, state *tfsdk.State, diags *fwdiag.Diagnostics,
) {
	// Saved first, so that a failed read keeps the object in state.
	diags.Append(state.Set(ctx, plan)...)
	cfg, err := readAuthDomainOIDC(ctx, r.client, plan.ID.ValueString())
	if err != nil {
		diags.AddError(err.Error(), "")

		return
	}
	if cfg.ID == "" {
		diags.AddError(fmt.Sprintf("domain_name %s not found after the write", plan.DomainName.ValueString()), "")

		return
	}
	diags.Append(state.Set(ctx, plan.fromAPI(cfg))...)
}

func (m authDomainOIDCModel) toJSON() jsonAuthDomainOIDC {
	return jsonAuthDomainOIDC{
		DomainName:         m.DomainName.ValueString(),
		Type:               typeOIDC,
		Description:        m.Description.ValueString(),
		IsDefault:          m.IsDefault.ValueBool(),
		AuthDomainName:     m.AuthDomainName.ValueString(),
		ExternalAuths:      nonNil(m.ExternalAuths),
		SecondaryAuth:      nonNil(m.SecondaryAuth),
		DefaultLanguage:    m.DefaultLanguage.ValueString(),
		DefaultEmailDomain: m.DefaultEmailDomain.ValueString(),
		Label:              m.Label.ValueString(),
	}
}

// fromAPI returns the state of the domain the API returned, m being the prior state or the plan.
func (m authDomainOIDCModel) fromAPI(cfg jsonAuthDomainOIDC) authDomainOIDCModel {
	return authDomainOIDCModel{
		ID:                 m.ID,
		DomainName:         types.StringValue(cfg.DomainName),
		AuthDomainName:     types.StringValue(cfg.AuthDomainName),
		DefaultEmailDomain: types.StringValue(cfg.DefaultEmailDomain),
		DefaultLanguage:    types.StringValue(cfg.DefaultLanguage),
		ExternalAuths:      nonNil(cfg.ExternalAuths),
		Label:              types.StringValue(cfg.Label),
		Description:        optionalString(cfg.Description, m.Description, true),
		IsDefault:          optionalBool(cfg.IsDefault, m.IsDefault, true),
		SecondaryAuth:      optionalStrings(cfg.SecondaryAuth, m.SecondaryAuth),
	}
}

// optionalStrings applies the rule of optionalString to a list, nil being null.
func optionalStrings(api, prior []string) []string {
	if len(api) == 0 && prior == nil {
		return nil
	}

	return nonNil(api)
}

// searchAuthDomainOIDC returns the id of the authentication domain named name, "" if none.
func searchAuthDomainOIDC(ctx context.Context, c *Client, name string) (string, error) {
	body, code, err := c.newRequest(ctx, "/authdomains/?"+searchQuery(skDomainName, name), http.MethodGet, nil)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK {
		return "", fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonAuthDomainOIDC
	if err := json.Unmarshal([]byte(body), &results); err != nil {
		return "", fmt.Errorf("unmarshaling json: %w", err)
	}
	for _, r := range results {
		if r.DomainName == name {
			return r.ID, nil
		}
	}

	return "", nil
}

// readAuthDomainOIDC returns the domain of id, with an empty ID if not found.
func readAuthDomainOIDC(ctx context.Context, c *Client, id string) (jsonAuthDomainOIDC, error) {
	var result jsonAuthDomainOIDC
	body, code, err := c.newRequest(ctx, "/authdomains/"+id, http.MethodGet, nil)
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
