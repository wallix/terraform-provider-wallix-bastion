package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
)

type jsonCluster struct {
	ID                string   `json:"id,omitempty"`
	ClusterName       string   `json:"cluster_name"`
	Description       string   `json:"description"`
	Accounts          []string `json:"accounts"`
	AccountMappings   []string `json:"account_mappings"`
	InteractiveLogins []string `json:"interactive_logins"`
}

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceClusterImport,
		},
		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"accounts": {
				Type:         schema.TypeSet,
				Optional:     true,
				AtLeastOneOf: []string{"accounts", "account_mappings", "interactive_logins"},
				Elem:         &schema.Schema{Type: schema.TypeString},
			},
			"account_mappings": {
				Type:         schema.TypeSet,
				Optional:     true,
				AtLeastOneOf: []string{"accounts", "account_mappings", "interactive_logins"},
				Elem:         &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"interactive_logins": {
				Type:         schema.TypeSet,
				Optional:     true,
				AtLeastOneOf: []string{"accounts", "account_mappings", "interactive_logins"},
				Elem:         &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}
func resourceClusterVersionCheck(version string) error {
	if bchk.InSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_cluster not available with api version %s", version)
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceClusterVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceCluster(ctx, d.Get("cluster_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("cluster_name %s already exists", d.Get("cluster_name").(string)))
	}
	err = addCluster(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceCluster(ctx, d.Get("cluster_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("cluster_name %s not found after POST", d.Get("cluster_name").(string)))
	}
	d.SetId(id)

	return resourceClusterRead(ctx, d, m)
}
func resourceClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceClusterVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readClusterOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillCluster(d, cfg)
	}

	return nil
}
func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceClusterVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateCluster(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceClusterRead(ctx, d, m)
}
func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceClusterVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteCluster(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceClusterImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceClusterVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceCluster(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find cluster_name with id %s (id must be <cluster_name>", d.Id())
	}
	cfg, err := readClusterOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillCluster(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceCluster(ctx context.Context, clusterName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/clusters/?fields=cluster_name,id&limit=-1", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonCluster
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	for _, v := range results {
		if v.ClusterName == clusterName {
			return v.ID, true, nil
		}
	}

	return "", false, nil
}

func addCluster(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareClusterJSON(d)
	body, code, err := c.newRequest(ctx, "/clusters/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateCluster(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareClusterJSON(d)
	body, code, err := c.newRequest(ctx, "/clusters/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteCluster(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/clusters/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareClusterJSON(d *schema.ResourceData) jsonCluster {
	var jsonData jsonCluster
	jsonData.ClusterName = d.Get("cluster_name").(string)
	if len(d.Get("accounts").(*schema.Set).List()) > 0 {
		for _, v := range d.Get("accounts").(*schema.Set).List() {
			jsonData.Accounts = append(jsonData.Accounts, v.(string))
		}
	} else {
		jsonData.Accounts = make([]string, 0)
	}
	if len(d.Get("account_mappings").(*schema.Set).List()) > 0 {
		for _, v := range d.Get("account_mappings").(*schema.Set).List() {
			jsonData.AccountMappings = append(jsonData.AccountMappings, v.(string))
		}
	} else {
		jsonData.AccountMappings = make([]string, 0)
	}
	jsonData.Description = d.Get("description").(string)
	if len(d.Get("interactive_logins").(*schema.Set).List()) > 0 {
		for _, v := range d.Get("interactive_logins").(*schema.Set).List() {
			jsonData.InteractiveLogins = append(jsonData.InteractiveLogins, v.(string))
		}
	} else {
		jsonData.InteractiveLogins = make([]string, 0)
	}

	return jsonData
}

func readClusterOptions(
	ctx context.Context, clusterID string, m interface{}) (jsonCluster, error) {
	c := m.(*Client)
	var result jsonCluster
	body, code, err := c.newRequest(ctx, "/clusters/"+clusterID, http.MethodGet, nil)
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

func fillCluster(d *schema.ResourceData, jsonData jsonCluster) {
	if tfErr := d.Set("cluster_name", jsonData.ClusterName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("accounts", jsonData.Accounts); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("account_mappings", jsonData.AccountMappings); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("interactive_logins", jsonData.InteractiveLogins); tfErr != nil {
		panic(tfErr)
	}
}
