package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDomainAccountCred_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			tvRandomProviderName: {
				Source: tvRandomProviderSource,
			},
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceDomainAccountCredConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceDomainAccountCredConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.wallix-bastion_domain_account_credential.testacc_dataDomainAccountCred",
						"id"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_domain_account_credential.testacc_dataDomainAccountCred",
						"type", "ssh_key"),
					resource.TestCheckResourceAttrSet(
						"data.wallix-bastion_domain_account_credential.testacc_dataDomainAccountCred",
						"public_key"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceDomainAccountCredConfigCreate() string {
	return `
resource "wallix-bastion_domain" "testacc_DomainAccountCred" {
  domain_name = "testacc_DomainAccountCred"
}
resource "wallix-bastion_domain_account" "testacc_DomainAccountCred" {
  domain_id     = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_name  = "testacc_DomainAccountCred_Admin"
  account_login = "admin"
}
resource "wallix-bastion_domain_account_credential" "testacc_DomainAccountCred" {
  domain_id  = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_id = wallix-bastion_domain_account.testacc_DomainAccountCred.id
  type       = "password"
  password   = random_password.testacc_DomainAccountCred.result
}
resource "random_password" "testacc_DomainAccountCred" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
resource "wallix-bastion_domain_account_credential" "testacc_DomainAccountCred2" {
  domain_id   = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_id  = wallix-bastion_domain_account.testacc_DomainAccountCred.id
  type        = "ssh_key"
  private_key = "generate:RSA_4096"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceDomainAccountCredConfigData() string {
	return `
resource "wallix-bastion_domain" "testacc_DomainAccountCred" {
  domain_name = "testacc_DomainAccountCred"
}
resource "wallix-bastion_domain_account" "testacc_DomainAccountCred" {
  domain_id     = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_name  = "testacc_DomainAccountCred_Admin"
  account_login = "admin"
}
resource "wallix-bastion_domain_account_credential" "testacc_DomainAccountCred" {
  domain_id  = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_id = wallix-bastion_domain_account.testacc_DomainAccountCred.id
  type       = "password"
  password   = random_password.testacc_DomainAccountCred.result
}
resource "random_password" "testacc_DomainAccountCred" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
resource "wallix-bastion_domain_account_credential" "testacc_DomainAccountCred2" {
  domain_id   = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_id  = wallix-bastion_domain_account.testacc_DomainAccountCred.id
  type        = "ssh_key"
  private_key = "generate:RSA_4096"
}

data "wallix-bastion_domain_account_credential" "testacc_dataDomainAccountCred" {
  domain_id  = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_id = wallix-bastion_domain_account.testacc_DomainAccountCred.id
  type       = wallix-bastion_domain_account_credential.testacc_DomainAccountCred2.type
}
`
}
