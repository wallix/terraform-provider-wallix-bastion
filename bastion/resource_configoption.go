package bastion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/mod/semver"
)

const skValues = "values"

// configFloat is the value of a float option: the API rejects a whole number without decimal
// point (e.g. 60 instead of 60.0), which is what jsonencode produces.
type configFloat float64

func (f configFloat) MarshalJSON() ([]byte, error) {
	s := strconv.FormatFloat(float64(f), 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}

	return []byte(s), nil
}

// configSection is a section of a configuration file, as opposed to an option whose value
// happens to be a JSON object: a section is filtered key by key, an option value is kept whole.
type configSection map[string]interface{}

func resourceConfigoption() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConfigoptionCreate,
		ReadContext:   resourceConfigoptionRead,
		UpdateContext: resourceConfigoptionUpdate,
		DeleteContext: resourceConfigoptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceConfigoptionImport,
		},
		Schema: map[string]*schema.Schema{
			"config_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     validation.StringIsNotEmpty,
				DiffSuppressFunc: suppressSameConfigID,
			},
			skValues: {
				Type:             schema.TypeString,
				Sensitive:        true,
				Required:         true,
				ValidateDiagFunc: validateConfigoptionValues,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"config_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// PUT on configoptions doesn't exist before API v3.12.
func resourceConfigoptionVersionCheck(version string) error {
	if semver.Compare(version, VersionWallixAPI312) >= 0 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_configoption not available with api version %s", version)
}

func resourceConfigoptionCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigoptionVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConfigoption(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(d.Get("config_id").(string))

	return resourceConfigoptionRead(ctx, d, m)
}

func resourceConfigoptionRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigoptionVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	var managed map[string]interface{}
	if v := d.Get(skValues).(string); v != "" {
		if err := json.Unmarshal([]byte(v), &managed); err != nil {
			return diag.FromErr(fmt.Errorf("unmarshaling values: %w", err))
		}
	}
	options := make([]string, 0, len(managed))
	for k := range managed {
		options = append(options, k)
	}
	slices.Sort(options)
	cfg, code, err := getConfigoption(ctx, m, d.Get("config_id").(string), options, false)
	if err != nil {
		return diag.FromErr(err)
	}
	if code == http.StatusNotFound {
		d.SetId("")

		return nil
	}
	values := configOptionsToValues(cfg.Options)
	if managed != nil {
		values = filterConfigValues(values, managed)
	}
	valuesJSON, err := json.Marshal(values)
	if err != nil {
		return diag.FromErr(fmt.Errorf("marshaling values: %w", err))
	}
	if tfErr := d.Set(skValues, string(valuesJSON)); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("config_name", cfg.ConfigName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("name", cfg.Name); tfErr != nil {
		panic(tfErr)
	}
	d.SetId(cfg.ID)

	return nil
}

func resourceConfigoptionUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceConfigoptionVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateConfigoption(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return resourceConfigoptionRead(ctx, d, m)
}

func resourceConfigoptionDelete(
	_ context.Context, d *schema.ResourceData, _ interface{},
) diag.Diagnostics {
	// A configuration file can't be deleted and the API doesn't reset options to their
	// default value; just drop it from state.
	d.SetId("")

	return nil
}

// Import ID is <config_id> to manage every option of the configuration,
// or <config_id>:<option>,<section>.<option> to manage only the listed options.
func resourceConfigoptionImport(
	_ context.Context, d *schema.ResourceData, _ interface{},
) ([]*schema.ResourceData, error) {
	configID, optionsList, hasOptions := strings.Cut(d.Id(), ":")
	if configID == "" {
		return nil, fmt.Errorf("invalid id %q, must be <config_id> or <config_id>:<option>,<option>", d.Id())
	}
	if hasOptions {
		managed := make(map[string]interface{})
		for _, path := range strings.Split(optionsList, ",") {
			if path == "" {
				return nil, fmt.Errorf("invalid id %q, empty option name", d.Id())
			}
			addConfigValuesPath(managed, strings.Split(path, "."))
		}
		managedJSON, err := json.Marshal(managed)
		if err != nil {
			return nil, fmt.Errorf("marshaling values: %w", err)
		}
		if tfErr := d.Set(skValues, string(managedJSON)); tfErr != nil {
			panic(tfErr)
		}
	}
	if tfErr := d.Set("config_id", configID); tfErr != nil {
		panic(tfErr)
	}
	d.SetId(configID)

	return []*schema.ResourceData{d}, nil
}

func updateConfigoption(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	var values map[string]interface{}
	if err := json.Unmarshal([]byte(d.Get(skValues).(string)), &values); err != nil {
		return fmt.Errorf("unmarshaling values: %w", err)
	}
	configID := d.Get("config_id").(string)
	options := make([]string, 0, len(values))
	for k := range values {
		options = append(options, k)
	}
	slices.Sort(options)
	cfg, code, err := getConfigoption(ctx, m, configID, options, true)
	if err != nil {
		return err
	}
	if code == http.StatusNotFound {
		return fmt.Errorf("configuration %s doesn't exist", configID)
	}
	if err = prepareConfigValues(values, cfg.Options, ""); err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx,
		"/configoptions/"+url.PathEscape(configID), http.MethodPut, values)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

// getConfigoption reads a configuration file, restricted to the given options if any,
// with the type of the options if metainfo. A 404 is returned as code without error.
func getConfigoption(
	ctx context.Context, m interface{}, configID string, options []string, metainfo bool,
) (
	jsonConfigOptions, int, error,
) {
	c := m.(*Client)
	var result jsonConfigOptions
	var params []string
	if len(options) > 0 {
		escaped := make([]string, len(options))
		for i, v := range options {
			escaped[i] = url.QueryEscape(v)
		}
		params = append(params, "options="+strings.Join(escaped, ","))
	}
	if metainfo {
		params = append(params, "metainfo=true")
	}
	uri := "/configoptions/" + url.PathEscape(configID)
	if len(params) > 0 {
		uri += "?" + strings.Join(params, "&")
	}
	body, code, err := c.newRequest(ctx, uri, http.MethodGet, nil)
	if err != nil {
		return result, code, err
	}
	if code == http.StatusNotFound {
		return result, code, nil
	}
	if code != http.StatusOK {
		return result, code, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return result, code, fmt.Errorf("unmarshaling json: %w", err)
	}

	return result, code, nil
}

// configOptionsToValues converts the options returned by GET (a list of sections and options
// with a name) to the object of values expected by PUT.
func configOptionsToValues(options []interface{}) configSection {
	values := make(configSection, len(options))
	for _, o := range options {
		item, ok := o.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := item["name"].(string)
		if name == "" {
			continue
		}
		if v, isOption := item[skValue]; isOption {
			values[name] = v
		} else if sub, isSection := item["options"].([]interface{}); isSection {
			values[name] = configOptionsToValues(sub)
		}
	}

	return values
}

// prepareConfigValues adapts values to the type of the options, given by the options returned
// by GET with metainfo: a number is sent as a float for a float option and as a string for a string
// or option (choice) option, as jsonencode can't produce 60.0 and the API rejects 60 for "60".
// A section must be set with an object and an option can't be, the API returns an internal
// error otherwise.
func prepareConfigValues(values map[string]interface{}, options []interface{}, parent string) error {
	for _, o := range options {
		item, ok := o.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := item["name"].(string)
		v, isSet := values[name]
		if !isSet {
			continue
		}
		if _, isOption := item[skValue]; isOption {
			if _, isObject := v.(map[string]interface{}); isObject {
				return fmt.Errorf("%s%s is an option, it can't be set with an object", parent, name)
			}
			f, isNumber := v.(float64)
			if !isNumber {
				continue
			}
			switch item[skType] {
			case "float":
				values[name] = configFloat(f)
			case "string", "option":
				values[name] = strconv.FormatFloat(f, 'f', -1, 64)
			}
		} else if sub, isSection := item["options"].([]interface{}); isSection {
			subValues, isObject := v.(map[string]interface{})
			if !isObject {
				return fmt.Errorf("%s%s is a section, it must be set with an object of options", parent, name)
			}
			if err := prepareConfigValues(subValues, sub, parent+name+"."); err != nil {
				return err
			}
		}
	}

	return nil
}

// filterConfigValues keeps only the options of current that are in managed, so options
// not managed by the resource don't show up as a diff. An option missing on the API is dropped.
func filterConfigValues(current configSection, managed map[string]interface{}) configSection {
	result := make(configSection, len(managed))
	for k, mv := range managed {
		cv, ok := current[k]
		if !ok {
			continue
		}
		mSub, mIsObject := mv.(map[string]interface{})
		cSub, cIsSection := cv.(configSection)
		if mIsObject && cIsSection {
			result[k] = filterConfigValues(cSub, mSub)
		} else {
			result[k] = cv
		}
	}

	return result
}

// addConfigValuesPath adds a null option to values at path, creating its sections.
func addConfigValuesPath(values map[string]interface{}, path []string) {
	if len(path) == 1 {
		if _, ok := values[path[0]]; !ok {
			values[path[0]] = nil
		}

		return
	}
	sub, ok := values[path[0]].(map[string]interface{})
	if !ok {
		sub = make(map[string]interface{})
		values[path[0]] = sub
	}
	addConfigValuesPath(sub, path[1:])
}

func validateConfigoptionValues(v interface{}, path cty.Path) diag.Diagnostics {
	var values map[string]interface{}
	if err := json.Unmarshal([]byte(v.(string)), &values); err != nil {
		return diag.Diagnostics{{
			Severity:      diag.Error,
			Summary:       "values must be a JSON object",
			Detail:        err.Error(),
			AttributePath: path,
		}}
	}
	if len(values) == 0 {
		return diag.FromErr(errors.New("values must contain at least one option"))
	}

	return nil
}

// The configuration can be set by id or name: changing one for the other is not a change.
func suppressSameConfigID(_, oldValue, newValue string, d *schema.ResourceData) bool {
	if oldValue == "" || d.Id() == "" {
		return false
	}

	return newValue == d.Id() || newValue == d.Get("config_name").(string)
}

// suppressEquivalentJSON compares values as JSON, a number being equal to the same number as a
// string: a number set for a string option is sent and read back as a string.
func suppressEquivalentJSON(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	var o, n interface{}
	if err := json.Unmarshal([]byte(oldValue), &o); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(newValue), &n); err != nil {
		return false
	}

	return equivalentConfigValues(o, n)
}

func equivalentConfigValues(a, b interface{}) bool {
	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			w, ok := bv[k]
			if !ok || !equivalentConfigValues(v, w) {
				return false
			}
		}

		return true
	case float64:
		if bs, ok := b.(string); ok {
			return strconv.FormatFloat(av, 'f', -1, 64) == bs
		}
	case string:
		if bf, ok := b.(float64); ok {
			return strconv.FormatFloat(bf, 'f', -1, 64) == av
		}
	}

	return reflect.DeepEqual(a, b)
}
