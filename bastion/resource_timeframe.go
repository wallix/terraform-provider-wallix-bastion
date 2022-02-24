package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
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

func resourceTimeframe() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTimeframeCreate,
		ReadContext:   resourceTimeframeRead,
		UpdateContext: resourceTimeframeUpdate,
		DeleteContext: resourceTimeframeDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTimeframeImport,
		},
		Schema: map[string]*schema.Schema{
			"timeframe_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_overtimable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"periods": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_date": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(
								`^([12]\d{3}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01]))$`), "Must respect the format `yyyy-mm-dd`"),
						},
						"end_date": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(
								`^([12]\d{3}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01]))$`), "Must respect the format `yyyy-mm-dd`"),
						},
						"start_time": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(
								`^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`), "Must respect the format `hh:mm`"),
						},
						"end_time": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(
								`^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`), "Must respect the format `hh:mm`"),
						},
						"week_days": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}
func resourceTimeframeVersionCheck(version string) error {
	if bchk.StringInSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_timeframe not validate with api version %s", version)
}

func resourceTimeframeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceTimeframeVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	ex, err := checkResourceTimeframeExits(ctx, d.Get("timeframe_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("timeframe_name %s already exists", d.Get("timeframe_name").(string)))
	}
	err = addTimeframe(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	ex, err = checkResourceTimeframeExits(ctx, d.Get("timeframe_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("timeframe_name %s can't find after POST", d.Get("timeframe_name").(string)))
	}
	d.SetId(d.Get("timeframe_name").(string))

	return resourceTimeframeRead(ctx, d, m)
}
func resourceTimeframeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceTimeframeVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readTimeframeOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.TimeframeName == "" {
		d.SetId("")
	} else {
		fillTimeframe(d, cfg)
	}

	return nil
}
func resourceTimeframeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceTimeframeVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateTimeframe(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceTimeframeRead(ctx, d, m)
}
func resourceTimeframeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceTimeframeVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteTimeframe(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceTimeframeImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceTimeframeVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	ex, err := checkResourceTimeframeExits(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find timeframe_name with id %s (id must be <timeframe_name>", d.Id())
	}
	cfg, err := readTimeframeOptions(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	fillTimeframe(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(cfg.TimeframeName)
	result[0] = d

	return result, nil
}

func checkResourceTimeframeExits(ctx context.Context, timeframeName string, m interface{}) (bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/timeframes/"+timeframeName, http.MethodGet, nil)
	if err != nil {
		return false, err
	}
	if code == http.StatusNotFound {
		return false, nil
	}
	if code != http.StatusOK {
		return false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}

	return true, nil
}

func addTimeframe(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData, err := prepareTimeframeJSON(d)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx, "/timeframes/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateTimeframe(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData, err := prepareTimeframeJSON(d)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx, "/timeframes/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteTimeframe(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/timeframes/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareTimeframeJSON(d *schema.ResourceData) (jsonTimeframe, error) {
	var jsonData jsonTimeframe
	jsonData.TimeframeName = d.Get("timeframe_name").(string)
	jsonData.Description = d.Get("description").(string)
	jsonData.IsOvertimable = d.Get("is_overtimable").(bool)
	if v := d.Get("periods").(*schema.Set).List(); len(v) > 0 {
		for _, v2 := range v {
			period := v2.(map[string]interface{})
			jsonPeriod := jsonTimeFramePeriod{
				StartDate: period["start_date"].(string),
				EndDate:   period["end_date"].(string),
				StartTime: period["start_time"].(string),
				EndTime:   period["end_time"].(string),
			}
			for _, d := range period["week_days"].(*schema.Set).List() {
				if !bchk.StringInSlice(d.(string), []string{
					"monday",
					"tuesday",
					"wednesday",
					"thursday",
					"friday",
					"saturday",
					"sunday",
				}) {
					return jsonData, fmt.Errorf("`%s` isn't a valid week_day", d.(string))
				}
				jsonPeriod.WeekDays = append(jsonPeriod.WeekDays, d.(string))
			}
			jsonData.Periods = append(jsonData.Periods, jsonPeriod)
		}
	} else {
		jsonData.Periods = make([]jsonTimeFramePeriod, 0)
	}

	return jsonData, nil
}

func readTimeframeOptions(
	ctx context.Context, timeframeID string, m interface{}) (jsonTimeframe, error) {
	c := m.(*Client)
	var result jsonTimeframe
	body, code, err := c.newRequest(ctx, "/timeframes/"+timeframeID, http.MethodGet, nil)
	if err != nil {
		return result, err
	}
	if code == http.StatusNotFound {
		return result, nil
	}
	if code != http.StatusOK {
		return result, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("json.Unmarshal failed : %w", err)
	}

	return result, nil
}

func fillTimeframe(d *schema.ResourceData, jsonData jsonTimeframe) {
	if tfErr := d.Set("timeframe_name", jsonData.TimeframeName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_overtimable", jsonData.IsOvertimable); tfErr != nil {
		panic(tfErr)
	}
	periods := make([]map[string]interface{}, 0)
	for _, v := range jsonData.Periods {
		periods = append(periods, map[string]interface{}{
			"start_date": v.StartDate,
			"end_date":   v.EndDate,
			"start_time": v.StartTime,
			"end_time":   v.EndTime,
			"week_days":  v.WeekDays,
		})
	}
	if tfErr := d.Set("periods", periods); tfErr != nil {
		panic(tfErr)
	}
}
