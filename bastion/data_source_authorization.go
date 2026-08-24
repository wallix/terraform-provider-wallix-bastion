package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAuthorization() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthorizationRead,
		Schema: map[string]*schema.Schema{
			"authorization_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorize_password_retrieval": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"authorize_sessions": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"authorize_session_sharing": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"session_sharing_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skSubprotocols: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"is_critical": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_recorded": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skApprovalRequired: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"approvers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"active_quorum": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"inactive_quorum": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"approval_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"has_comment": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"has_ticket": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mandatory_comment": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mandatory_ticket": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"single_connection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAuthorizationVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_authorization not available with api version %s", version)
}

func dataSourceAuthorizationRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceAuthorizationVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthorization(ctx, d.Get("authorization_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authorization_name %s doesn't exists", d.Get("authorization_name").(string)))
	}
	cfg, err := readAuthorizationOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceAuthorization(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceAuthorization(d *schema.ResourceData, jsonData jsonAuthorization) {
	if tfErr := d.Set("authorization_name", jsonData.AuthorizationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("user_group", jsonData.UserGroup); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("target_group", jsonData.TargetGroup); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("authorize_password_retrieval", jsonData.AuthorizePasswordRetrieval); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("authorize_sessions", jsonData.AuthorizeSessions); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("authorize_session_sharing", jsonData.AuthorizeSessionSharing); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("session_sharing_mode", jsonData.SessionSharingMode); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skSubprotocols, jsonData.SubProtocols); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_critical", jsonData.IsCritical); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_recorded", jsonData.IsRecorded); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skApprovalRequired, jsonData.ApprovalRequired); tfErr != nil {
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
