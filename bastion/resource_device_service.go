package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type jsonDeviceService struct {
	ID               string    `json:"id,omitempty"`
	ServiceName      string    `json:"service_name,omitempty"`
	Protocol         string    `json:"protocol,omitempty"`
	Port             int       `json:"port"`
	SubProtocols     *[]string `json:"subprotocols,omitempty"`
	ConnectionPolicy string    `json:"connection_policy"`
	GlobalDomains    []string  `json:"global_domains,omitempty"`
}

func resourceDeviceService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeviceServiceCreate,
		ReadContext:   resourceDeviceServiceRead,
		UpdateContext: resourceDeviceServiceUpdate,
		DeleteContext: resourceDeviceServiceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDeviceServiceImport,
		},
		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"connection_policy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"SSH", "RAWTCPIP", "RDP", "RLOGIN", "TELNET", "VNC"}, false),
			},
			"global_domains": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subprotocols": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}
func resourveDeviceServiceVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_device_service not validate with api version %s", version)
}

func resourceDeviceServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceServiceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDeviceOptions(ctx, d.Get("device_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		return diag.FromErr(fmt.Errorf("device with ID %s doesn't exists", d.Get("device_id").(string)))
	}
	_, ex, err := searchResourceDeviceService(ctx, d.Get("device_id").(string), d.Get("service_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("service_name %s on device_id %s already exists",
			d.Get("service_name").(string), d.Get("device_id").(string)))
	}
	err = addDeviceService(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDeviceService(ctx, d.Get("device_id").(string), d.Get("service_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("service_name %s on device_id %s can't find after POST",
			d.Get("service_name").(string), d.Get("device_id").(string)))
	}
	d.SetId(id)

	return resourceDeviceServiceRead(ctx, d, m)
}
func resourceDeviceServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceServiceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDeviceServiceOptions(ctx, d.Get("device_id").(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillDeviceService(d, cfg)
	}

	return nil
}
func resourceDeviceServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourveDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDeviceService(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDeviceServiceRead(ctx, d, m)
}
func resourceDeviceServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveDeviceServiceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDeviceService(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceDeviceServiceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourveDeviceServiceVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, fmt.Errorf("id msut be <device_id>/<service_name>")
	}
	id, ex, err := searchResourceDeviceService(ctx, idSplit[0], idSplit[1], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find service_name with id %s (id must be <device_id>/<service_name>", d.Id())
	}
	cfg, err := readDeviceServiceOptions(ctx, idSplit[0], id, m)
	if err != nil {
		return nil, err
	}
	fillDeviceService(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set("device_id", idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceDeviceService(ctx context.Context,
	deviceID, serviceName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/devices/"+deviceID+"/services/", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonDeviceService
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	for _, v := range results {
		if v.ServiceName == serviceName {
			return v.ID, true, nil
		}
	}

	return "", false, nil
}

func addDeviceService(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	json, err := prepareDeviceServiceJSON(d, true)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx, "/devices/"+d.Get("device_id").(string)+"/services/", http.MethodPost, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateDeviceService(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	json, err := prepareDeviceServiceJSON(d, false)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/services/"+d.Id()+"?force=true", http.MethodPost, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteDeviceService(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get("device_id").(string)+"/services/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func sshSubProtocolsValid() []string {
	return []string{"SSH_SHELL_SESSION", "SSH_REMOTE_COMMAND", "SSH_SCP_UP", "SSH_SCP_DOWN",
		"SSH_X11", "SFTP_SESSION", "SSH_DIRECT_TCPIP", "SSH_REVERSE_TCPIP", "SSH_AUTH_AGENT"}
}
func rdpSubProtocolsValid() []string {
	return []string{"RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_CLIPBOARD_FILE", "RDP_PRINTER",
		"RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_AUDIO_OUTPUT"}
}
func prepareDeviceServiceJSON(d *schema.ResourceData, newResource bool) (jsonDeviceService, error) {
	var json jsonDeviceService
	if newResource {
		json.ServiceName = d.Get("service_name").(string)
		json.Protocol = d.Get("protocol").(string)
	}
	json.ConnectionPolicy = d.Get("connection_policy").(string)
	json.Port = d.Get("port").(int)
	for _, v := range d.Get("global_domains").([]interface{}) {
		json.GlobalDomains = append(json.GlobalDomains, v.(string))
	}
	if v := d.Get("subprotocols").([]interface{}); len(v) > 0 {
		subProtocols := make([]string, 0)
		for _, v2 := range v {
			switch d.Get("protocol").(string) {
			case "SSH":
				if !stringInSlice(v2.(string), sshSubProtocolsValid()) {
					return json, fmt.Errorf("subprotocols %s not valid for SSH service", v2)
				}
				subProtocols = append(subProtocols, v2.(string))
			case "RDP":
				if !stringInSlice(v2.(string), rdpSubProtocolsValid()) {
					return json, fmt.Errorf("subprotocols %s not valid for RDP service", v2)
				}
				subProtocols = append(subProtocols, v2.(string))
			default:
				return json, fmt.Errorf("subprotocols need to not set for %s service", d.Get("protocol").(string))
			}
		}
		json.SubProtocols = &subProtocols
	}

	return json, nil
}

func readDeviceServiceOptions(
	ctx context.Context, deviceID, serviceID string, m interface{}) (jsonDeviceService, error) {
	c := m.(*Client)
	var result jsonDeviceService
	body, code, err := c.newRequest(ctx, "/devices/"+deviceID+"/services/"+serviceID, http.MethodGet, nil)
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

func fillDeviceService(d *schema.ResourceData, json jsonDeviceService) {
	if tfErr := d.Set("service_name", json.ServiceName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("connection_policy", json.ConnectionPolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("port", json.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("protocol", json.Protocol); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("global_domains", json.GlobalDomains); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("subprotocols", *json.SubProtocols); tfErr != nil {
		panic(tfErr)
	}
}
