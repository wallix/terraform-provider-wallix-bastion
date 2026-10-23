package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type jsonTimeframe struct {
	IsOvertimable bool                  `json:"is_overtimable"`
	TimeframeName string                `json:"timeframe_name"`
	Description   string                `json:"description"`
	Periods       []jsonTimeFramePeriod `json:"periods"`
}

type jsonTimeFramePeriod struct {
	StartDate string   `json:"start_date"`
	EndDate   string   `json:"end_date"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	WeekDays  []string `json:"week_days"`
}

var (
	_ resource.ResourceWithConfigure   = (*timeframeResource)(nil)
	_ resource.ResourceWithImportState = (*timeframeResource)(nil)
)

type timeframeResource struct {
	client *Client
}

type timeframeModel struct {
	ID            types.String           `tfsdk:"id"`
	TimeframeName types.String           `tfsdk:"timeframe_name"`
	Description   types.String           `tfsdk:"description"`
	IsOvertimable types.Bool             `tfsdk:"is_overtimable"`
	Periods       []timeframePeriodModel `tfsdk:"periods"`
}

type timeframePeriodModel struct {
	StartDate string   `tfsdk:"start_date"`
	EndDate   string   `tfsdk:"end_date"`
	StartTime string   `tfsdk:"start_time"`
	EndTime   string   `tfsdk:"end_time"`
	WeekDays  []string `tfsdk:"week_days"`
}

func (r *timeframeResource) Metadata(
	_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_timeframe"
}

// Schema is the SDKv2 one as Terraform sees it: id Optional and Computed, periods a set of blocks.
func (r *timeframeResource) Schema(
	_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	date := stringvalidator.RegexMatches(regexp.MustCompile(
		`^([12]\d{3}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01]))$`), "Must respect the format `yyyy-mm-dd`")
	hour := stringvalidator.RegexMatches(regexp.MustCompile(
		`^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`), "Must respect the format `hh:mm`")
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{idFromName{}},
			},
			"timeframe_name": schema.StringAttribute{Required: true},
			skDescription:    schema.StringAttribute{Optional: true},
			"is_overtimable": schema.BoolAttribute{Optional: true},
		},
		Blocks: map[string]schema.Block{
			"periods": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"start_date": schema.StringAttribute{Required: true, Validators: []validator.String{date}},
						"end_date":   schema.StringAttribute{Required: true, Validators: []validator.String{date}},
						"start_time": schema.StringAttribute{Required: true, Validators: []validator.String{hour}},
						"end_time":   schema.StringAttribute{Required: true, Validators: []validator.String{hour}},
						// The API requires at least one day, in lower case.
						"week_days": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								setvalidator.ValueStringsAre(stringvalidator.OneOf(
									"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
								)),
							},
						},
					},
				},
			},
		},
	}
}

// idFromName plans id as timeframe_name: the API identifies a timeframe by its name and renames
// it in place.
type idFromName struct{}

func (m idFromName) Description(_ context.Context) string {
	return "id follows timeframe_name"
}

func (m idFromName) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m idFromName) PlanModifyString(
	ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse,
) {
	if !req.ConfigValue.IsNull() {
		return
	}
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("timeframe_name"), &resp.PlanValue)...)
}

func (r *timeframeResource) Configure(
	_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse,
) {
	r.client = providerClient(req.ProviderData, &resp.Diagnostics)
}

func resourceTimeframeVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_timeframe not available with api version %s", version)
}

func (r *timeframeResource) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var plan timeframeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := resourceTimeframeVersionCheck(r.client.bastionAPIVersion); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	name := plan.TimeframeName.ValueString()
	ex, err := checkResourceTimeframeExits(ctx, name, r.client)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if ex {
		resp.Diagnostics.AddError(fmt.Sprintf("timeframe_name %s already exists", name), "")

		return
	}
	if err := addTimeframe(ctx, r.client, plan.toJSON()); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	plan.ID = plan.TimeframeName
	r.readBack(ctx, plan, &resp.State, resp.Private, &resp.Diagnostics)
}

func (r *timeframeResource) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var state timeframeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := resourceTimeframeVersionCheck(r.client.bastionAPIVersion); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	cfg, err := readTimeframeOptions(ctx, state.ID.ValueString(), r.client)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if cfg.TimeframeName == "" {
		resp.State.RemoveResource(ctx)

		return
	}
	fromFramework, diags := frameworkState(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state.fromAPI(cfg, fromFramework))...)
	resp.Diagnostics.Append(markFrameworkState(ctx, resp.Private)...)
}

func (r *timeframeResource) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var plan, state timeframeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := resourceTimeframeVersionCheck(r.client.bastionAPIVersion); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if err := updateTimeframe(ctx, r.client, state.ID.ValueString(), plan.toJSON()); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	// The API renames the timeframe: follow the new name, or the next read drops it from state.
	plan.ID = plan.TimeframeName
	r.readBack(ctx, plan, &resp.State, resp.Private, &resp.Diagnostics)
}

func (r *timeframeResource) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var state timeframeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := resourceTimeframeVersionCheck(r.client.bastionAPIVersion); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if err := deleteTimeframe(ctx, r.client, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
	}
}

// ImportState takes the timeframe name as id, then Read fills the state.
func (r *timeframeResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	if err := resourceTimeframeVersionCheck(r.client.bastionAPIVersion); err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	cfg, err := readTimeframeOptions(ctx, req.ID, r.client)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")

		return
	}
	if cfg.TimeframeName == "" {
		resp.Diagnostics.AddError(
			fmt.Sprintf("don't find timeframe_name with id %s (id must be <timeframe_name>)", req.ID), "")

		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), cfg.TimeframeName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("timeframe_name"), cfg.TimeframeName)...)
}

// readBack stores the timeframe of plan as the API returns it after a write.
func (r *timeframeResource) readBack(
	ctx context.Context, plan timeframeModel, state *tfsdk.State, private privateState, diags *fwdiag.Diagnostics,
) {
	// Saved first, so that a failed read keeps the object in state.
	diags.Append(state.Set(ctx, plan)...)
	diags.Append(markFrameworkState(ctx, private)...)
	cfg, err := readTimeframeOptions(ctx, plan.ID.ValueString(), r.client)
	if err != nil {
		diags.AddError(err.Error(), "")

		return
	}
	if cfg.TimeframeName == "" {
		diags.AddError(fmt.Sprintf("timeframe_name %s not found after the write", plan.ID.ValueString()), "")

		return
	}
	diags.Append(state.Set(ctx, plan.fromAPI(cfg, true))...)
}

func (m timeframeModel) toJSON() jsonTimeframe {
	jsonData := jsonTimeframe{
		Description:   m.Description.ValueString(),
		IsOvertimable: m.IsOvertimable.ValueBool(),
		TimeframeName: m.TimeframeName.ValueString(),
		Periods:       make([]jsonTimeFramePeriod, len(m.Periods)),
	}
	for i, p := range m.Periods {
		jsonData.Periods[i] = jsonTimeFramePeriod(p)
	}

	return jsonData
}

// fromAPI returns the state of the timeframe the API returned, m being the prior state or the plan.
func (m timeframeModel) fromAPI(cfg jsonTimeframe, fromFramework bool) timeframeModel {
	// Empty, not nil: no periods block is an empty set in the configuration.
	periods := make([]timeframePeriodModel, len(cfg.Periods))
	for i, p := range cfg.Periods {
		periods[i] = timeframePeriodModel(p)
	}

	return timeframeModel{
		ID:            m.ID,
		TimeframeName: types.StringValue(cfg.TimeframeName),
		Description:   optionalString(cfg.Description, m.Description, fromFramework),
		IsOvertimable: optionalBool(cfg.IsOvertimable, m.IsOvertimable, fromFramework),
		Periods:       periods,
	}
}

func checkResourceTimeframeExits(
	ctx context.Context, timeframeName string, c *Client,
) (
	bool, error,
) {
	body, code, err := c.newRequest(ctx, "/timeframes/"+url.PathEscape(timeframeName), http.MethodGet, nil)
	if err != nil {
		return false, err
	}
	if code == http.StatusNotFound {
		return false, nil
	}
	if code != http.StatusOK {
		return false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}

	return true, nil
}

func addTimeframe(
	ctx context.Context, c *Client, jsonData jsonTimeframe,
) error {
	body, code, err := c.newRequest(ctx, "/timeframes/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateTimeframe(
	ctx context.Context, c *Client, timeframeID string, jsonData jsonTimeframe,
) error {
	body, code, err := c.newRequest(ctx,
		"/timeframes/"+url.PathEscape(timeframeID)+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteTimeframe(
	ctx context.Context, c *Client, timeframeID string,
) error {
	body, code, err := c.newRequest(ctx, "/timeframes/"+url.PathEscape(timeframeID), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if !deleteDone(code, body) {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func readTimeframeOptions(
	ctx context.Context, timeframeID string, c *Client,
) (
	jsonTimeframe, error,
) {
	var result jsonTimeframe
	body, code, err := c.newRequest(ctx, "/timeframes/"+url.PathEscape(timeframeID), http.MethodGet, nil)
	if err != nil {
		return result, err
	}
	if code == http.StatusNotFound {
		return result, nil
	}
	if code != http.StatusOK {
		return result, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("unmarshaling json: %w", err)
	}

	return result, nil
}
