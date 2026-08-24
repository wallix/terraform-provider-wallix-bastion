package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceExternalAuthLDAP_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceExternalAuthLDAPConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceExternalAuthLDAPConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"authentication_name", "testacc_dataExternalAuthLDAP"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"cn_attribute", "sAMAccountName"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"host", "192.168.100.20"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"ldap_base", "OU=FR,DC=test,DC=com"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"login_attribute", "sAMAccountName"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"port", "636"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"timeout", "10"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"is_ssl", "true"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"is_active_directory", "true"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"login", "svc1"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"description", "testacc ExternalAuthLDAP"),
					resource.TestCheckResourceAttr(
						"data.wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP",
						"is_protected_user", "true"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceExternalAuthLDAPConfigCreate() string {
	return `
resource "wallix-bastion_externalauth_ldap" "testacc_dataExternalAuthLDAP" {
  authentication_name = "testacc_dataExternalAuthLDAP"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.20"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_active_directory = true
  login               = "svc1"
  password            = "aPassword"
  description         = "testacc ExternalAuthLDAP"
  is_protected_user   = true
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceExternalAuthLDAPConfigData() string {
	return `
resource "wallix-bastion_externalauth_ldap" "testacc_dataExternalAuthLDAP" {
  authentication_name = "testacc_dataExternalAuthLDAP"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.20"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_active_directory = true
  login               = "svc1"
  password            = "aPassword"
  description         = "testacc ExternalAuthLDAP"
  is_protected_user   = true
}

data "wallix-bastion_externalauth_ldap" "testacc_dataExternalAuthLDAP" {
  authentication_name = wallix-bastion_externalauth_ldap.testacc_dataExternalAuthLDAP.authentication_name
}
`
}
