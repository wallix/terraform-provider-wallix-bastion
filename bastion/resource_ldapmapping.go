package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	govers "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
)

type jsonLdapMapping struct {
	Domain    string `json:"domain"`
	UserGroup string `json:"user_group"`
	LdapGroup string `json:"ldap_group"`
}

func resourceLdapMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLdapMappingCreate,
		ReadContext:   resourceLdapMappingRead,
		DeleteContext: resourceLdapMappingDelete,
		Importer: &schema.ResourceImporter{
			State: resourceLdapMappingImport,
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ldap_group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLdapMappingVersionCheck(version string) error {
	if bchk.InSlice(version, []string{VersionWallixAPI33, VersionWallixAPI36}) {
		return nil
	}
	if vers, err := govers.NewVersion(version); err == nil {
		versionResourceRename, _ := govers.NewVersion(VersionWallixAPI38)
		if vers.GreaterThanOrEqual(versionResourceRename) {
			return fmt.Errorf(
				"resource wallix-bastion_ldapmapping not available with api version %s\n"+
					" use wallix-bastion_authdomain_mapping instead",
				version)
		}
	}

	return fmt.Errorf("resource wallix-bastion_ldapmapping not available with api version %s", version)
}

func resourceLdapMappingCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceLdapMappingVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	ex, err := checkResourceLdapMappingExists(ctx,
		d.Get("domain").(string), d.Get("user_group").(string), d.Get("ldap_group").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("ldapmapping %s/%s/%s already exists",
			d.Get("domain").(string), d.Get("user_group").(string), d.Get("ldap_group").(string)))
	}
	err = addLdapMapping(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(d.Get("domain").(string) + "/" + d.Get("user_group").(string) + "/" + d.Get("ldap_group").(string))

	return nil
}

func resourceLdapMappingRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceLdapMappingVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	ex, err := checkResourceLdapMappingExists(ctx,
		d.Get("domain").(string), d.Get("user_group").(string), d.Get("ldap_group").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		d.SetId("")
	}

	return nil
}

func resourceLdapMappingDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceLdapMappingVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteLdapMapping(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceLdapMappingImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceLdapMappingVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 3 {
		return nil, fmt.Errorf("id must be <domain>/<user_group>/<ldap_group>")
	}
	ex, err := checkResourceLdapMappingExists(ctx, idSplit[0], idSplit[1], idSplit[2], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find ldapmapping with id %s (id must be <domain>/<user_group>/<ldap_group>", d.Id())
	}
	cfg := jsonLdapMapping{
		Domain:    idSplit[0],
		UserGroup: idSplit[1],
		LdapGroup: idSplit[2],
	}
	fillLdapMapping(d, cfg)
	result := make([]*schema.ResourceData, 1)
	result[0] = d

	return result, nil
}

func checkResourceLdapMappingExists(
	ctx context.Context, domain, userGroup, ldapGroup string, m interface{},
) (
	bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/ldapmappings/?q=domain="+domain+url.QueryEscape("&&")+"user_group="+userGroup, http.MethodGet, nil)
	if err != nil {
		return false, err
	}
	if code != http.StatusOK {
		return false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonLdapMapping
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	if len(results) == 1 && results[0].LdapGroup == ldapGroup {
		return true, nil
	}

	return false, nil
}

func addLdapMapping(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareLdapMappingJSON(d)
	body, code, err := c.newRequest(ctx, "/ldapmappings/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteLdapMapping(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 3 {
		return fmt.Errorf("id must be <domain>/<user_group>/<ldap_group>")
	}
	body, code, err := c.newRequest(ctx, "/ldapmappings/"+idSplit[0]+"/"+idSplit[1]+"/"+idSplit[2], http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareLdapMappingJSON(d *schema.ResourceData) jsonLdapMapping {
	return jsonLdapMapping{
		Domain:    d.Get("domain").(string),
		UserGroup: d.Get("user_group").(string),
		LdapGroup: d.Get("ldap_group").(string),
	}
}

func fillLdapMapping(d *schema.ResourceData, jsonData jsonLdapMapping) {
	if tfErr := d.Set("domain", jsonData.Domain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("user_group", jsonData.UserGroup); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ldap_group", jsonData.LdapGroup); tfErr != nil {
		panic(tfErr)
	}
}
