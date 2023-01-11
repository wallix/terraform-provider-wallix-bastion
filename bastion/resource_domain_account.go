package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
)

type jsonDomainAccount struct {
	ID                   string            `json:"id,omitempty"`
	AccountName          string            `json:"account_name"`
	AccountLogin         string            `json:"account_login"`
	Description          string            `json:"description"`
	DomainPasswordChange *bool             `json:"domain_password_change,omitempty"`
	AutoChangePassword   bool              `json:"auto_change_password"`
	AutoChangeSSHKey     bool              `json:"auto_change_ssh_key"`
	CheckoutPolicy       string            `json:"checkout_policy"`
	CertificateValidity  string            `json:"certificate_validity,omitempty"`
	Resources            *[]string         `json:"resources,omitempty"`
	Credentials          *[]jsonCredential `json:"credentials,omitempty"`
}

func resourceDomainAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainAccountCreate,
		ReadContext:   resourceDomainAccountRead,
		UpdateContext: resourceDomainAccountUpdate,
		DeleteContext: resourceDomainAccountDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDomainAccountImport,
		},
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_login": {
				Type:     schema.TypeString,
				Required: true,
			},
			"auto_change_password": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"auto_change_ssh_key": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"certificate_validity": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"checkout_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"credentials": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_password_change": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"resources": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}
func resourceDomainAccountVersionCheck(version string) error {
	if bchk.InSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_domain_account not available with api version %s", version)
}

func resourceDomainAccountCreate(ctx context.Context,
	d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfgDomain, err := readDomainOptions(ctx, d.Get("domain_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDomain.ID == "" {
		return diag.FromErr(fmt.Errorf("domain_id with ID %s doesn't exists", d.Get("domain_id").(string)))
	}
	_, ex, err := searchResourceDomainAccount(ctx, d.Get("domain_id").(string), d.Get("account_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s already exists",
			d.Get("account_name").(string), d.Get("domain_id").(string)))
	}
	err = addDomainAccount(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDomainAccount(ctx, d.Get("domain_id").(string), d.Get("account_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s not found after POST",
			d.Get("account_name").(string), d.Get("domain_id").(string)))
	}
	d.SetId(id)

	return resourceDomainAccountRead(ctx, d, m)
}
func resourceDomainAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDomainAccountOptions(ctx, d.Get("domain_id").(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillDomainAccount(d, cfg)
	}

	return nil
}
func resourceDomainAccountUpdate(ctx context.Context,
	d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDomainAccount(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDomainAccountRead(ctx, d, m)
}
func resourceDomainAccountDelete(ctx context.Context,
	d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDomainAccount(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceDomainAccountImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, fmt.Errorf("id must be <domain_id>/<account_name>")
	}
	id, ex, err := searchResourceDomainAccount(ctx, idSplit[0], idSplit[1], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find account_name with id %s "+
			"(id must be <domain_id>/<account_name>", d.Id())
	}
	cfg, err := readDomainAccountOptions(ctx, idSplit[0], id, m)
	if err != nil {
		return nil, err
	}
	fillDomainAccount(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set("domain_id", idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceDomainAccount(ctx context.Context,
	domainID, accountName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/domains/"+domainID+"/accounts/?q=account_name="+accountName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonDomainAccount
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addDomainAccount(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData, err := prepareDomainAccountJSON(d)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx,
		"/domains/"+d.Get("domain_id").(string)+"/accounts/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateDomainAccount(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData, err := prepareDomainAccountJSON(d)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx,
		"/domains/"+d.Get("domain_id").(string)+"/accounts/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteDomainAccount(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/domains/"+d.Get("domain_id").(string)+"/accounts/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareDomainAccountJSON(d *schema.ResourceData) (jsonDomainAccount, error) {
	var jsonData jsonDomainAccount
	jsonData.AccountName = d.Get("account_name").(string)
	jsonData.AccountLogin = d.Get("account_login").(string)
	jsonData.CheckoutPolicy = d.Get("checkout_policy").(string)
	jsonData.AutoChangePassword = d.Get("auto_change_password").(bool)
	jsonData.AutoChangeSSHKey = d.Get("auto_change_ssh_key").(bool)
	jsonData.CertificateValidity = d.Get("certificate_validity").(string)
	jsonData.Description = d.Get("description").(string)
	if d.HasChange("resources") {
		resources := make([]string, 0)
		for _, v := range d.Get("resources").(*schema.Set).List() {
			vSplt := strings.Split(v.(string), ":")
			if len(vSplt) != 2 {
				return jsonData, fmt.Errorf("resource must have format device:service or application:APP")
			}
			resources = append(resources, v.(string))
		}
		jsonData.Resources = &resources
	}

	return jsonData, nil
}

func readDomainAccountOptions(
	ctx context.Context, localDomainID, accountID string, m interface{}) (
	jsonDomainAccount, error) {
	c := m.(*Client)
	var result jsonDomainAccount
	body, code, err := c.newRequest(ctx, "/domains/"+localDomainID+"/accounts/"+accountID, http.MethodGet, nil)
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

func fillDomainAccount(d *schema.ResourceData, jsonData jsonDomainAccount) {
	if tfErr := d.Set("account_name", jsonData.AccountName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("account_login", jsonData.AccountLogin); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("checkout_policy", jsonData.CheckoutPolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auto_change_password", jsonData.AutoChangePassword); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auto_change_ssh_key", jsonData.AutoChangeSSHKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("certificate_validity", jsonData.CertificateValidity); tfErr != nil {
		panic(tfErr)
	}
	credentials := make([]map[string]interface{}, 0)
	for _, v := range *jsonData.Credentials {
		credentials = append(credentials, map[string]interface{}{
			"id":         v.ID,
			"public_key": v.PublicKey,
			"type":       v.Type,
		})
	}
	if tfErr := d.Set("credentials", credentials); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_password_change", jsonData.DomainPasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("resources", jsonData.Resources); tfErr != nil {
		panic(tfErr)
	}
}
