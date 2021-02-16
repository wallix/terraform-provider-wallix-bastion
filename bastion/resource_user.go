package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type jsonUser struct {
	ForceChangePwd    *bool     `json:"force_change_pwd,omitempty"`
	IsDisabled        bool      `json:"is_disabled"`
	UserName          string    `json:"user_name,"`
	CertificateCN     string    `json:"certificate_dn"`
	DisplayName       string    `json:"display_name"`
	Email             string    `json:"email"`
	ExpirationDate    string    `json:"expiration_date"`
	IPSource          string    `json:"ip_source"`
	Password          string    `json:"password,omitempty"`
	PreferredLanguage string    `json:"preferred_language,omitempty"`
	Profile           string    `json:"profile"`
	SSHPublicKey      string    `json:"ssh_public_key"`
	UserAuths         []string  `json:"user_auths"`
	Groups            *[]string `json:"groups,omitempty"`
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			State: resourceUserImport,
		},
		Schema: map[string]*schema.Schema{
			"user_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"profile": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_auths": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"certificate_dn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"expiration_date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"force_change_pwd": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"ip_source": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"preferred_language": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"de", "en", "es", "fr", "ru"}, false),
				Computed:     true,
			},
			"ssh_public_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourveUserVersionCheck(version string) error {
	if version == versionValidate3_3 {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_user not validate with api version %s", version)
}
func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveUserVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	ex, err := checkResourceUserExists(ctx, d.Get("user_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("user_name %s already exists", d.Get("user_name").(string)))
	}
	err = addUser(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(d.Get("user_name").(string))

	return resourceUserRead(ctx, d, m)
}
func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveUserVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readUserOptions(ctx, d.Get("user_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.UserName == "" {
		d.SetId("")
	} else {
		fillUser(d, cfg)
	}

	return nil
}
func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourveUserVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateUser(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceUserRead(ctx, d, m)
}
func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	if err := resourveUserVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteUser(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceUserImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourveUserVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	ex, err := checkResourceUserExists(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find user_name with id %s (id must be <user_name>", d.Id())
	}
	cfg, err := readUserOptions(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	fillUser(d, cfg)
	result := make([]*schema.ResourceData, 1)
	result[0] = d

	return result, nil
}

func checkResourceUserExists(ctx context.Context, userName string, m interface{}) (bool, error) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/users/"+userName, http.MethodGet, nil)
	if err != nil {
		return false, err
	}
	if code == http.StatusNotFound {
		return false, nil
	}
	if code != http.StatusOK {
		return false, fmt.Errorf("api doesn't return OK : %d with body :\n%s", code, body)
	}

	return true, nil
}

func addUser(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	json := prepareUserJSON(d, true)
	body, code, err := c.newRequest(ctx, "/users/", http.MethodPost, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func updateUser(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	json := prepareUserJSON(d, false)
	body, code, err := c.newRequest(ctx, "/users/"+d.Get("user_name").(string)+"?force=true", http.MethodPut, json)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}
func deleteUser(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/users/"+d.Get("user_name").(string), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent : %d with body :\n%s", code, body)
	}

	return nil
}

func prepareUserJSON(d *schema.ResourceData, newResource bool) jsonUser {
	b := true
	user := jsonUser{
		UserName:       d.Get("user_name").(string),
		DisplayName:    d.Get("display_name").(string),
		Email:          d.Get("email").(string),
		IPSource:       d.Get("ip_source").(string),
		Profile:        d.Get("profile").(string),
		SSHPublicKey:   d.Get("ssh_public_key").(string),
		CertificateCN:  d.Get("certificate_dn").(string),
		ExpirationDate: d.Get("expiration_date").(string),
		IsDisabled:     d.Get("is_disabled").(bool),
	}
	if newResource {
		user.PreferredLanguage = d.Get("preferred_language").(string)
		user.Password = d.Get("password").(string)
		if d.Get("force_change_pwd").(bool) {
			user.ForceChangePwd = &b
		}
		if d.Get("groups") != nil {
			groups := make([]string, 0)
			for _, v := range d.Get("groups").(*schema.Set).List() {
				groups = append(groups, v.(string))
			}
			user.Groups = &groups
		}
	}
	if d.HasChanges("groups") {
		groups := make([]string, 0)
		for _, v := range d.Get("groups").(*schema.Set).List() {
			groups = append(groups, v.(string))
		}
		user.Groups = &groups
	}
	for _, v := range d.Get("user_auths").(*schema.Set).List() {
		user.UserAuths = append(user.UserAuths, v.(string))
	}

	return user
}

func readUserOptions(ctx context.Context, userName string, m interface{}) (jsonUser, error) {
	c := m.(*Client)
	var result jsonUser
	body, code, err := c.newRequest(ctx, "/users/"+userName, http.MethodGet, nil)
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
		return result, err
	}

	return result, nil
}

func fillUser(d *schema.ResourceData, json jsonUser) {
	if tfErr := d.Set("email", json.Email); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("profile", json.Profile); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("user_auths", json.UserAuths); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("certificate_dn", json.CertificateCN); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("display_name", json.DisplayName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("expiration_date", json.ExpirationDate); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("force_change_pwd", *json.ForceChangePwd); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("groups", json.Groups); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ip_source", json.IPSource); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_disabled", json.IsDisabled); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("preferred_language", json.PreferredLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ssh_public_key", json.SSHPublicKey); tfErr != nil {
		panic(tfErr)
	}
}
