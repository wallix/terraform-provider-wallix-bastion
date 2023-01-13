package bastion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
)

type jsonAuthorization struct {
	ApprovalRequired           bool      `json:"approval_required"`
	AuthorizePasswordRetrieval bool      `json:"authorize_password_retrieval"`
	AuthorizeSessions          bool      `json:"authorize_sessions"`
	IsCritical                 bool      `json:"is_critical"`
	IsRecorded                 bool      `json:"is_recorded"`
	ID                         string    `json:"id,omitempty"`
	AuthorizationName          string    `json:"authorization_name"`
	Description                string    `json:"description"`
	TargetGroup                string    `json:"target_group,omitempty"`
	UserGroup                  string    `json:"user_group,omitempty"`
	HasComment                 *bool     `json:"has_comment,omitempty"`
	HasTicket                  *bool     `json:"has_ticket,omitempty"`
	MandatoryComment           *bool     `json:"mandatory_comment,omitempty"`
	MandatoryTicket            *bool     `json:"mandatory_ticket,omitempty"`
	SingleConnection           *bool     `json:"single_connection,omitempty"`
	ActiveQuorum               *int      `json:"active_quorum,omitempty"`
	InactiveQuorum             *int      `json:"inactive_quorum,omitempty"`
	ApprovalTimeout            *int      `json:"approval_timeout,omitempty"`
	Approvers                  *[]string `json:"approvers,omitempty"`
	SubProtocols               *[]string `json:"subprotocols,omitempty"`
}

func resourceAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAuthorizationCreate,
		ReadContext:   resourceAuthorizationRead,
		UpdateContext: resourceAuthorizationUpdate,
		DeleteContext: resourceAuthorizationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAuthorizationImport,
		},
		Schema: map[string]*schema.Schema{
			"authorization_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target_group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"authorize_password_retrieval": {
				Type:         schema.TypeBool,
				Optional:     true,
				AtLeastOneOf: []string{"authorize_sessions", "authorize_password_retrieval"},
			},
			"authorize_sessions": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"subprotocols"},
				AtLeastOneOf: []string{"authorize_sessions", "authorize_password_retrieval"},
			},
			"subprotocols": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"is_critical": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"is_recorded": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"approval_required": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"approvers"},
			},
			"approvers": {
				Type:         schema.TypeList,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				RequiredWith: []string{"approval_required"},
			},
			"active_quorum": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1,
				RequiredWith: []string{"approval_required"},
			},
			"inactive_quorum": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1,
				RequiredWith: []string{"approval_required"},
			},
			"approval_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"approval_required"},
			},
			"has_comment": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"approval_required"},
			},
			"has_ticket": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"approval_required"},
			},
			"mandatory_comment": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"approval_required"},
			},
			"mandatory_ticket": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"approval_required"},
			},
			"single_connection": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"approval_required"},
			},
		},
	}
}

func resourceAuthorizationVersionCheck(version string) error {
	if bchk.InSlice(version, defaultVersionsValid()) {
		return nil
	}

	return fmt.Errorf("resource wallix-bastion_authorization not available with api version %s", version)
}

func resourceAuthorizationCreate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthorizationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	_, ex, err := searchResourceAuthorization(ctx, d.Get("authorization_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if ex {
		return diag.FromErr(fmt.Errorf("authorization_name %s already exists", d.Get("authorization_name").(string)))
	}
	err = addAuthorization(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthorization(ctx, d.Get("authorization_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authorization_name %s not found after POST", d.Get("authorization_name").(string)))
	}
	d.SetId(id)

	return resourceAuthorizationRead(ctx, d, m)
}

func resourceAuthorizationRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthorizationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readAuthorizationOptions(ctx, d.Id(), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if cfg.ID == "" {
		d.SetId("")
	} else {
		fillAuthorization(d, cfg)
	}

	return nil
}

