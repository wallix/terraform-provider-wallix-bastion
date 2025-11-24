package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAuthDomainAD_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceAuthDomainADConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceAuthDomainADConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_ad.testacc_dataDomain",
						"domain_name", "testacc-domain-ds"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_ad.testacc_dataDomain",
						"auth_domain_name", "testacc-auth-domain-ds"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_ad.testacc_dataDomain",
						"default_language", "en"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_ad.testacc_dataDomain",
						"default_email_domain", "example.com"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceAuthDomainADConfigCreate() string {
	return `
resource "wallix-bastion_externalauth_ldap" "testacc_dataAuthDomain_ds" {
  authentication_name = "testacc_dataAuthDomain_ds"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.20"
  ldap_base           = "OU=Test,DC=example,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}

resource "wallix-bastion_authdomain_ad" "testacc_dataAuthDomain_ds" {
  domain_name          = "testacc-domain-ds"
  auth_domain_name     = "testacc-auth-domain-ds"
  default_language     = "en"
  default_email_domain = "example.com"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_dataAuthDomain_ds.authentication_name]
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceAuthDomainADConfigData() string {
	return `
resource "wallix-bastion_externalauth_ldap" "testacc_dataAuthDomain_ds" {
  authentication_name = "testacc_dataAuthDomain_ds"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.20"
  ldap_base           = "OU=Test,DC=example,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}

resource "wallix-bastion_authdomain_ad" "testacc_dataAuthDomain_ds" {
  domain_name          = "testacc-domain-ds"
  auth_domain_name     = "testacc-auth-domain-ds"
  default_language     = "en"
  default_email_domain = "example.com"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_dataAuthDomain_ds.authentication_name]
}

data "wallix-bastion_authdomain_ad" "testacc_dataDomain" {
  domain_name      = wallix-bastion_authdomain_ad.testacc_dataAuthDomain_ds.domain_name
  auth_domain_name = wallix-bastion_authdomain_ad.testacc_dataAuthDomain_ds.auth_domain_name
}
`
}
