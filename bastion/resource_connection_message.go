package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
)

type jsonConnectionMessage struct {
	MessageName string `json:"-"`
	Message     string `json:"message"`
}

func resourceConnectionMessage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectionMessageCreate,
		ReadContext:   resourceConnectionMessageRead,
		UpdateContext: resourceConnectionMessageUpdate,
		DeleteContext: resourceConnectionMessageDelete,
		Importer: &schema.ResourceImporter{
			State: resourceConnectionMessageImport,
		},
		Schema: map[string]*schema.Schema{
			"message_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						"login_en", "login_fr", "login_de", "login_es", "login_ru",
						"motd_en", "motd_fr", "motd_de", "motd_es", "motd_ru",
					},
					false,
				),
			},
			"message": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceConnectionMessageVersionCheck(version string) error {
	if bchk.InSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_connection_message not available with api version %s", version)
}

func resourceConnectionMessageCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConnectionMessageVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConnectionMessage(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(d.Get("message_name").(string))

	return resourceConnectionMessageRead(ctx, d, m)
}

func resourceConnectionMessageRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConnectionMessageVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConnectionMessage(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	cfg.MessageName = d.Get("message_name").(string)
	fillConnectionMessage(d, cfg)

	return nil
}

func resourceConnectionMessageUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceConnectionMessageVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConnectionMessage(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceConnectionMessageRead(ctx, d, m)
}

func resourceConnectionMessageDelete(
	_ context.Context, _ *schema.ResourceData, _ interface{},
) diag.Diagnostics {
	return nil
}

func resourceConnectionMessageImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceConnectionMessageVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	cfg, err := readConnectionMessage(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	cfg.MessageName = d.Id()
	fillConnectionMessage(d, cfg)
	result := make([]*schema.ResourceData, 1)
	result[0] = d

	return result, nil
}

func updateConnectionMessage(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(
		ctx,
		"/connectionmessages/"+d.Get("message_name").(string),
		http.MethodPut,
		prepareConnectionMessageJSON(d),
	)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareConnectionMessageJSON(d *schema.ResourceData) jsonConnectionMessage {
	jsonData := jsonConnectionMessage{
		Message: d.Get("message").(string),
	}

	return jsonData
}

func readConnectionMessage(
	ctx context.Context, connectionMessageName string, m interface{},
) (
	jsonConnectionMessage, error,
) {
	c := m.(*Client)
	var result jsonConnectionMessage
	body, code, err := c.newRequest(ctx, "/connectionmessages/"+connectionMessageName, http.MethodGet, nil)
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

func fillConnectionMessage(d *schema.ResourceData, jsonData jsonConnectionMessage) {
	if tfErr := d.Set("message_name", jsonData.MessageName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("message", jsonData.Message); tfErr != nil {
		panic(tfErr)
	}
}
