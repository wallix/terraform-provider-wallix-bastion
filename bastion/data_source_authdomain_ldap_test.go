package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAuthDomainLdap_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceAuthDomainLdapConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceAuthDomainLdapConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_ldap.testacc_dataAuthDomainLdap",
						"domain_name", "testacc-domain-ldap-ds"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_ldap.testacc_dataAuthDomainLdap",
						"auth_domain_name", "test3-ds.com"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_ldap.testacc_dataAuthDomainLdap",
						"default_language", "fr"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_ldap.testacc_dataAuthDomainLdap",
						"default_email_domain", "test3-ds.com"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceAuthDomainLdapConfigCreate() string {
	return `
resource "wallix-bastion_authdomain_ldap" "testacc_dataAuthDomainLdap_ds" {
  domain_name          = "testacc-domain-ldap-ds"
  auth_domain_name     = "test3-ds.com"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_dataAuthDomainLdap_ds.authentication_name]
  default_language     = "fr"
  default_email_domain = "test3-ds.com"
}
resource "wallix-bastion_externalauth_ldap" "testacc_dataAuthDomainLdap_ds" {
  authentication_name = "testacc_dataAuthDomainLdap_ds"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.21"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceAuthDomainLdapConfigData() string {
	return `
resource "wallix-bastion_authdomain_ldap" "testacc_dataAuthDomainLdap_ds" {
  domain_name          = "testacc-domain-ldap-ds"
  auth_domain_name     = "test3-ds.com"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_dataAuthDomainLdap_ds.authentication_name]
  default_language     = "fr"
  default_email_domain = "test3-ds.com"
}
resource "wallix-bastion_externalauth_ldap" "testacc_dataAuthDomainLdap_ds" {
  authentication_name = "testacc_dataAuthDomainLdap_ds"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.21"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}

data "wallix-bastion_authdomain_ldap" "testacc_dataAuthDomainLdap" {
  domain_name = wallix-bastion_authdomain_ldap.testacc_dataAuthDomainLdap_ds.domain_name
}
`
}