func resourceAuthorizationUpdate(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	c := m.(*Client)
	if err := resourceAuthorizationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := updateAuthorization(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceAuthorizationRead(ctx, d, m)
}

func resourceAuthorizationDelete(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceAuthorizationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	if err := deleteAuthorization(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAuthorizationImport(
	d *schema.ResourceData, m interface{},
) (
	[]*schema.ResourceData, error,
) {
	ctx := context.Background()
	c := m.(*Client)
	if err := resourceAuthorizationVersionCheck(c.bastionAPIVersion); err != nil {
		return nil, err
	}
	id, ex, err := searchResourceAuthorization(ctx, d.Id(), m)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("don't find authorization_name with id %s (id must be <authorization_name>", d.Id())
	}
	cfg, err := readAuthorizationOptions(ctx, id, m)
	if err != nil {
		return nil, err
	}
	fillAuthorization(d, cfg)
	result := make([]*schema.ResourceData, 1)
	d.SetId(id)
	result[0] = d

	return result, nil
}

func searchResourceAuthorization(
	ctx context.Context, authorizationName string, m interface{},
) (
	string, bool, error,
) {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/authorizations/?q=authorization_name="+authorizationName, http.MethodGet, nil)
	if err != nil {
		return "", false, err
	}
	if code != http.StatusOK {
		return "", false, fmt.Errorf("api doesn't return OK: %d with body:\n%s", code, body)
	}
	var results []jsonAuthorization
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return "", false, fmt.Errorf("unmarshaling json: %w", err)
	}
	if len(results) == 1 {
		return results[0].ID, true, nil
	}

	return "", false, nil
}

func addAuthorization(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAuthorizationJSON(d, true)
	body, code, err := c.newRequest(ctx, "/authorizations/", http.MethodPost, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func updateAuthorization(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	jsonData := prepareAuthorizationJSON(d, false)
	body, code, err := c.newRequest(ctx, "/authorizations/"+d.Id()+"?force=true", http.MethodPut, jsonData)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func deleteAuthorization(
	ctx context.Context, d *schema.ResourceData, m interface{},
) error {
	c := m.(*Client)
	body, code, err := c.newRequest(ctx, "/authorizations/"+d.Id(), http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if code != http.StatusOK && code != http.StatusNoContent {
		return fmt.Errorf("api doesn't return OK or NoContent: %d with body:\n%s", code, body)
	}

	return nil
}

func prepareAuthorizationJSON(d *schema.ResourceData, newResource bool) jsonAuthorization {
	jsonData := jsonAuthorization{
		AuthorizationName:          d.Get("authorization_name").(string),
		AuthorizePasswordRetrieval: d.Get("authorize_password_retrieval").(bool),
		AuthorizeSessions:          d.Get("authorize_sessions").(bool),
		Description:                d.Get("description").(string),
		ApprovalRequired:           d.Get("approval_required").(bool),
		IsCritical:                 d.Get("is_critical").(bool),
		IsRecorded:                 d.Get("is_recorded").(bool),
	}
	if newResource {
		jsonData.UserGroup = d.Get("user_group").(string)
		jsonData.TargetGroup = d.Get("target_group").(string)
	}
	if d.Get("approval_required").(bool) {
		activeQuorum := d.Get("active_quorum").(int)
		jsonData.ActiveQuorum = &activeQuorum
		inactiveQuorum := d.Get("inactive_quorum").(int)
		jsonData.InactiveQuorum = &inactiveQuorum
		approvalTimeout := d.Get("approval_timeout").(int)
		jsonData.ApprovalTimeout = &approvalTimeout
		approvers := make([]string, 0)
		for _, v := range d.Get("approvers").([]interface{}) {
			approvers = append(approvers, v.(string))
		}
		jsonData.Approvers = &approvers
		hasComment := d.Get("has_comment").(bool)
		jsonData.HasComment = &hasComment
		hasTicket := d.Get("has_ticket").(bool)
		jsonData.HasTicket = &hasTicket
		mandatoryComment := d.Get("mandatory_comment").(bool)
		jsonData.MandatoryComment = &mandatoryComment
		mandatoryTicket := d.Get("mandatory_ticket").(bool)
		jsonData.MandatoryTicket = &mandatoryTicket
		singleConnection := d.Get("single_connection").(bool)
		jsonData.SingleConnection = &singleConnection
	}
	if v := d.Get("subprotocols").(*schema.Set).List(); len(v) > 0 {
		subProtocols := make([]string, 0)
		for _, v2 := range v {
			subProtocols = append(subProtocols, v2.(string))
		}
		jsonData.SubProtocols = &subProtocols
	}

	return jsonData
}

func readAuthorizationOptions(
	ctx context.Context, authorizationID string, m interface{},
) (
	jsonAuthorization, error,
) {
	c := m.(*Client)
	var result jsonAuthorization
	body, code, err := c.newRequest(ctx, "/authorizations/"+authorizationID, http.MethodGet, nil)
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

func fillAuthorization(d *schema.ResourceData, jsonData jsonAuthorization) {
	if tfErr := d.Set("authorization_name", jsonData.AuthorizationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("user_group", jsonData.UserGroup); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("target_group", jsonData.TargetGroup); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("authorize_password_retrieval", jsonData.AuthorizePasswordRetrieval); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("authorize_sessions", jsonData.AuthorizeSessions); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("subprotocols", jsonData.SubProtocols); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_critical", jsonData.IsCritical); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_recorded", jsonData.IsRecorded); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("approval_required", jsonData.ApprovalRequired); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("approvers", jsonData.Approvers); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("active_quorum", jsonData.ActiveQuorum); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("inactive_quorum", jsonData.InactiveQuorum); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("approval_timeout", jsonData.ApprovalTimeout); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("has_comment", jsonData.HasComment); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("has_ticket", jsonData.HasTicket); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("mandatory_comment", jsonData.MandatoryComment); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("mandatory_ticket", jsonData.MandatoryTicket); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("single_connection", jsonData.SingleConnection); tfErr != nil {
		panic(tfErr)
	}
}
