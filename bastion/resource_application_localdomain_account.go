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

type jsonApplicationLocalDomainAccount struct {
	ID                   string           `json:"id,omitempty"`
	AccountName          string           `json:"account_name"`
	AccountLogin         string           `json:"account_login"`
	Description          string           `json:"description"`
	DomainPasswordChange *bool            `json:"domain_password_change,omitempty"`
	AutoChangePassword   bool             `json:"auto_change_password"`
	CheckoutPolicy       string           `json:"checkout_policy"`
	Credentials          []jsonCredential `json:"credentials"`
}

func resourceApplicationLocalDomainAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationLocalDomainAccountCreate,
		ReadContext:   resourceApplicationLocalDomainAccountRead,
		UpdateContext: resourceApplicationLocalDomainAccountUpdate,
		DeleteContext: resourceApplicationLocalDomainAccountDelete,
		Importer: &schema.ResourceImporter{
			State: resourceApplicationLocalDomainAccountImport,
		},
		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
			"checkout_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_password_change": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}
func resourceApplicationLocalDomainAccountVersionCheck(version string) error {
	if bchk.InSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_application_localdomain_account not available with api version %s", version)
}

func resourceApplicationLocalDomainAccountCreate(ctx context.Context,
	d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfgApplication, err := readApplicationOptions(ctx, d.Get("application_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgApplication.ID == "" {
		return diag.FromErr(fmt.Errorf("application with ID %s doesn't exists", d.Get("application_id").(string)))
	}
	cfgDomain, err := readApplicationLocalDomainOptions(ctx,
		d.Get("application_id").(string), d.Get("domain_id").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDomain.ID == "" {
		return diag.FromErr(fmt.Errorf("domain_id with ID %s on application_id %s doesn't exists",
			d.Get("domain_id").(string), d.Get("application_id").(string)))
	}
	_, ex, err := searchResourceApplicationLocalDomainAccount(ctx,
		d.Get("application_id").(string), d.Get("domain_id").(string), d.Get("account_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s, application_id %s already exists",
			d.Get("account_name").(string), d.Get("domain_id").(string), d.Get("application_id").(string)))
	}
	err = addApplicationLocalDomainAccount(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceApplicationLocalDomainAccount(ctx,
		d.Get("application_id").(string), d.Get("domain_id").(string), d.Get("account_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s, application_id %s not found after POST",
			d.Get("account_name").(string), d.Get("domain_id").(string), d.Get("application_id").(string)))
	}
	d.SetId(id)

	return resourceApplicationLocalDomainAccountRead(ctx, d, m)
}
func resourceApplicationLocalDomainAccountRead(ctx context.Context,
	d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readApplicationLocalDomainAccountOptions(ctx,
		d.Get("application_id").(string), d.Get("domain_id").(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillApplicationLocalDomainAccount(d, cfg)
	}

	return nil
}
func resourceApplicationLocalDomainAccountUpdate(ctx context.Context,
	d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceApplicationLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateApplicationLocalDomainAccount(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceApplicationLocalDomainAccountRead(ctx, d, m)
}
func resourceApplicationLocalDomainAccountDelete(ctx context.Context,
	d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceApplicationLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteApplicationLocalDomainAccount(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceApplicationLocalDomainAccountImport(
	d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceApplicationLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 3 {
		return nil, fmt.Errorf("id must be <application_id>/<domain_id>/<account_name>")
	}
	id, ex, err := searchResourceApplicationLocalDomainAccount(ctx, idSplit[0], idSplit[1], idSplit[2], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find account_name with id %s "+
			"(id must be <application_id>/<domain_id>/<account_name>", d.Id())
	}
	cfg, err := readApplicationLocalDomainAccountOptions(ctx, idSplit[0], idSplit[1], id, m)
	if err != nil {
		return nil, err
	}
	fillApplicationLocalDomainAccount(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set("application_id", idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_id", idSplit[1]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceApplicationLocalDomainAccount(ctx context.Context,
	applicationID, domainID, accountName string, m interface{}) (string, bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/applications/"+applicationID+"/localdomains/"+domainID+
		"/accounts/?q=account_name="+accountName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}
	var results []jsonApplicationLocalDomainAccount
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("json.Unmarshal failed : %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addApplicationLocalDomainAccount(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareApplicationLocalDomainAccountJSON(d)
	body, code, err := c.newRequest(ctx,
		"/applications/"+d.Get("application_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateApplicationLocalDomainAccount(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	jsonData := prepareApplicationLocalDomainAccountJSON(d)
	body, code, err := c.newRequest(ctx,
		"/applications/"+d.Get("application_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func deleteApplicationLocalDomainAccount(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/applications/"+d.Get("application_id").(string)+"/localdomains/"+d.Get("domain_id").(string)+
			"/accounts/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareApplicationLocalDomainAccountJSON(d *schema.ResourceData) jsonApplicationLocalDomainAccount {
	var jsonData jsonApplicationLocalDomainAccount
	jsonData.AccountName = d.Get("account_name").(string)
	jsonData.AccountLogin = d.Get("account_login").(string)
	jsonData.CheckoutPolicy = d.Get("checkout_policy").(string)
	jsonData.AutoChangePassword = d.Get("auto_change_password").(bool)
	jsonData.Description = d.Get("description").(string)
	credentials := make([]jsonCredential, 0)
	if d.Get("password").(string) != "" {
		credentials = append(credentials, jsonCredential{
			Type:     "password",
			Password: d.Get("password").(string),
		})
	}
	jsonData.Credentials = credentials

	return jsonData
}

func readApplicationLocalDomainAccountOptions(
	ctx context.Context, applicationID, localDomainID, accountID string, m interface{}) (
	jsonApplicationLocalDomainAccount, error) {
	c := m.(*Client)
	var result jsonApplicationLocalDomainAccount
	body, code, err := c.newRequest(ctx,
		"/applications/"+applicationID+"/localdomains/"+localDomainID+
			"/accounts/"+accountID, http.MethodGet, nil)
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

func fillApplicationLocalDomainAccount(d *schema.ResourceData, jsonData jsonApplicationLocalDomainAccount) {
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
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_password_change", jsonData.DomainPasswordChange); tfErr != nil {
		panic(tfErr)
	}
}
