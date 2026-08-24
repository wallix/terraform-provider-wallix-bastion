package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTimeframe() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTimeframeRead,
		Schema: map[string]*schema.Schema{
			"timeframe_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_overtimable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"periods": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"start_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"week_days": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceTimeframeVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_timeframe not available with api version %s", version)
}

func dataSourceTimeframeRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceTimeframeVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	timeframeName := d.Get("timeframe_name").(string)
	ex, err := checkResourceTimeframeExits(ctx, timeframeName, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("timeframe_name %s doesn't exists", timeframeName))
	}
	cfg, err := readTimeframeOptions(ctx, timeframeName, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceTimeframe(d, cfg)
	d.SetId(timeframeName)

	return nil
}

func fillSourceTimeframe(d *schema.ResourceData, jsonData jsonTimeframe) {
	if tfErr := d.Set("timeframe_name", jsonData.TimeframeName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_overtimable", jsonData.IsOvertimable); tfErr != nil {
		panic(tfErr)
	}
	periods := make([]map[string]interface{}, len(jsonData.Periods))
	for i, v := range jsonData.Periods {
		periods[i] = map[string]interface{}{
			"start_date": v.StartDate,
			"end_date":   v.EndDate,
			"start_time": v.StartTime,
			"end_time":   v.EndTime,
			"week_days":  v.WeekDays,
		}
	}
	if tfErr := d.Set("periods", periods); tfErr != nil {
		panic(tfErr)
	}
}
