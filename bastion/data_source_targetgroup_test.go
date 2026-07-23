package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTargetGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceTargetGroupConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceTargetGroupConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_targetgroup.testacc_dataTargetgroup",
						"group_name", "testacc_dataTargetgroup"),
					resource.TestCheckResourceAttr("data.wallix-bastion_targetgroup.testacc_dataTargetgroup",
						"description", "testacc data Targetgroup"),
					resource.TestCheckResourceAttr("data.wallix-bastion_targetgroup.testacc_dataTargetgroup",
						"restrictions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.wallix-bastion_targetgroup.testacc_dataTargetgroup",
						"restrictions.*", map[string]string{
							"action":      "notify",
							"rules":       "command",
							"subprotocol": "SSH_REMOTE_COMMAND",
						}),
					resource.TestCheckResourceAttr("data.wallix-bastion_targetgroup.testacc_dataTargetgroup",
						"password_retrieval_accounts.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.wallix-bastion_targetgroup.testacc_dataTargetgroup",
						"password_retrieval_accounts.*", map[string]string{
							"account":     "testacc_dataTargetgroup_Admin",
							"domain":      "testacc_dataTargetgroup",
							"domain_type": "global",
						}),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceTargetGroupConfigCreate() string {
	return `
resource "wallix-bastion_domain" "testacc_dataTargetgroup" {
  domain_name = "testacc_dataTargetgroup"
}
resource "wallix-bastion_domain_account" "testacc_dataTargetgroup" {
  domain_id     = wallix-bastion_domain.testacc_dataTargetgroup.id
  account_name  = "testacc_dataTargetgroup_Admin"
  account_login = "admin"
}
resource "wallix-bastion_targetgroup" "testacc_dataTargetgroup" {
  group_name  = "testacc_dataTargetgroup"
  description = "testacc data Targetgroup"
  restrictions {
    action      = "notify"
    rules       = "command"
    subprotocol = "SSH_REMOTE_COMMAND"
  }
  password_retrieval_accounts {
    account     = wallix-bastion_domain_account.testacc_dataTargetgroup.account_name
    domain      = wallix-bastion_domain.testacc_dataTargetgroup.domain_name
    domain_type = "global"
  }
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceTargetGroupConfigData() string {
	return `
resource "wallix-bastion_domain" "testacc_dataTargetgroup" {
  domain_name = "testacc_dataTargetgroup"
}
resource "wallix-bastion_domain_account" "testacc_dataTargetgroup" {
  domain_id     = wallix-bastion_domain.testacc_dataTargetgroup.id
  account_name  = "testacc_dataTargetgroup_Admin"
  account_login = "admin"
}
resource "wallix-bastion_targetgroup" "testacc_dataTargetgroup" {
  group_name  = "testacc_dataTargetgroup"
  description = "testacc data Targetgroup"
  restrictions {
    action      = "notify"
    rules       = "command"
    subprotocol = "SSH_REMOTE_COMMAND"
  }
  password_retrieval_accounts {
    account     = wallix-bastion_domain_account.testacc_dataTargetgroup.account_name
    domain      = wallix-bastion_domain.testacc_dataTargetgroup.domain_name
    domain_type = "global"
  }
}

data "wallix-bastion_targetgroup" "testacc_dataTargetgroup" {
  group_name = wallix-bastion_targetgroup.testacc_dataTargetgroup.group_name
}
`
}
