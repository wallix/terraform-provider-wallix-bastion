package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDomainAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceDomainAccountConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceDomainAccountConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.wallix-bastion_domain_account.testacc_dataDomainAccount",
						"id"),
					resource.TestCheckResourceAttr("data.wallix-bastion_domain_account.testacc_dataDomainAccount",
						"account_name", "testacc_DomainAccount_Admin"),
					resource.TestCheckResourceAttr("data.wallix-bastion_domain_account.testacc_dataDomainAccount",
						"account_login", "admin"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceDomainAccountConfigCreate() string {
	return `
resource "wallix-bastion_domain" "testacc_DomainAccount" {
  domain_name = "testacc_DomainAccount"
}
resource "wallix-bastion_domain_account" "testacc_DomainAccount" {
  domain_id     = wallix-bastion_domain.testacc_DomainAccount.id
  account_name  = "testacc_DomainAccount_Admin"
  account_login = "admin"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceDomainAccountConfigData() string {
	return `
resource "wallix-bastion_domain" "testacc_DomainAccount" {
  domain_name = "testacc_DomainAccount"
}
resource "wallix-bastion_domain_account" "testacc_DomainAccount" {
  domain_id     = wallix-bastion_domain.testacc_DomainAccount.id
  account_name  = "testacc_DomainAccount_Admin"
  account_login = "admin"
}

data "wallix-bastion_domain_account" "testacc_dataDomainAccount" {
  domain_id    = wallix-bastion_domain.testacc_DomainAccount.id
  account_name = wallix-bastion_domain_account.testacc_DomainAccount.account_name
}
`
}
