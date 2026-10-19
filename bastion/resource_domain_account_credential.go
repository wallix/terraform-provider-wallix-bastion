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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDomainAccountCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainAccountCredentialCreate,
		ReadContext:   resourceDomainAccountCredentialRead,
		UpdateContext: resourceDomainAccountCredentialUpdate,
		DeleteContext: resourceDomainAccountCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainAccountCredentialImport,
		},
		CustomizeDiff: validateDomainAccountCredentialPropagate,
		Schema: map[string]*schema.Schema{
			skDomainID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			skAccountID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			skType: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{skPassword, "ssh_key"}, false),
			},
			skPassphrase: {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				RequiredWith: []string{skPrivateKey},
			},
			skPassword: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			skPrivateKey: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			skPublicKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skPropagateCredentialChange: {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceDomainAccountCredentialVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_domain_account_credential not available with api version %s", version)
}

func resourceDomainAccountCredentialCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfgDomain, err := readDomainOptions(ctx, d.Get(skDomainID).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgDomain.ID == "" {
		return diag.FromErr(fmt.Errorf("domain_id with ID %s doesn't exists", d.Get(skDomainID).(string)))
	}
	cfgAccount, err := readDomainAccountOptions(ctx, d.Get(skDomainID).(string), d.Get(skAccountID).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfgAccount.ID == "" {
		return diag.FromErr(fmt.Errorf("account_id with ID %s on domain_id %s doesn't exists",
			d.Get(skAccountID).(string), d.Get(skDomainID).(string)))
	}
	_, ex, err := searchResourceDomainAccountCredential(ctx,
		d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Get(skType).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("credential type %s on account_id %s, domain_id %s already exists",
			d.Get(skType).(string), d.Get(skAccountID).(string), d.Get(skDomainID).(string)))
	}
	id, err := addDomainAccountCredential(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		// Fallback for Bastion versions that don't return the X-Object-Id header on creation.
		id, ex, err = searchResourceDomainAccountCredential(ctx,
			d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Get(skType).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if !ex {
			return diag.FromErr(fmt.Errorf(
				"credential type %s on account_id %s, domain_id %s not found after POST",
				d.Get(skType).(string), d.Get(skAccountID).(string), d.Get(skDomainID).(string)))
		}
	}
	d.SetId(id)
	// The credential exists on the Bastion from here: if the propagation fails, the resource is
	// saved as tainted and replaced on the next apply, instead of leaving it out of state.
	if d.Get(skPropagateCredentialChange).(bool) {
		if err := propagateDomainAccountCredential(ctx, d, m); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDomainAccountCredentialRead(ctx, d, m)
}

func resourceDomainAccountCredentialRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readDomainAccountCredentialOptions(ctx,
		d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		// WALLIX rotates a credential (manual regenerate, or an automatic
		// rotation policy) by deleting the old object and creating a new one
		// with a different internal ID, so the GET by the stored ID 404s.
		// Retry a lookup by (account, type) before treating it as deleted,
		// as the device and application variants do.
		newID, found, err := searchResourceDomainAccountCredential(ctx,
			d.Get(skDomainID).(string), d.Get(skAccountID).(string), d.Get(skType).(string), m)
		if err != nil {
			return diag.FromErr(err)
		}
		if found {
			cfg, err = readDomainAccountCredentialOptions(ctx,
				d.Get(skDomainID).(string), d.Get(skAccountID).(string), newID, m)
			if err != nil {
				return diag.FromErr(err)
			}
			if cfg.ID != "" {
				d.SetId(newID)
			}
		}
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillDomainAccountCredential(d, cfg)
	}

	return nil
}

func resourceDomainAccountCredentialUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateDomainAccountCredential(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceDomainAccountCredentialRead(ctx, d, m)
}

func resourceDomainAccountCredentialDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteDomainAccountCredential(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDomainAccountCredentialImport(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	c := m.(*Client)
	if err := resourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 3 {
		return nil, errors.New("id must be <domain_id>/<account_id>/<type>")
	}
	id, ex, err := searchResourceDomainAccountCredential(ctx, idSplit[0], idSplit[1], idSplit[2], m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find credential with id %s "+
			"(id must be <domain_id>/<account_id>/<type>)", d.Id())
	}
	cfg, err := readDomainAccountCredentialOptions(ctx, idSplit[0], idSplit[1], id, m)
	if err != nil {
		return nil, err
	}
	fillDomainAccountCredential(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	if tfErr := d.Set(skDomainID, idSplit[0]); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAccountID, idSplit[1]); tfErr != nil {
		panic(tfErr)
	}
	result[0] = d

	return result, nil
}

func searchResourceDomainAccountCredential(
	ctx context.Context, domainID, accountID, typeCred string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/domains/"+domainID+"/accounts/"+accountID+
			"/credentials/", http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	// The account no longer exists: neither does its credential.
	if code == http.StatusNotFound {
		return "", false, nil
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonCredential
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	for _, v := range results {
		if v.Type == typeCred {
			return v.ID, true, nil
		}
	}

	return "", false, nil
}

func addDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) (string, error) {
	c := m.(*Client)
	body, headers, code, err := c.newRequestWithHeaders(ctx,
		"/domains/"+d.Get(skDomainID).(string)+"/accounts/"+d.Get(skAccountID).(string)+"/credentials/",
		http.MethodPost, prepareDomainAccountCredentialJSON(d))
	if err != nil {
		return "", err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return "", fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return headers.Get("X-Object-Id"), nil
}

func updateDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	if d.Get(skPropagateCredentialChange).(bool) {
		// Only a new value is pushed to the target host: enabling the propagation alone changes
		// nothing on it.
		if !d.HasChanges(skPassword, skPassphrase) {
			return nil
		}

		return propagateDomainAccountCredential(ctx, d, m)
	}
	c := m.(*Client)
	body, code, err := c.newRequest(ctx,
		"/domains/"+d.Get(skDomainID).(string)+"/accounts/"+d.Get(skAccountID).(string)+"/credentials/"+d.Id(),
		http.MethodPut, prepareDomainAccountCredentialJSON(d))
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

// propagateDomainAccountCredential changes the credential on the Bastion and on the target host of
// the account (PUT /accountchangepassword/<account_id>/<password|ssh_key>). Per the API, an empty
// body makes the Bastion generate a new value and change it on the target host: the configured
// value is always sent, and a missing one is an error instead of an unrequested change.
func propagateDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	credentialType := d.Get(skType).(string)
	var jsonData jsonCredential
	if credentialType == skPassword {
		jsonData.Password = d.Get(skPassword).(string)
		if jsonData.Password == "" {
			return errors.New(errPropagateWithoutPassword)
		}
	} else {
		jsonData.PrivateKey = d.Get(skPrivateKey).(string)
		jsonData.Passphrase = d.Get(skPassphrase).(string)
		if jsonData.PrivateKey == "" || strings.HasPrefix(jsonData.PrivateKey, "generate:") {
			return errors.New(errPropagateWithoutPrivateKey)
		}
	}
	body, code, err := c.newRequest(ctx,
		"/accountchangepassword/"+d.Get(skAccountID).(string)+"/"+credentialType, http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("propagating the credential to the target host: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteDomainAccountCredential(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	domainID := d.Get(skDomainID).(string)
	accountID := d.Get(skAccountID).(string)
	credentials := "/domains/" + domainID + "/accounts/" + accountID + "/credentials/"
	body, code, err := c.newRequest(ctx, credentials+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code == http.StatusNotFound {
		// A rotation recreates the credential under a new id, which the read only follows after a
		// refresh: without one (-refresh=false, a saved plan), delete the credential of the same
		// account and type, or the rotated credential would stay on the Bastion out of state.
		rotatedID, found, err := searchResourceDomainAccountCredential(ctx,
			domainID, accountID, d.Get(skType).(string), m)
		if err != nil {
			return err
		}
		if found && rotatedID != d.Id() {
			body, code, err = c.newRequest(ctx, credentials+rotatedID, http.MethodDelete, nil)
			if err != nil {
				return err
			}
		}
	}
	if !deleteDone(code, body) {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareDomainAccountCredentialJSON(d *schema.ResourceData) jsonCredential {
	jsonData := jsonCredential{Type: d.Get(skType).(string)}
	switch jsonData.Type {
	case skPassword:
		jsonData.Password = d.Get(skPassword).(string)
	case "ssh_key":
		jsonData.PrivateKey = d.Get(skPrivateKey).(string)
		jsonData.Passphrase = d.Get(skPassphrase).(string)
	}

	return jsonData
}

const (
	errPropagateWithoutPassword = "propagate_credential_change requires password: without it, the Bastion " +
		"generates a new password and changes it on the target host"
	errPropagateWithoutPrivateKey = "propagate_credential_change requires an explicit private_key, not a " +
		"generate: one: without it, the Bastion generates a new SSH key and changes it on the target host"
)

// validateDomainAccountCredentialPropagate refuses at plan time a propagation without the value to
// propagate, which the Bastion would replace with a generated one on the target host. A value not
// known yet (another resource's result) is checked again at apply time.
func validateDomainAccountCredentialPropagate(
	_ context.Context, diff *schema.ResourceDiff, _ interface{},
) error {
	if !diff.Get(skPropagateCredentialChange).(bool) {
		return nil
	}
	switch diff.Get(skType).(string) {
	case skPassword:
		if diff.NewValueKnown(skPassword) && diff.Get(skPassword).(string) == "" {
			return errors.New(errPropagateWithoutPassword)
		}
	default: // ssh_key, the only other type the schema accepts
		if !diff.NewValueKnown(skPrivateKey) {
			return nil
		}
		if key := diff.Get(skPrivateKey).(string); key == "" || strings.HasPrefix(key, "generate:") {
			return errors.New(errPropagateWithoutPrivateKey)
		}
	}

	return nil
}

func readDomainAccountCredentialOptions(
	ctx context.Context, domainID, accountID, credentialID string, m interface{},
) (
	jsonCredential, error,
) {
	c := m.(*Client)
	var result jsonCredential
	body, code, err := c.newRequest(ctx,
		"/domains/"+domainID+"/accounts/"+accountID+"/credentials/"+credentialID,
		http.MethodGet, nil)
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
	// avoid the bug when the credential still exists but not linked to the account
	credsID, found, err := searchResourceDomainAccountCredential(ctx, domainID, accountID, result.Type, m)
	if err != nil {
		return result, err
	}
	if !found {
		return jsonCredential{}, nil
	}
	if credsID != result.ID {
		return jsonCredential{}, nil
	}

	return result, nil
}

func fillDomainAccountCredential(d *schema.ResourceData, jsonData jsonCredential) {
	if tfErr := d.Set(skType, jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPublicKey, jsonData.PublicKey); tfErr != nil {
		panic(tfErr)
	}
}
