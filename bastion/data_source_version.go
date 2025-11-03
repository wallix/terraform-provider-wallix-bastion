package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type jsonVersion struct {
	Version                 string  `json:"version"`
	VersionDecimal          float64 `json:"version_decimal"`
	WABVersion              string  `json:"wab_version"`
	WABVersionDecimal       float64 `json:"wab_version_decimal"`
	WABVersioHotfix         string  `json:"wab_version_hotfix"`
	WABVersionHotfixDecimal float64 `json:"wab_version_hotfix_decimal"`
	WABCompleteVersion      string  `json:"wab_complete_version"`
}

func dataSourceVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVersionRead,
		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_decimal": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wab_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wab_version_decimal": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wab_version_hotfix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wab_version_hotfix_decimal": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wab_complete_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVersionRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	cfg, err := readVersionOptions(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceVersion(d, cfg)
	d.SetId("version")

	return nil
}

func readVersionOptions(
	ctx context.Context, m interface{},
) (
	jsonVersion, error,
) {
	c := m.(*Client)
	var result jsonVersion

	body, code, err := c.newRequest(ctx, "/version", http.MethodGet, nil)
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

func fillSourceVersion(d *schema.ResourceData, jsonData jsonVersion) {
	if tfErr := d.Set("version", jsonData.Version); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("version_decimal", fmt.Sprintf("%f", jsonData.VersionDecimal)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("wab_version", jsonData.WABVersion); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("wab_version_decimal", fmt.Sprintf("%f", jsonData.WABVersionDecimal)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("wab_version_hotfix", jsonData.WABVersioHotfix); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("wab_version_hotfix_decimal", fmt.Sprintf("%f", jsonData.WABVersionHotfixDecimal)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("wab_complete_version", jsonData.WABCompleteVersion); tfErr != nil {
		panic(tfErr)
	}
}
