package bastion

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

const providerTypeName = "wallix-bastion"

var _ provider.Provider = (*frameworkProvider)(nil)

// frameworkProvider serves the resources migrated to terraform-plugin-framework, next to the SDKv2
// provider in the mux: both schemas must stay identical.
type frameworkProvider struct{}

type frameworkProviderModel struct {
	IP                 types.String `tfsdk:"ip"`
	User               types.String `tfsdk:"user"`
	Port               types.Int64  `tfsdk:"port"`
	Token              types.String `tfsdk:"token"`
	Password           types.String `tfsdk:"password"`
	APIVersion         types.String `tfsdk:"api_version"`
	SessionTimeout     types.Int64  `tfsdk:"session_timeout"`
	CSRFEnabled        types.Bool   `tfsdk:"csrf_enabled"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

func (p *frameworkProvider) Metadata(
	_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse,
) {
	resp.TypeName = providerTypeName
}

func (p *frameworkProvider) Schema(
	_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse,
) {
	// SDKv2 exports a Required argument as Optional when the env variable of its DefaultFunc is set.
	ipFromEnv, userFromEnv := os.Getenv(envBastionHost) != "", os.Getenv(envBastionUser) != ""
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ip":          schema.StringAttribute{Required: !ipFromEnv, Optional: ipFromEnv},
			skUser:        schema.StringAttribute{Required: !userFromEnv, Optional: userFromEnv},
			skPort:        schema.Int64Attribute{Optional: true},
			"token":       schema.StringAttribute{Optional: true},
			skPassword:    schema.StringAttribute{Optional: true},
			"api_version": schema.StringAttribute{Optional: true},
			"session_timeout": schema.Int64Attribute{
				Optional:           true,
				Description:        descSessionTimeout,
				DeprecationMessage: deprecationSessionTimeout,
			},
			"csrf_enabled":         schema.BoolAttribute{Optional: true, Description: descCSRFEnabled},
			"insecure_skip_verify": schema.BoolAttribute{Optional: true, Description: descInsecureSkipVerify},
		},
	}
}

func (p *frameworkProvider) Configure(
	ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse,
) {
	var data frameworkProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	config, err := data.config()
	if err != nil {
		resp.Diagnostics.AddError("Invalid provider configuration", err.Error())

		return
	}
	client, diags := sharedClient(config)
	resp.Diagnostics.Append(frameworkDiags(diags)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *frameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &timeframeResource{} },
	}
}

func (p *frameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &timeframeDataSource{} },
	}
}

// config applies the env variables and defaults of the SDKv2 schema to the arguments left null.
func (m frameworkProviderModel) config() (Config, error) {
	port, err := intOrEnv(m.Port, envBastionPort, defaultBastionPort)
	if err != nil {
		return Config{}, err
	}
	sessionTimeout, err := intOrEnv(m.SessionTimeout, envSessionTimeout, 0)
	if err != nil {
		return Config{}, err
	}
	csrfEnabled, err := boolOrEnv(m.CSRFEnabled, envCSRFEnabled, true)
	if err != nil {
		return Config{}, err
	}
	insecureSkipVerify, err := boolOrEnv(m.InsecureSkipVerify, envInsecureSkipVerify, false)
	if err != nil {
		return Config{}, err
	}

	return Config{
		BastionAPIVersion:  stringOrEnv(m.APIVersion, envBastionAPIVersion, VersionWallixAPI312),
		BastionIP:          stringOrEnv(m.IP, envBastionHost, ""),
		BastionPort:        port,
		BastionToken:       stringOrEnv(m.Token, envBastionToken, ""),
		BastionUser:        stringOrEnv(m.User, envBastionUser, ""),
		BastionPwd:         stringOrEnv(m.Password, envBastionPassword, ""),
		SessionTimeout:     sessionTimeout,
		CSRFEnabled:        csrfEnabled,
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}

// stringOrEnv behaves like schema.EnvDefaultFunc: the env variable applies only to a null argument.
func stringOrEnv(v types.String, env, def string) string {
	if !v.IsNull() {
		return v.ValueString()
	}
	if e := os.Getenv(env); e != "" {
		return e
	}

	return def
}

func intOrEnv(v types.Int64, env string, def int) (int, error) {
	if !v.IsNull() {
		return int(v.ValueInt64()), nil
	}
	e := os.Getenv(env)
	if e == "" {
		return def, nil
	}
	// Same parsing as the SDKv2 field reader.
	i, err := strconv.ParseInt(e, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("parsing %s: %w", env, err)
	}

	return int(i), nil
}

func boolOrEnv(v types.Bool, env string, def bool) (bool, error) {
	if !v.IsNull() {
		return v.ValueBool(), nil
	}
	e := os.Getenv(env)
	if e == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(e)
	if err != nil {
		return false, fmt.Errorf("parsing %s: %w", env, err)
	}

	return b, nil
}

// providerClient returns the client of a configured provider, nil before ConfigureProvider.
func providerClient(data any, diags *fwdiag.Diagnostics) *Client {
	if data == nil {
		return nil
	}
	client, ok := data.(*Client)
	if !ok {
		diags.AddError("Unexpected provider data", fmt.Sprintf("expected *bastion.Client, got %T", data))
	}

	return client
}

// privateKeyFrameworkState marks a resource state written by the Framework: unlike SDKv2, which
// stored "" and false for an unset Optional argument, its zero values come from the configuration.
const privateKeyFrameworkState = "framework_state"

type privateState interface {
	GetKey(ctx context.Context, key string) ([]byte, fwdiag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) fwdiag.Diagnostics
}

func frameworkState(ctx context.Context, private privateState) (bool, fwdiag.Diagnostics) {
	value, diags := private.GetKey(ctx, privateKeyFrameworkState)

	return value != nil, diags
}

func markFrameworkState(ctx context.Context, private privateState) fwdiag.Diagnostics {
	return private.SetKey(ctx, privateKeyFrameworkState, []byte("true"))
}

// optionalString reads back an Optional, non-Computed argument: the API returns "" when it is
// unset, kept as "" only when a Framework state shows that the configuration sets it.
func optionalString(api string, prior types.String, fromFramework bool) types.String {
	if api == "" && (!fromFramework || prior.IsNull()) {
		return types.StringNull()
	}

	return types.StringValue(api)
}

// optionalBool applies the rule of optionalString to false.
func optionalBool(api bool, prior types.Bool, fromFramework bool) types.Bool {
	if !api && (!fromFramework || prior.IsNull()) {
		return types.BoolNull()
	}

	return types.BoolValue(api)
}

// frameworkDiags converts the SDKv2 diagnostics of the code shared by both servers.
func frameworkDiags(diags diag.Diagnostics) fwdiag.Diagnostics {
	var out fwdiag.Diagnostics
	for _, d := range diags {
		if d.Severity == diag.Error {
			out.AddError(d.Summary, d.Detail)
		} else {
			out.AddWarning(d.Summary, d.Detail)
		}
	}

	return out
}
