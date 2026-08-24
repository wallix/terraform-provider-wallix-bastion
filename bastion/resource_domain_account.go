package bastion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			StateContext: resourceDomainAccountImport,
		},
		Schema: map[string]*schema.Schema{
			skDomainID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			skAccountName: {
				Type:     schema.TypeString,
				Required: true,
			},
			skAccountLogin: {
				Type:     schema.TypeString,
				Required: true,
			},
			skAutoChangePassword: {
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
			skCheckoutPolicy: {
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
						skPublicKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			skDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			skDomainPasswordChange: {
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
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_domain_account not available with api version %s", version)
}

func resourceDomainAccountCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfgDomain, err := readDomainOptions(ctx, d.Get(skDomainID).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDomain.ID == "" {
		return diag.FromErr(fmt.Errorf("domain_id with ID %s doesn't exists", d.Get(skDomainID).(string)))
	}
	_, ex, err := searchResourceDomainAccount(ctx, d.Get(skDomainID).(string), d.Get(skAccountName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s already exists",
			d.Get(skAccountName).(string), d.Get(skDomainID).(string)))
	}
	id, err := addDomainAccount(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceDomainAccount(ctx, d.Get(skDomainID).(string), d.Get(skAccountName).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s not found after POST",
				d.Get(skAccountName).(string), d.Get(skDomainID).(string)))
		}
	}
	d.SetId(id)

	return resourceDomainAccountRead(ctx, d, m)
}

func resourceDomainAccountRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDomainAccountOptions(ctx, d.Get(skDomainID).(string), d.Id(), m)
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

func resourceDomainAccountUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
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

func resourceDomainAccountDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDomainAccount(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDomainAccountImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, errors.New("id must be <domain_id>/<account_name>")
	}
	id, ex, err := searchResourceDomainAccount(ctx, idSplit[0], idSplit[1], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find account_name with id %s "+
			"(id must be <domain_id>/<account_name>)", d.Id())
	}
	cfg, err := readDomainAccountOptions(ctx, idSplit[0], id, m)
	if err != nil {
		return nil, err
	}
	fillDomainAccount(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set(skDomainID, idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceDomainAccount(
	ctx context.Context, domainID, accountName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/domains/"+domainID+"/accounts/?q=account_name="+accountName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonDomainAccount
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addDomainAccount(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	jsonData, err := prepareDomainAccountJSON(d)
	if err != nil {
		return "", err
	}
	body, headers, code, err := c.newRequestWithHeaders(ctx,
		"/domains/"+d.Get(skDomainID).(string)+"/accounts/", http.MethodPost, jsonData)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateDomainAccount(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData, err := prepareDomainAccountJSON(d)
	if err != nil {
		return err
	}
	body, code, err := c.newRequest(ctx,
		"/domains/"+d.Get(skDomainID).(string)+"/accounts/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteDomainAccount(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/domains/"+d.Get(skDomainID).(string)+"/accounts/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareDomainAccountJSON(d *schema.ResourceData) (jsonDomainAccount, error) {
	jsonData := jsonDomainAccount{
		AccountLogin:        d.Get(skAccountLogin).(string),
		AccountName:         d.Get(skAccountName).(string),
		AutoChangePassword:  d.Get(skAutoChangePassword).(bool),
		AutoChangeSSHKey:    d.Get("auto_change_ssh_key").(bool),
		CertificateValidity: d.Get("certificate_validity").(string),
		CheckoutPolicy:      d.Get(skCheckoutPolicy).(string),
		Description:         d.Get(skDescription).(string),
	}

	listResources := d.Get("resources").(*schema.Set).List()
	if len(listResources) > 0 {
		resources := make([]string, len(listResources))
		for i, v := range listResources {
			vSplt := strings.Split(v.(string), ":")
			if len(vSplt) != 2 {
				return jsonData, errors.New("resource must have format device:service or application:APP")
			}
			resources[i] = v.(string)
		}
		jsonData.Resources = &resources
	}

	return jsonData, nil
}

func readDomainAccountOptions(
	ctx context.Context, localDomainID, accountID string, m interface{},
) (
	jsonDomainAccount, error,
) {
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
		return result, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, fmt.Errorf("unmarshaling json: %w", err)
	}

	return result, nil
}

func fillDomainAccount(d *schema.ResourceData, jsonData jsonDomainAccount) {
	if tfErr := d.Set(skAccountName, jsonData.AccountName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAccountLogin, jsonData.AccountLogin); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skCheckoutPolicy, jsonData.CheckoutPolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAutoChangePassword, jsonData.AutoChangePassword); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auto_change_ssh_key", jsonData.AutoChangeSSHKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("certificate_validity", jsonData.CertificateValidity); tfErr != nil {
		panic(tfErr)
	}
	credentials := make([]map[string]interface{}, 0)
	if jsonData.Credentials != nil {
		credentials = make([]map[string]interface{}, len(*jsonData.Credentials))
		for i, v := range *jsonData.Credentials {
			credentials[i] = map[string]interface{}{
				"id":        v.ID,
				skPublicKey: v.PublicKey,
				skType:      v.Type,
			}
		}
	}
	if tfErr := d.Set("credentials", credentials); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDomainPasswordChange, jsonData.DomainPasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.Resources == nil {
		if tfErr := d.Set("resources", []string{}); tfErr != nil {
			panic(tfErr)
		}
	} else {
		if tfErr := d.Set("resources", *jsonData.Resources); tfErr != nil {
			panic(tfErr)
		}
	}
}
