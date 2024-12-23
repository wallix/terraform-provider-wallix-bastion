package bastion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/mod/semver"
)

type jsonApplication struct {
	ID               string                        `json:"id,omitempty"`
	ApplicationName  string                        `json:"application_name"`
	ConnectionPolicy string                        `json:"connection_policy"`
	Category         string                        `json:"category,omitempty"`
	ApplicationURL   *string                       `json:"application_url,omitempty"`
	Browser          *string                       `json:"browser,omitempty"`
	BrowserVersion   *string                       `json:"browser_version,omitempty"`
	Description      string                        `json:"description"`
	Parameters       string                        `json:"parameters"`
	Target           *string                       `json:"target,omitempty"`
	GlobalDomains    *[]string                     `json:"global_domains,omitempty"`
	Paths            *[]jsonApplicationPath        `json:"paths,omitempty"`
	LocalDomains     *[]jsonApplicationLocalDomain `json:"local_domains,omitempty"`
}

type jsonApplicationPath struct {
	Target     string `json:"target"`
	Program    string `json:"program"`
	WorkingDir string `json:"working_dir"`
}

func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationCreate,
		ReadContext:   resourceApplicationRead,
		UpdateContext: resourceApplicationUpdate,
		DeleteContext: resourceApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceApplicationImport,
		},
		Schema: map[string]*schema.Schema{
			"application_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"connection_policy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"category": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "standard",
				ValidateFunc: validation.StringInSlice([]string{"standard", "jumphost"}, false),
			},
			"application_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"browser": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"browser_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"global_domains": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"parameters": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"paths": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target": {
							Type:     schema.TypeString,
							Required: true,
						},
						"program": {
							Type:     schema.TypeString,
							Required: true,
						},
						"working_dir": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},
			"target": {
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
						"admin_account": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_name": {
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
		},
	}
}

func resourceApplicationVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_application not available with api version %s", version)
}

func resourceApplicationCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceApplication(ctx, d.Get("application_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("application_name %s already exists", d.Get("application_name").(string)))
	}
	err = addApplication(ctx, d, m, c.bastionAPIVersion)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceApplication(ctx, d.Get("application_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("application_name %s not found after POST", d.Get("application_name").(string)))
	}
	d.SetId(id)

	return resourceApplicationRead(ctx, d, m)
}

func resourceApplicationRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readApplicationOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillApplication(d, cfg)
	}

	return nil
}

func resourceApplicationUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceApplicationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateApplication(ctx, d, m, c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceApplicationRead(ctx, d, m)
}

func resourceApplicationDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteApplication(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceApplicationImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceApplicationVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceApplication(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find application_name with id %s (id must be <application_name>)", d.Id())
	}
	cfg, err := readApplicationOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillApplication(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceApplication(
	ctx context.Context, applicationName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/applications/?q=application_name="+applicationName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonApplication
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addApplication(
	ctx context.Context, d *schema.ResourceData, m interface{}, apiVersion string,
) error {
	c := m.(*Client)
	jsonData, err := prepareApplicationJSON(d, true, apiVersion)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx, "/applications/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateApplication(
	ctx context.Context, d *schema.ResourceData, m interface{}, apiVersion string,
) error {
	c := m.(*Client)
	jsonData, err := prepareApplicationJSON(d, false, apiVersion)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx, "/applications/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteApplication(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/applications/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareApplicationJSON(
	d *schema.ResourceData, newResource bool, apiVersion string,
) (
	jsonApplication, error,
) {
	jsonData := jsonApplication{
		ApplicationName:  d.Get("application_name").(string),
		ConnectionPolicy: d.Get("connection_policy").(string),
		Description:      d.Get("description").(string),
		Parameters:       d.Get("parameters").(string),
	}
	if newResource &&
		semver.Compare(apiVersion, VersionWallixAPI312) >= 0 {
		jsonData.Category = d.Get("category").(string)
	}
	switch jsonData.Category {
	case "", "standard":
		if d.Get("application_url").(string) != "" {
			return jsonData, errors.New("application_url cannot be configured when category = standard")
		}
		if d.Get("browser").(string) != "" {
			return jsonData, errors.New("browser cannot be configured when category = standard")
		}
		if d.Get("browser_version").(string) != "" {
			return jsonData, errors.New("browser_version cannot be configured when category = standard")
		}

		target := d.Get("target").(string)
		if target == "" {
			return jsonData, errors.New("target must be specified when category = standard")
		}
		jsonData.Target = &target

		listPaths := d.Get("paths").(*schema.Set).List()
		if len(listPaths) == 0 {
			return jsonData, errors.New("paths must be specified when category = standard")
		}
		jsonDataPaths := make([]jsonApplicationPath, len(listPaths))
		for i, v := range listPaths {
			paths := v.(map[string]interface{})
			jsonDataPaths[i] = jsonApplicationPath{
				Target:     paths["target"].(string),
				Program:    paths["program"].(string),
				WorkingDir: paths["working_dir"].(string),
			}
		}
		jsonData.Paths = &jsonDataPaths

		listGlobalDomains := d.Get("global_domains").(*schema.Set).List()
		jsonDataGlobalDomains := make([]string, len(listGlobalDomains))
		for i, v := range listGlobalDomains {
			jsonDataGlobalDomains[i] = v.(string)
		}
		jsonData.GlobalDomains = &jsonDataGlobalDomains

	case "jumphost":
		if semver.Compare(apiVersion, VersionWallixAPI312) < 0 {
			return jsonData, fmt.Errorf("category = jumphost not available with api version %s", apiVersion)
		}
		if d.Get("target").(string) != "" {
			return jsonData, errors.New("target cannot be configured when category = jumphost")
		}
		if len(d.Get("paths").(*schema.Set).List()) > 0 {
			return jsonData, errors.New("paths cannot be configured when category = jumphost")
		}
		if len(d.Get("global_domains").(*schema.Set).List()) > 0 {
			return jsonData, errors.New("paths cannot be configured when category = jumphost")
		}

		applicationURL := d.Get("application_url").(string)
		if applicationURL == "" {
			return jsonData, errors.New("application_url must be specified when category = jumphost")
		}
		jsonData.ApplicationURL = &applicationURL

		browser := d.Get("browser").(string)
		if browser == "" {
			return jsonData, errors.New("browser must be specified when category = jumphost")
		}
		jsonData.Browser = &browser

		browserVersion := d.Get("browser_version").(string)
		jsonData.BrowserVersion = &browserVersion
	}

	return jsonData, nil
}

func readApplicationOptions(
	ctx context.Context, applicationID string, m interface{},
) (
	jsonApplication, error,
) {
	c := m.(*Client)
	var result jsonApplication
	body, code, err := c.newRequest(ctx, "/applications/"+applicationID, http.MethodGet, nil)
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

func fillApplication(d *schema.ResourceData, jsonData jsonApplication) { //nolint:gocognit
	if tfErr := d.Set("application_name", jsonData.ApplicationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("connection_policy", jsonData.ConnectionPolicy); tfErr != nil {
		panic(tfErr)
	}
	category := jsonData.Category
	if category == "" {
		category = "standard"
	}
	if tfErr := d.Set("category", category); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.ApplicationURL != nil {
		if tfErr := d.Set("application_url", *jsonData.ApplicationURL); tfErr != nil {
			panic(tfErr)
		}
	} else {
		if tfErr := d.Set("application_url", ""); tfErr != nil {
			panic(tfErr)
		}
	}
	if jsonData.Browser != nil {
		if tfErr := d.Set("browser", *jsonData.Browser); tfErr != nil {
			panic(tfErr)
		}
	} else {
		if tfErr := d.Set("browser", ""); tfErr != nil {
			panic(tfErr)
		}
	}
	if jsonData.BrowserVersion != nil {
		if tfErr := d.Set("browser_version", *jsonData.BrowserVersion); tfErr != nil {
			panic(tfErr)
		}
	} else {
		if tfErr := d.Set("browser_version", ""); tfErr != nil {
			panic(tfErr)
		}
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("global_domains", jsonData.GlobalDomains); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("parameters", jsonData.Parameters); tfErr != nil {
		panic(tfErr)
	}
	paths := make([]map[string]interface{}, 0)
	if jsonData.Paths != nil {
		paths = make([]map[string]interface{}, len(*jsonData.Paths))
		for i, v := range *jsonData.Paths {
			paths[i] = map[string]interface{}{
				"target":      v.Target,
				"program":     v.Program,
				"working_dir": v.WorkingDir,
			}
		}
	}
	if tfErr := d.Set("paths", paths); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.Target != nil {
		if tfErr := d.Set("target", *jsonData.Target); tfErr != nil {
			panic(tfErr)
		}
	} else {
		if tfErr := d.Set("target", ""); tfErr != nil {
			panic(tfErr)
		}
	}
	localDomains := make([]map[string]interface{}, len(*jsonData.LocalDomains))
	for i, v := range *jsonData.LocalDomains {
		localDomains[i] = map[string]interface{}{
			"id":                     v.ID,
			"admin_account":          v.AdminAccount,
			"domain_name":            v.DomainName,
			"description":            v.Description,
			"enable_password_change": v.EnablePasswordChange,
			"password_change_policy": v.PasswordChangePolicy,
			"password_change_plugin": v.PasswordChangePlugin,
		}
		pluginParameters, _ := json.Marshal(v.PasswordChangePluginParameters) //nolint: errchkjson
		localDomains[len(localDomains)-1]["password_change_plugin_parameters"] = string(pluginParameters)
	}
	if tfErr := d.Set("local_domains", localDomains); tfErr != nil {
		panic(tfErr)
	}
}
