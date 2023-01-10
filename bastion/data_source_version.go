package bastion

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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

func dataSourceVersionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cfg, err := readVersionOptions(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceVersion(d, cfg)
	d.SetId("version")

	return nil
}

func readVersionOptions(ctx context.Context, m interface{}) (jsonVersion, error) {
	c := m.(*Client)
	var result jsonVersion
	url := "https://" + c.bastionIP + ":" + strconv.Itoa(c.bastionPort) + "/api/version"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("X-Auth-Key", c.bastionToken)
	req.Header.Add("X-Auth-User", c.bastionUser)
	if err != nil {
		return result, err
	}
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, //nolint: gosec
		DisableKeepAlives: true,
	}
	httpClient := &http.Client{Transport: tr}
	resp, err := httpClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("api doesn't return OK : %d with body :\n%s", resp.StatusCode, string(respBody))
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return result, fmt.Errorf("json.Unmarshal failed : %w", err)
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
