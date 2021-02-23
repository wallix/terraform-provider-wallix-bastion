package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type jsonDevice struct {
	ID             string                   `json:"id,omitempty"`
	Alias          string                   `json:"alias"`
	Description    string                   `json:"description"`
	DeviceName     string                   `json:"device_name"`
	Host           string                   `json:"host"`
	LastConnection string                   `json:"last_connection,omitempty"`
	LocalDomains   *[]jsonDeviceLocalDomain `json:"local_domains,omitempty"`
	Services       *[]jsonDeviceService     `json:"services,omitempty"`
}

func resourceDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeviceCreate,
		ReadContext:   resourceDeviceRead,
		UpdateContext: resourceDeviceUpdate,
		DeleteContext: resourceDeviceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDeviceImport,
		},
		Schema: map[string]*schema.Schema{
			"device_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"host": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
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
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"admin_account": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ca_public_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enable_password_change": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"password_change_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password_change_plugin": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password_change_plugin_parameters": {
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
						"service_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"connection_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"global_domains": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subprotocols": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}
func resourveDeviceVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_device not validate with api version %s", version)
}

func resourceDeviceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceDevice(ctx, d.Get("device_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("device_name %s already exists", d.Get("device_name").(string)))
	}
	err = addDevice(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDevice(ctx, d.Get("device_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("device_name %s can't find after POST", d.Get("device_name").(string)))
	}
	d.SetId(id)

	return resourceDeviceRead(ctx, d, m)
}
func resourceDeviceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceVersionCheck(c.bastionAPIVersion); err != nil {
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
func resourceDeviceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourveDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDevice(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDeviceRead(ctx, d, m)
}
func resourceDeviceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDevice(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceDeviceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourveDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceDevice(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find device_name with id %s (id must be <device_name>", d.Id())
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

func searchResourceDevice(ctx context.Context, deviceName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/devices/", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonDevice
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	for _, v := range results {
		if v.DeviceName == deviceName {
			return v.ID, true, nil
		}
	}

	return "", false, nil
}

func addDevice(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareDeviceJSON(d)
	body, code, err := c.newRequest(ctx, "/devices/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateDevice(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareDeviceJSON(d)
	body, code, err := c.newRequest(ctx, "/devices/"+d.Id(), http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteDevice(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/devices/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareDeviceJSON(d *schema.ResourceData) jsonDevice {
	return jsonDevice{
		DeviceName:  d.Get("device_name").(string),
		Host:        d.Get("host").(string),
		Alias:       d.Get("alias").(string),
		Description: d.Get("description").(string),
	}
}

func readDeviceOptions(
	ctx context.Context, deviceID string, m interface{}) (jsonDevice, error) {
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
		return result, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("json.Unmarshal failed : %w", err)
	}

	return result, nil
}

func fillDevice(d *schema.ResourceData, jsonData jsonDevice) {
	if tfErr := d.Set("device_name", jsonData.DeviceName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("host", jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("alias", jsonData.Alias); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	localDomains := make([]map[string]interface{}, 0)
	for _, v := range *jsonData.LocalDomains {
		localDomains = append(localDomains, map[string]interface{}{
			"id":                     v.ID,
			"admin_account":          v.AdminAccount,
			"domain_name":            v.DomainName,
			"ca_public_key":          v.CAPublicKey,
			"description":            v.Description,
			"enable_password_change": v.EnablePasswordChange,
			"password_change_policy": v.PasswordChangePolicy,
			"password_change_plugin": v.PasswordChangePlugin,
		})
		pluginParameters, _ := json.Marshal(v.PasswordChangePluginParameters)
		localDomains[len(localDomains)-1]["password_change_plugin_parameters"] = string(pluginParameters)
	}
	if tfErr := d.Set("local_domains", localDomains); tfErr != nil {
		panic(tfErr)
	}
	services := make([]map[string]interface{}, 0)
	for _, v := range *jsonData.Services {
		service := map[string]interface{}{
			"id":                v.ID,
			"service_name":      v.ServiceName,
			"connection_policy": v.ConnectionPolicy,
			"port":              v.Port,
			"protocol":          v.Protocol,
			"global_domains":    make([]string, 0),
			"subprotocols":      make([]string, 0),
		}
		for _, v2 := range v.GlobalDomains {
			service["global_domains"] = append(service["global_domains"].([]string), v2)
		}
		if v.SubProtocols != nil {
			for _, v2 := range *v.SubProtocols {
				service["subprotocols"] = append(service["subprotocols"].([]string), v2)
			}
		}
		services = append(services, service)
	}
	if tfErr := d.Set("services", services); tfErr != nil {
		panic(tfErr)
	}
}
