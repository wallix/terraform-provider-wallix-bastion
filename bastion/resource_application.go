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
	ID                  string                        `json:"id,omitempty"`
	ApplicationName     string                        `json:"application_name"`
	ConnectionPolicy    string                        `json:"connection_policy"`
	Category            string                        `json:"category,omitempty"`
	ApplicationURL      *string                       `json:"application_url,omitempty"`
	LoginFormURL        *string                       `json:"login_form_url,omitempty"`
	LoginButtonSelector *string                       `json:"login_button_selector,omitempty"`
	AllowNonPostForm    *bool                         `json:"allow_non_post_form,omitempty"`
	Browser             *string                       `json:"browser,omitempty"`
	BrowserVersion      *string                       `json:"browser_version,omitempty"`
	Description         string                        `json:"description"`
	Parameters          *string                       `json:"parameters,omitempty"`
	Target              *string                       `json:"target,omitempty"`
	GlobalDomains       *[]string                     `json:"global_domains,omitempty"`
	Paths               *[]jsonApplicationPath        `json:"paths,omitempty"`
	LocalDomains        *[]jsonApplicationLocalDomain `json:"local_domains,omitempty"`
	Tags                *[]map[string]string          `json:"tags,omitempty"`
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
			StateContext: resourceApplicationImport,
		},
		Schema: map[string]*schema.Schema{
			"application_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			skConnectionPolicy: {
				Type:     schema.TypeString,
				Required: true,
			},
			"category": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      skStandard,
				ValidateFunc: validation.StringInSlice([]string{skStandard, "jumphost", "web_application"}, false),
			},
			skAllowNonPostForm: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			skGlobalDomains: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			skLoginButtonSelector: {
				Type:     schema.TypeString,
				Optional: true,
			},
			skLoginFormURL: {
				Type:     schema.TypeString,
				Optional: true,
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
						skTarget: {
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
			skTarget: {
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
						skAdminAccount: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skDomainName: {
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
	id, err := addApplication(ctx, d, m, c.bastionAPIVersion)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceApplication(ctx, d.Get("application_name").(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("application_name %s not found after POST", d.Get("application_name").(string)))
		}
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
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
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
) (string, error) {
	c := m.(*Client)
	jsonData, err := prepareApplicationJSON(d, true, apiVersion)
	if err != nil {
		return "", err
	}
	body, headers, code, err := c.newRequestWithHeaders(ctx, "/applications/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
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

//nolint:gocognit
func prepareApplicationJSON(
	d *schema.ResourceData, newResource bool, apiVersion string,
) (
	jsonApplication, error,
) {
	jsonData := jsonApplication{
		ApplicationName:  d.Get("application_name").(string),
		ConnectionPolicy: d.Get(skConnectionPolicy).(string),
		Description:      d.Get(skDescription).(string),
	}
	if v, ok := d.GetOk("parameters"); ok {
		parameters := v.(string)
		jsonData.Parameters = &parameters
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

	// The category drives validation and the payload on every call, but it is only sent at
	// creation: it is ForceNew in the schema, so an update never changes it.
	category := d.Get("category").(string)
	if newResource &&
		semver.Compare(apiVersion, VersionWallixAPI312) >= 0 {
		jsonData.Category = category
	}
	switch category {
	case "", skStandard:
		if d.Get("application_url").(string) != "" {
			return jsonData, errors.New("application_url cannot be configured when category = standard")
		}
		if d.Get("browser").(string) != "" {
			return jsonData, errors.New("browser cannot be configured when category = standard")
		}
		if d.Get("browser_version").(string) != "" {
			return jsonData, errors.New("browser_version cannot be configured when category = standard")
		}
		if err := rejectApplicationWebLoginFields(d, skStandard); err != nil {
			return jsonData, err
		}

		target := d.Get(skTarget).(string)
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
				Target:     paths[skTarget].(string),
				Program:    paths["program"].(string),
				WorkingDir: paths["working_dir"].(string),
			}
		}
		jsonData.Paths = &jsonDataPaths

		listGlobalDomains := d.Get(skGlobalDomains).(*schema.Set).List()
		jsonDataGlobalDomains := make([]string, len(listGlobalDomains))
		for i, v := range listGlobalDomains {
			jsonDataGlobalDomains[i] = v.(string)
		}
		jsonData.GlobalDomains = &jsonDataGlobalDomains

	case "jumphost":
		// jumphost was introduced in API v3.9 and deprecated/removed in API v3.12
		if apiVersion != "" && semver.Compare(apiVersion, VersionWallixAPI312) >= 0 {
			return jsonData, fmt.Errorf(
				"category = jumphost is no longer supported in API version %s (deprecated in v3.12, use 'web_application' instead)",
				apiVersion,
			)
		}
		if d.Get(skTarget).(string) != "" {
			return jsonData, errors.New("target cannot be configured when category = jumphost")
		}
		if len(d.Get("paths").(*schema.Set).List()) > 0 {
			return jsonData, errors.New("paths cannot be configured when category = jumphost")
		}
		if len(d.Get(skGlobalDomains).(*schema.Set).List()) > 0 {
			return jsonData, errors.New("global_domains cannot be configured when category = jumphost")
		}
		if err := rejectApplicationWebLoginFields(d, "jumphost"); err != nil {
			return jsonData, err
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

	case "web_application":
		// web_application category was introduced in API v3.12 to replace jumphost
		if apiVersion != "" && semver.Compare(apiVersion, VersionWallixAPI312) < 0 {
			return jsonData, fmt.Errorf(
				"category = web_application not available with API version %s (requires v3.12+)",
				apiVersion,
			)
		}
		if d.Get(skTarget).(string) != "" {
			return jsonData, errors.New("target cannot be configured when category = web_application")
		}
		if len(d.Get("paths").(*schema.Set).List()) > 0 {
			return jsonData, errors.New("paths cannot be configured when category = web_application")
		}
		if d.Get("browser").(string) != "" {
			return jsonData, errors.New("browser cannot be configured when category = web_application")
		}
		if d.Get("browser_version").(string) != "" {
			return jsonData, errors.New("browser_version cannot be configured when category = web_application")
		}
		if d.Get("parameters").(string) != "" {
			return jsonData, errors.New("parameters cannot be configured when category = web_application")
		}

		applicationURL := d.Get("application_url").(string)
		if applicationURL == "" {
			return jsonData, errors.New("application_url must be specified when category = web_application")
		}
		jsonData.ApplicationURL = &applicationURL
		jsonData.LoginFormURL = applicationStringToSend(d, skLoginFormURL, newResource)
		jsonData.LoginButtonSelector = applicationStringToSend(d, skLoginButtonSelector, newResource)
		jsonData.AllowNonPostForm = applicationBoolToSend(d, skAllowNonPostForm, newResource)

		listGlobalDomains := d.Get(skGlobalDomains).(*schema.Set).List()
		jsonDataGlobalDomains := make([]string, len(listGlobalDomains))
		for i, v := range listGlobalDomains {
			jsonDataGlobalDomains[i] = v.(string)
		}
		jsonData.GlobalDomains = &jsonDataGlobalDomains
	}

	return jsonData, nil
}

// rejectApplicationWebLoginFields returns an error when one of the login automation fields,
// which only apply to category = web_application, is set for another category.
func rejectApplicationWebLoginFields(d *schema.ResourceData, category string) error {
	for _, key := range []string{skLoginFormURL, skLoginButtonSelector} {
		if d.Get(key).(string) != "" {
			return fmt.Errorf("%s cannot be configured when category = %s", key, category)
		}
	}
	if d.Get(skAllowNonPostForm).(bool) {
		return fmt.Errorf("%s cannot be configured when category = %s", skAllowNonPostForm, category)
	}

	return nil
}

// applicationStringToSend returns the value of an optional field for the request, or nil to leave
// it out. An empty value is left out at creation and when it is unchanged, so a configuration
// that never sets the field never sends it; it is only sent to clear the field on update.
func applicationStringToSend(d *schema.ResourceData, key string, newResource bool) *string {
	v := d.Get(key).(string)
	if v == "" && (newResource || !d.HasChange(key)) {
		return nil
	}

	return &v
}

// applicationBoolToSend is applicationStringToSend for a boolean field, with false as the empty value.
func applicationBoolToSend(d *schema.ResourceData, key string, newResource bool) *bool {
	v := d.Get(key).(bool)
	if !v && (newResource || !d.HasChange(key)) {
		return nil
	}

	return &v
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

func fillApplication(d *schema.ResourceData, jsonData jsonApplication) {
	if tfErr := d.Set("application_name", jsonData.ApplicationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skConnectionPolicy, jsonData.ConnectionPolicy); tfErr != nil {
		panic(tfErr)
	}
	category := jsonData.Category
	if category == "" {
		category = skStandard
	}
	if tfErr := d.Set("category", category); tfErr != nil {
		panic(tfErr)
	}
	setApplicationOptionalString(d, "application_url", jsonData.ApplicationURL)
	setApplicationOptionalString(d, skLoginFormURL, jsonData.LoginFormURL)
	setApplicationOptionalString(d, skLoginButtonSelector, jsonData.LoginButtonSelector)
	setApplicationOptionalBool(d, skAllowNonPostForm, jsonData.AllowNonPostForm)
	setApplicationOptionalString(d, "browser", jsonData.Browser)
	setApplicationOptionalString(d, "browser_version", jsonData.BrowserVersion)
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skGlobalDomains, jsonData.GlobalDomains); tfErr != nil {
		panic(tfErr)
	}
	setApplicationOptionalString(d, "parameters", jsonData.Parameters)
	if tfErr := d.Set("paths", fillApplicationPaths(jsonData.Paths)); tfErr != nil {
		panic(tfErr)
	}
	setApplicationOptionalString(d, skTarget, jsonData.Target)
	if tfErr := d.Set("local_domains", fillApplicationLocalDomains(jsonData.LocalDomains)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("tags", fillApplicationTags(jsonData.Tags)); tfErr != nil {
		panic(tfErr)
	}
}

// setApplicationOptionalString sets key to the dereferenced value, or "" when value is nil -
// shared by the *string fields that follow the same "unset means empty" convention.
func setApplicationOptionalString(d *schema.ResourceData, key string, value *string) {
	v := ""
	if value != nil {
		v = *value
	}
	if tfErr := d.Set(key, v); tfErr != nil {
		panic(tfErr)
	}
}

// setApplicationOptionalBool is setApplicationOptionalString for a *bool field, with false
// when value is nil.
func setApplicationOptionalBool(d *schema.ResourceData, key string, value *bool) {
	v := value != nil && *value
	if tfErr := d.Set(key, v); tfErr != nil {
		panic(tfErr)
	}
}

func fillApplicationPaths(jsonPaths *[]jsonApplicationPath) []map[string]interface{} {
	paths := make([]map[string]interface{}, 0)
	if jsonPaths != nil {
		paths = make([]map[string]interface{}, len(*jsonPaths))
		for i, v := range *jsonPaths {
			paths[i] = map[string]interface{}{
				skTarget:      v.Target,
				"program":     v.Program,
				"working_dir": v.WorkingDir,
			}
		}
	}

	return paths
}

func fillApplicationLocalDomains(jsonLocalDomains *[]jsonApplicationLocalDomain) []map[string]interface{} {
	localDomains := make([]map[string]interface{}, 0)
	if jsonLocalDomains != nil {
		localDomains = make([]map[string]interface{}, len(*jsonLocalDomains))
		for i, v := range *jsonLocalDomains {
			localDomains[i] = map[string]interface{}{
				"id":                   v.ID,
				skAdminAccount:         v.AdminAccount,
				skDomainName:           v.DomainName,
				skDescription:          v.Description,
				skEnablePasswordChange: v.EnablePasswordChange,
				skPasswordChangePolicy: v.PasswordChangePolicy,
				skPasswordChangePlugin: v.PasswordChangePlugin,
			}
			pluginParameters, _ := json.Marshal(v.PasswordChangePluginParameters) //nolint: errchkjson
			localDomains[i][skPasswordChangePluginParameters] = string(pluginParameters)
		}
	}

	return localDomains
}

func fillApplicationTags(jsonTags *[]map[string]string) []interface{} {
	stateTags := make([]interface{}, 0)
	if jsonTags != nil {
		stateTags = make([]interface{}, len(*jsonTags))
		for i, tagMap := range *jsonTags {
			stateTags[i] = map[string]interface{}{
				skKey:   tagMap[skKey],
				skValue: tagMap[skValue],
			}
		}
	}

	return stateTags
}
