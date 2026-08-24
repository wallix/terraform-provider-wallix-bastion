package bastion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type jsonDeviceService struct {
	Port             int       `json:"port"`
	ID               string    `json:"id,omitempty"`
	ConnectionPolicy string    `json:"connection_policy"`
	Protocol         string    `json:"protocol,omitempty"`
	ServiceName      string    `json:"service_name,omitempty"`
	GlobalDomains    *[]string `json:"global_domains,omitempty"`
	SubProtocols     *[]string `json:"subprotocols,omitempty"`
}

func resourceDeviceService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeviceServiceCreate,
		ReadContext:   resourceDeviceServiceRead,
		UpdateContext: resourceDeviceServiceUpdate,
		DeleteContext: resourceDeviceServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDeviceServiceImport,
		},
		Schema: map[string]*schema.Schema{
			skDeviceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			skServiceName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			skConnectionPolicy: {
				Type:     schema.TypeString,
				Required: true,
			},
			skPort: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			skProtocol: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{skProtoSSH, "RAWTCPIP", skProtoRDP, skProtoRLOGIN, skProtoTELNET, "VNC"},
					false,
				),
			},
			skGlobalDomains: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			skSubprotocols: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDeviceServiceVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_device_service not available with api version %s", version)
}

func resourceDeviceServiceCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceServiceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDeviceOptions(ctx, d.Get(skDeviceID).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		return diag.FromErr(fmt.Errorf("device with ID %s doesn't exists", d.Get(skDeviceID).(string)))
	}
	_, ex, err := searchResourceDeviceService(ctx, d.Get(skDeviceID).(string), d.Get(skServiceName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("service_name %s on device_id %s already exists",
			d.Get(skServiceName).(string), d.Get(skDeviceID).(string)))
	}
	id, err := addDeviceService(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceDeviceService(ctx, d.Get(skDeviceID).(string), d.Get(skServiceName).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("service_name %s on device_id %s not found after POST",
				d.Get(skServiceName).(string), d.Get(skDeviceID).(string)))
		}
	}
	d.SetId(id)

	return resourceDeviceServiceRead(ctx, d, m)
}

func resourceDeviceServiceRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceServiceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDeviceServiceOptions(ctx, d.Get(skDeviceID).(string), d.Id(), m)
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

func resourceDeviceServiceUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceDeviceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDeviceService(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDeviceServiceRead(ctx, d, m)
}

func resourceDeviceServiceDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDeviceServiceVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDeviceService(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDeviceServiceImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceDeviceServiceVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, errors.New("id must be <device_id>/<service_name>")
	}
	id, ex, err := searchResourceDeviceService(ctx, idSplit[0], idSplit[1], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find service_name with id %s (id must be <device_id>/<service_name>)", d.Id())
	}
	cfg, err := readDeviceServiceOptions(ctx, idSplit[0], id, m)
	if err != nil {
		return nil, err
	}
	fillDeviceService(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set(skDeviceID, idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceDeviceService(
	ctx context.Context, deviceID, serviceName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/devices/"+deviceID+
		"/services/?q=service_name="+serviceName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonDeviceService
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addDeviceService(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData, err := prepareDeviceServiceJSON(d, true)
	if err != nil {
		return "", err
	}
	body, headers, code, err := c.newRequestWithHeaders(
		ctx, "/devices/"+d.Get(skDeviceID).(string)+"/services/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateDeviceService(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	json, err := prepareDeviceServiceJSON(d, false)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get(skDeviceID).(string)+"/services/"+d.Id()+"?force=true", http.MethodPut, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteDeviceService(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/devices/"+d.Get(skDeviceID).(string)+"/services/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func sshSubProtocolsValid() []string {
	return []string{
		"SSH_SHELL_SESSION",
		"SSH_REMOTE_COMMAND",
		"SSH_SCP_UP",
		"SSH_SCP_DOWN",
		"SSH_X11",
		"SFTP_SESSION",
		"SSH_DIRECT_TCPIP",
		"SSH_REVERSE_TCPIP",
		"SSH_AUTH_AGENT",
		"SSH_DIRECT_UNIXSOCK",
		"SSH_REVERSE_UNIXSOCK",
	}
}

func rdpSubProtocolsValid() []string {
	return []string{
		"RDP_CLIPBOARD_UP",
		"RDP_CLIPBOARD_DOWN",
		"RDP_CLIPBOARD_FILE",
		"RDP_PRINTER",
		"RDP_COM_PORT",
		"RDP_DRIVE",
		"RDP_SMARTCARD",
		"RDP_AUDIO_OUTPUT",
		"RDP_AUDIO_INPUT",
	}
}

func prepareDeviceServiceJSON(
	d *schema.ResourceData, newResource bool,
) (
	jsonDeviceService, error,
) {
	jsonData := jsonDeviceService{
		ConnectionPolicy: d.Get(skConnectionPolicy).(string),
		Port:             d.Get(skPort).(int),
	}

	if newResource {
		jsonData.ServiceName = d.Get(skServiceName).(string)
		jsonData.Protocol = d.Get(skProtocol).(string)
	}

	if d.HasChange(skGlobalDomains) {
		listGlobalDomains := d.Get(skGlobalDomains).(*schema.Set).List()
		globalDomains := make([]string, len(listGlobalDomains))
		for i, v := range listGlobalDomains {
			globalDomains[i] = v.(string)
		}
		jsonData.GlobalDomains = &globalDomains
	}

	if listSubProtocols := d.Get(skSubprotocols).(*schema.Set).List(); len(listSubProtocols) > 0 {
		subProtocols := make([]string, len(listSubProtocols))
		for i, v := range listSubProtocols {
			switch d.Get(skProtocol).(string) {
			case skProtoSSH:
				if !slices.Contains(sshSubProtocolsValid(), v.(string)) {
					return jsonData, fmt.Errorf("subprotocols %s not valid for SSH service", v)
				}
				subProtocols[i] = v.(string)
			case skProtoRDP:
				if !slices.Contains(rdpSubProtocolsValid(), v.(string)) {
					return jsonData, fmt.Errorf("subprotocols %s not valid for RDP service", v)
				}
				subProtocols[i] = v.(string)
			default:
				return jsonData, fmt.Errorf("subprotocols need to not set for %s service", d.Get(skProtocol).(string))
			}
		}
		jsonData.SubProtocols = &subProtocols
	}

	return jsonData, nil
}

func readDeviceServiceOptions(
	ctx context.Context, deviceID, serviceID string, m interface{},
) (
	jsonDeviceService, error,
) {
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
		return result, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("unmarshaling json: %w", err)
	}

	return result, nil
}

func fillDeviceService(d *schema.ResourceData, jsonData jsonDeviceService) {
	if tfErr := d.Set(skServiceName, jsonData.ServiceName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skConnectionPolicy, jsonData.ConnectionPolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPort, jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skProtocol, jsonData.Protocol); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skGlobalDomains, jsonData.GlobalDomains); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skSubprotocols, jsonData.SubProtocols); tfErr != nil {
		panic(tfErr)
	}
}
