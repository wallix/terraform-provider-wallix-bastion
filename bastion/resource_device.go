package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type jsonDevice struct {
	ID           string                   `json:"id,omitempty"`
	Alias        string                   `json:"alias"`
	Description  string                   `json:"description"`
	DeviceName   string                   `json:"device_name"`
	Host         string                   `json:"host"`
	LocalDomains *[]jsonDeviceLocalDomain `json:"local_domains,omitempty"`
	Services     *[]jsonDeviceService     `json:"services,omitempty"`
	Tags         *[]map[string]string     `json:"tags,omitempty"`
}

func resourceDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeviceCreate,
		ReadContext:   resourceDeviceRead,
		UpdateContext: resourceDeviceUpdate,
		DeleteContext: resourceDeviceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDeviceImport,
		},
		Schema: map[string]*schema.Schema{
			"device_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skHost: {
				Type:     schema.TypeString,
				Required: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"local_domains": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						skDomainName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skAdminAccount: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skCAPublicKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skEnablePasswordChange: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						skPasswordChangePolicy: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skPasswordChangePlugin: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skPasswordChangePluginParameters: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"services": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						skServiceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skConnectionPolicy: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skPort: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						skProtocol: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skGlobalDomains: {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						skSubprotocols: {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						skKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						skValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: schema.HashResource(&schema.Resource{
					Schema: map[string]*schema.Schema{
						skKey: {
							Type: schema.TypeString,
						},
						skValue: {
							Type: schema.TypeString,
						},
					},
				}),
			},
		},
	}
}

func resourceDeviceVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_device not available with api version %s", version)
}

func resourceDeviceCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceDevice(ctx, d.Get("device_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("device_name %s already exists", d.Get("device_name").(string)))
	}
	id, err := addDevice(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceDevice(ctx, d.Get("device_name").(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("device_name %s not found after POST", d.Get("device_name").(string)))
		}
	}
	d.SetId(id)

	return resourceDeviceRead(ctx, d, m)
}

func resourceDeviceRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDeviceOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillDevice(d, cfg)
	}

	return nil
}

func resourceDeviceUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDevice(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDeviceRead(ctx, d, m)
}

func resourceDeviceDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDevice(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDeviceImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceDevice(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find device_name with id %s (id must be <device_name>)", d.Id())
	}
	cfg, err := readDeviceOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillDevice(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceDevice(
	ctx context.Context, deviceName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/devices/?q=device_name="+deviceName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonDevice
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addDevice(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := prepareDeviceJSON(d)
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/devices/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateDevice(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareDeviceJSON(d)
	body, code, err := c.newRequest(ctx, "/devices/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteDevice(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/devices/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareDeviceJSON(d *schema.ResourceData) jsonDevice {
	jsonData := jsonDevice{
		DeviceName:  d.Get("device_name").(string),
		Host:        d.Get(skHost).(string),
		Alias:       d.Get("alias").(string),
		Description: d.Get(skDescription).(string),
	}

	if v, ok := d.GetOk("tags"); ok {
		tagsSet := v.(*schema.Set)
		tagsList := tagsSet.List()

		tags := make([]map[string]string, len(tagsList))

		for i, tagData := range tagsList {
			tagMap := tagData.(map[string]interface{})

			tags[i] = map[string]string{
				skKey:   tagMap[skKey].(string),
				skValue: tagMap[skValue].(string),
			}
		}
		jsonData.Tags = &tags
	}

	return jsonData
}

func readDeviceOptions(
	ctx context.Context, deviceID string, m interface{},
) (
	jsonDevice, error,
) {
	c := m.(*Client)
	var result jsonDevice
	body, code, err := c.newRequest(ctx, "/devices/"+deviceID, http.MethodGet, nil)
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

func fillDevice(d *schema.ResourceData, jsonData jsonDevice) {
	if tfErr := d.Set("device_name", jsonData.DeviceName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skHost, jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("alias", jsonData.Alias); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	localDomains := make([]map[string]interface{}, len(*jsonData.LocalDomains))
	for i, v := range *jsonData.LocalDomains {
		localDomains[i] = map[string]interface{}{
			"id":                   v.ID,
			skAdminAccount:         v.AdminAccount,
			skDomainName:           v.DomainName,
			skCAPublicKey:          v.CAPublicKey,
			skDescription:          v.Description,
			skEnablePasswordChange: v.EnablePasswordChange,
			skPasswordChangePolicy: v.PasswordChangePolicy,
			skPasswordChangePlugin: v.PasswordChangePlugin,
		}
		pluginParameters, _ := json.Marshal(v.PasswordChangePluginParameters) //nolint: errchkjson
		localDomains[i][skPasswordChangePluginParameters] = string(pluginParameters)
	}
	if tfErr := d.Set("local_domains", localDomains); tfErr != nil {
		panic(tfErr)
	}
	services := make([]map[string]interface{}, len(*jsonData.Services))
	for i, v := range *jsonData.Services {
		service := map[string]interface{}{
			"id":               v.ID,
			skServiceName:      v.ServiceName,
			skConnectionPolicy: v.ConnectionPolicy,
			skPort:             v.Port,
			skProtocol:         v.Protocol,
			skGlobalDomains:    make([]string, 0),
			skSubprotocols:     make([]string, 0),
		}
		if v.GlobalDomains != nil {
			service[skGlobalDomains] = make(([]string), len(*v.GlobalDomains))
			copy(service[skGlobalDomains].([]string), *v.GlobalDomains)
		}
		if v.SubProtocols != nil {
			service[skSubprotocols] = make(([]string), len(*v.SubProtocols))
			copy(service[skSubprotocols].([]string), *v.SubProtocols)
		}
		services[i] = service
	}
	if tfErr := d.Set("services", services); tfErr != nil {
		panic(tfErr)
	}

	if jsonData.Tags != nil && len(*jsonData.Tags) > 0 {
		apiTags := *jsonData.Tags

		stateTags := make([]interface{}, len(apiTags))

		for i, tagMap := range apiTags {
			stateMap := map[string]interface{}{
				skKey:   tagMap[skKey],
				skValue: tagMap[skValue],
			}
			stateTags[i] = stateMap
		}

		if tfErr := d.Set("tags", stateTags); tfErr != nil {
			panic(tfErr)
		}
	} else {
		if tfErr := d.Set("tags", make([]interface{}, 0)); tfErr != nil {
			panic(tfErr)
		}
	}
}
