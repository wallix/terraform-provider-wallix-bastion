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

type jsonConnectionPolicy struct {
	ID                    string                 `json:"id,omitempty"`
	ConnectionPolicyName  string                 `json:"connection_policy_name"`
	Description           string                 `json:"description"`
	Protocol              string                 `json:"protocol"`
	Options               map[string]interface{} `json:"options"`
	AuthenticationMethods []string               `json:"authentication_methods"`
}

func resourceConnectionPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectionPolicyCreate,
		ReadContext:   resourceConnectionPolicyRead,
		UpdateContext: resourceConnectionPolicyUpdate,
		DeleteContext: resourceConnectionPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceConnectionPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"connection_policy_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"SSH", "RAWTCPIP", "RDP", "RLOGIN", "TELNET", "VNC"}, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"authentication_methods": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"options": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
			},
		},
	}
}
func resourceConnectionPolicyVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_connection_policy not validate with api version %s", version)
}

func resourceConnectionPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConnectionPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceConnectionPolicy(ctx, d.Get("connection_policy_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("connection_policy_name %s already exists", d.Get("connection_policy_name").(string)))
	}
	err = addConnectionPolicy(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceConnectionPolicy(ctx, d.Get("connection_policy_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("connection_policy_name %s can't find after POST",
			d.Get("connection_policy_name").(string)))
	}
	d.SetId(id)

	return resourceConnectionPolicyRead(ctx, d, m)
}
func resourceConnectionPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConnectionPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConnectionPolicyOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillConnectionPolicy(d, cfg)
	}

	return nil
}
func resourceConnectionPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceConnectionPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConnectionPolicy(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceConnectionPolicyRead(ctx, d, m)
}
func resourceConnectionPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConnectionPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteConnectionPolicy(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceConnectionPolicyImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceConnectionPolicyVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceConnectionPolicy(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find connection_policy_name with id %s (id must be <connection_policy_name>", d.Id())
	}
	cfg, err := readConnectionPolicyOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillConnectionPolicy(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceConnectionPolicy(ctx context.Context,
	connectionPolicyName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/connectionpolicies/?fields=connection_policy_name,id&limit=-1", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonConnectionPolicy
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	for _, v := range results {
		if v.ConnectionPolicyName == connectionPolicyName {
			return v.ID, true, nil
		}
	}

	return "", false, nil
}

func addConnectionPolicy(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData, err := prepareConnectionPolicyJSON(d)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx, "/connectionpolicies/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateConnectionPolicy(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData, err := prepareConnectionPolicyJSON(d)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx, "/connectionpolicies/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteConnectionPolicy(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/connectionpolicies/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareConnectionPolicyJSON(d *schema.ResourceData) (jsonConnectionPolicy, error) {
	var jsonData jsonConnectionPolicy
	jsonData.ConnectionPolicyName = d.Get("connection_policy_name").(string)
	jsonData.Description = d.Get("description").(string)
	jsonData.Protocol = d.Get("protocol").(string)
	if v := d.Get("authentication_methods").(*schema.Set).List(); len(v) > 0 {
		for _, vv := range v {
			if !bchk.StringInSlice(vv.(string), validAuthenticationMethods()) {
				return jsonData, fmt.Errorf("authentication_methods must be in %v", validAuthenticationMethods())
			}
			jsonData.AuthenticationMethods = append(jsonData.AuthenticationMethods, vv.(string))
		}
	} else {
		jsonData.AuthenticationMethods = make([]string, 0)
	}
	var options map[string]interface{}
	if v := d.Get("options").(string); v != "" {
		_ = json.Unmarshal([]byte(v), &options)
	} else {
		_ = json.Unmarshal([]byte(`{}`), &options)
	}
	jsonData.Options = options

	return jsonData, nil
}
func validAuthenticationMethods() []string {
	return []string{
		"KERBEROS_FORWARDING",
		"PASSWORD_INTERACTIVE",
		"PASSWORD_MAPPING",
		"PASSWORD_VAULT",
		"PUBKEY_AGENT_FORWARDING",
		"PUBKEY_VAULT",
	}
}

func readConnectionPolicyOptions(
	ctx context.Context, connectionPolicyID string, m interface{}) (jsonConnectionPolicy, error) {
	c := m.(*Client)
	var result jsonConnectionPolicy
	body, code, err := c.newRequest(ctx, "/connectionpolicies/"+connectionPolicyID, http.MethodGet, nil)
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

func fillConnectionPolicy(d *schema.ResourceData, jsonData jsonConnectionPolicy) {
	if tfErr := d.Set("connection_policy_name", jsonData.ConnectionPolicyName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("protocol", jsonData.Protocol); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("authentication_methods", jsonData.AuthenticationMethods); tfErr != nil {
		panic(tfErr)
	}
	options, _ := json.Marshal(jsonData.Options) // nolint: errchkjson
	if tfErr := d.Set("options", string(options)); tfErr != nil {
		panic(tfErr)
	}
}
