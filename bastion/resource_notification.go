package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type jsonNotification struct {
	ID               string   `json:"id,omitempty"`
	NotificationName string   `json:"notification_name"`
	Description      string   `json:"description"`
	Enabled          bool     `json:"enabled"`
	Type             string   `json:"type"`
	Destination      string   `json:"destination"`
	Language         string   `json:"language"`
	Events           []string `json:"events"`
}

func resourceNotification() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNotificationCreate,
		ReadContext:   resourceNotificationRead,
		UpdateContext: resourceNotificationUpdate,
		DeleteContext: resourceNotificationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNotificationImport,
		},
		Schema: map[string]*schema.Schema{
			"notification_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			skType: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{skEmail}, false),
			},
			"destination": {
				Type:     schema.TypeString,
				Required: true,
			},
			skLanguage: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"de", "en", "es", "fr", "ru"}, false),
			},
			"events": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"cx_equipment",
						"daily_reporting",
						"disk_space_critical",
						"external_storage_full",
						"filesystem_full",
						"integrity_error",
						"licence_notifications",
						"new_fingerprint",
						"password_expired",
						"pattern_found",
						"primary_cx_failed",
						"raid_error",
						"rdp_outcxn_found",
						"rdp_pattern_found",
						"rdp_process_found",
						"secondary_cx_failed",
						"sessionlog_purge",
						"watchdog_notifications",
						"wrong_fingerprint",
					}, false),
				},
			},
		},
	}
}

func resourceNotificationVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_notification not available with api version %s", version)
}

func resourceNotificationCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceNotificationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceNotification(ctx, d.Get("notification_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf(
			"notification_name %s already exists", d.Get("notification_name").(string)))
	}
	id, err := addNotification(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceNotification(ctx, d.Get("notification_name").(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf(
				"notification_name %s not found after POST", d.Get("notification_name").(string)))
		}
	}
	d.SetId(id)

	return resourceNotificationRead(ctx, d, m)
}

func resourceNotificationRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceNotificationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readNotificationOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillNotification(d, cfg)
	}

	return nil
}

func resourceNotificationUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceNotificationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateNotification(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceNotificationRead(ctx, d, m)
}

func resourceNotificationDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceNotificationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteNotification(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNotificationImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceNotificationVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceNotification(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf(
			"don't find notification_name with id %s (id must be <notification_name>)", d.Id())
	}
	cfg, err := readNotificationOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillNotification(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceNotification(
	ctx context.Context, notificationName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/notifications/?q=notification_name="+notificationName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonNotification
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addNotification(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData := prepareNotificationJSON(d)
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/notifications/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateNotification(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareNotificationJSON(d)
	// force=true makes the API fully replace the events list instead of merging it with the
	// existing one (confirmed live: without it, PUT unions the new events into the old set and
	// an empty list is a no-op, so removing an event from the config would silently not apply).
	body, code, err := c.newRequest(ctx, "/notifications/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteNotification(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/notifications/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareNotificationJSON(d *schema.ResourceData) jsonNotification {
	listEvents := d.Get("events").(*schema.Set).List()
	events := make([]string, len(listEvents))
	for i, v := range listEvents {
		events[i] = v.(string)
	}

	return jsonNotification{
		NotificationName: d.Get("notification_name").(string),
		Description:      d.Get(skDescription).(string),
		Enabled:          d.Get("enabled").(bool),
		Type:             d.Get(skType).(string),
		Destination:      d.Get("destination").(string),
		Language:         d.Get(skLanguage).(string),
		Events:           events,
	}
}

func readNotificationOptions(
	ctx context.Context, notificationID string, m interface{},
) (
	jsonNotification, error,
) {
	c := m.(*Client)
	var result jsonNotification
	body, code, err := c.newRequest(ctx, "/notifications/"+notificationID, http.MethodGet, nil)
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

func fillNotification(d *schema.ResourceData, jsonData jsonNotification) {
	if tfErr := d.Set("notification_name", jsonData.NotificationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("enabled", jsonData.Enabled); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skType, jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("destination", jsonData.Destination); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skLanguage, jsonData.Language); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("events", jsonData.Events); tfErr != nil {
		panic(tfErr)
	}
}
