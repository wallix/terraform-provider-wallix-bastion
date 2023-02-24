package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	balt "github.com/jeremmfr/go-utils/basicalter"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
)

type jsonConfigOptions struct {
	ID         string        `json:"id"`
	ConfigName string        `json:"config_name"`
	Name       string        `json:"name"`
	Date       string        `json:"date"`
	Options    []interface{} `json:"options"`
}

func dataSourceConfigoption() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConfigoptionRead,
		Schema: map[string]*schema.Schema{
			"config_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"options_list": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"config_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceConfigoptionVersionCheck(version string) error {
	if bchk.InSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_configoption not available with api version %s", version)
}

func dataSourceConfigoptionRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceConfigoptionVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConfigoption(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillConfigoption(d, cfg)
	d.SetId(cfg.ID)

	return nil
}

func readConfigoption(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	jsonConfigOptions, error,
) {
	c := m.(*Client)
	var result jsonConfigOptions
	var params string
	if optionsList := d.Get("options_list").(*schema.Set).List(); len(optionsList) > 0 {
		params += "?options="
		for _, v := range optionsList {
			params += v.(string) + ","
		}
		_ = balt.CutPrefixInString(&params, ",")
	}
	body, code, err := c.newRequest(ctx, "/configoptions/"+d.Get("config_id").(string)+params, http.MethodGet, nil)
	if err != nil {
		return result, err
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

func fillConfigoption(d *schema.ResourceData, jsonData jsonConfigOptions) {
	if tfErr := d.Set("config_name", jsonData.ConfigName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("name", jsonData.Name); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("date", jsonData.Date); tfErr != nil {
		panic(tfErr)
	}
	options := make([]string, len(jsonData.Options))
	for i, v := range jsonData.Options {
		b, err := json.Marshal(v)
		if err != nil {
			continue
		}
		options[i] = string(b)
	}
	if tfErr := d.Set("options", options); tfErr != nil {
		panic(tfErr)
	}
}
