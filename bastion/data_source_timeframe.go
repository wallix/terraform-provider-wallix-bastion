package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = (*timeframeDataSource)(nil)

type timeframeDataSource struct {
	client *Client
}

type timeframeDataSourceModel struct {
	ID            types.String           `tfsdk:"id"`
	TimeframeName types.String           `tfsdk:"timeframe_name"`
	Description   types.String           `tfsdk:"description"`
	IsOvertimable types.Bool             `tfsdk:"is_overtimable"`
	Periods       []timeframePeriodModel `tfsdk:"periods"`
}

func (d *timeframeDataSource) Metadata(
	_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_timeframe"
}

// Schema is the SDKv2 one as Terraform sees it: SDKv2 exports a Computed-only set of
// *schema.Resource as an attribute (set of objects), not as a block.
func (d *timeframeDataSource) Schema(
	_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{Optional: true, Computed: true},
			"timeframe_name": schema.StringAttribute{Required: true},
			skDescription:    schema.StringAttribute{Computed: true},
			"is_overtimable": schema.BoolAttribute{Computed: true},
			"periods": schema.SetAttribute{
				Computed: true,
				ElementType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"start_date": types.StringType,
					"end_date":   types.StringType,
					"start_time": types.StringType,
					"end_time":   types.StringType,
					"week_days":  types.SetType{ElemType: types.StringType},
				}},
			},
		},
	}
}

func (d *timeframeDataSource) Configure(
	_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse,
) {
	d.client = providerClient(req.ProviderData, &resp.Diagnostics)
}

func dataSourceTimeframeVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_timeframe not available with api version %s", version)
}

func (d *timeframeDataSource) Read(
	ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse,
) {
	var data timeframeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := dataSourceTimeframeVersionCheck(d.client.bastionAPIVersion); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	timeframeName := data.TimeframeName.ValueString()
	cfg, err := readTimeframeOptions(ctx, timeframeName, d.client)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if cfg.TimeframeName == "" {
		resp.Diagnostics.AddError(fmt.Sprintf("timeframe_name %s doesn't exists", timeframeName), "")

		return
	}
	// Computed values: "" and false are stored as SDKv2 did.
	data.ID = types.StringValue(timeframeName)
	data.TimeframeName = types.StringValue(cfg.TimeframeName)
	data.Description = types.StringValue(cfg.Description)
	data.IsOvertimable = types.BoolValue(cfg.IsOvertimable)
	data.Periods = make([]timeframePeriodModel, len(cfg.Periods))
	for i, p := range cfg.Periods {
		data.Periods[i] = timeframePeriodModel(p)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
