package bastion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
)

type jsonAuthDomain struct {
	ID string `json:"id,omitempty"`
}

type jsonAuthDomainMapping struct {
	ID            string `json:"id,omitempty"`
	Domain        string `json:"domain,omitempty"`
	UserGroup     string `json:"user_group"`
	ExternalGroup string `json:"external_group"`
}

func resourceAuthDomainMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuthDomainMappingCreate,
		ReadContext:   resourceAuthDomainMappingRead,
		UpdateContext: resourceAuthDomainMappingUpdate,
		DeleteContext: resourceAuthDomainMappingDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAuthDomainMappingImport,
		},
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_group": {
				Type:     schema.TypeString,
				Required: true,
			},
			"external_group": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAuthDomainMappingVersionCheck(version string) error {
	if bchk.InSlice(version, versions38Plus()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_authdomain_mapping not available with api version %s", version)
}

func resourceAuthDomainMappingCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainMappingVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	domainIDExists, err := checkAuthDomainID(ctx, d.Get("domain_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !domainIDExists {
		return diag.FromErr(fmt.Errorf("auth domain with ID %s doesn't exists", d.Get("domain_id").(string)))
	}
	_, ex, err := searchResourceAuthDomainMapping(ctx, d.Get("domain_id").(string), d.Get("user_group").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("auth domain mapping for user_group %s on domain_id %s already exists",
			d.Get("user_group").(string), d.Get("domain_id").(string)))
	}
	err = addAuthDomainMapping(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthDomainMapping(ctx, d.Get("domain_id").(string), d.Get("user_group").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("auth domain mapping for user_group %s on domain_id %s not found after POST",
			d.Get("user_group").(string), d.Get("domain_id").(string)))
	}
	d.SetId(id)

	return resourceAuthDomainMappingRead(ctx, d, m)
}

func resourceAuthDomainMappingRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainMappingVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readAuthDomainMappingOptions(ctx, d.Get("domain_id").(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillAuthDomainMapping(d, cfg)
	}

	return nil
}

func resourceAuthDomainMappingUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceAuthDomainMappingVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateAuthDomainMapping(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceAuthDomainMappingRead(ctx, d, m)
}

func resourceAuthDomainMappingDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthDomainMappingVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteAuthDomainMapping(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAuthDomainMappingImport(
	d *schema.ResourceData, m interface{},
) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceAuthDomainMappingVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, errors.New("id must be <domain_id>/<user_group>")
	}
	id, ex, err := searchResourceAuthDomainMapping(ctx, idSplit[0], idSplit[1], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find auth domain mapping with id %s (id must be <domain_id>/<user_group>)", d.Id())
	}
	cfg, err := readAuthDomainMappingOptions(ctx, idSplit[0], id, m)
	if err != nil {
		return nil, err
	}
	fillAuthDomainMapping(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set("domain_id", idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func checkAuthDomainID(
	ctx context.Context, domainID string, m interface{},
) (
	bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/authdomains/"+domainID, http.MethodGet, nil)
	if err != nil {
		return false, err
	}
	if code == http.StatusNotFound {
		return false, nil
	}
	if code != http.StatusOK {
		return false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var result jsonAuthDomain
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if result.ID == domainID {
		return true, nil
	}

	return false, nil
}

func searchResourceAuthDomainMapping(
	ctx context.Context, domainID, userGroup string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(
		ctx,
		"/authdomains/"+domainID+"/mappings/?q=user_group="+userGroup,
		http.MethodGet,
		nil,
	)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonAuthDomainMapping
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addAuthDomainMapping(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAuthDomainMappingJSON(d)
	body, code, err := c.newRequest(
		ctx,
		"/authdomains/"+d.Get("domain_id").(string)+"/mappings",
		http.MethodPost,
		jsonData,
	)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateAuthDomainMapping(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAuthDomainMappingJSON(d)
	body, code, err := c.newRequest(
		ctx,
		"/authdomains/"+d.Get("domain_id").(string)+"/mappings/"+d.Id(),
		http.MethodPut,
		jsonData,
	)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteAuthDomainMapping(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(
		ctx,
		"/authdomains/"+d.Get("domain_id").(string)+"/mappings/"+d.Id(),
		http.MethodDelete,
		nil,
	)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareAuthDomainMappingJSON(d *schema.ResourceData) jsonAuthDomainMapping {
	return jsonAuthDomainMapping{
		UserGroup:     d.Get("user_group").(string),
		ExternalGroup: d.Get("external_group").(string),
	}
}

func readAuthDomainMappingOptions(
	ctx context.Context, domainID, mappingID string, m interface{},
) (
	jsonAuthDomainMapping, error,
) {
	c := m.(*Client)
	var result jsonAuthDomainMapping
	body, code, err := c.newRequest(
		ctx,
		"/authdomains/"+domainID+"/mappings/"+mappingID,
		http.MethodGet,
		nil,
	)
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

func fillAuthDomainMapping(d *schema.ResourceData, jsonData jsonAuthDomainMapping) {
	if tfErr := d.Set("user_group", jsonData.UserGroup); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("external_group", jsonData.ExternalGroup); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain", jsonData.Domain); tfErr != nil {
		panic(tfErr)
	}
}
