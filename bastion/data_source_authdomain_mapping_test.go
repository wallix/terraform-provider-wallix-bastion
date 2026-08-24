package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAuthDomainMapping_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceAuthDomainMappingConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceAuthDomainMappingConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_mapping.testacc_dataAuthDomainMapping",
						"user_group", "testacc_dataAuthDomainMapping_ds"),
					resource.TestCheckResourceAttr("data.wallix-bastion_authdomain_mapping.testacc_dataAuthDomainMapping",
						"external_group", "CN=testaccds,OU=FR,DC=test,DC=com"),
					resource.TestCheckResourceAttrSet("data.wallix-bastion_authdomain_mapping.testacc_dataAuthDomainMapping",
						"domain"),
				),
			},
		},
	})
}

// Resource creation configuration.
func testAccDataSourceAuthDomainMappingConfigCreate() string {
	return `
resource "wallix-bastion_authdomain_ldap" "testacc_dataAuthDomainMapping_ds" {
  domain_name          = "testacc-domain-mapping-ds"
  auth_domain_name     = "testacc-auth-domain-mapping-ds"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_dataAuthDomainMapping_ds.authentication_name]
  default_language     = "fr"
  default_email_domain = "testacc-auth-domain-mapping-ds"
}
resource "wallix-bastion_externalauth_ldap" "testacc_dataAuthDomainMapping_ds" {
  authentication_name = "testacc_dataAuthDomainMapping_ds"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.20"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
resource "wallix-bastion_usergroup" "testacc_dataAuthDomainMapping_ds" {
  group_name = "testacc_dataAuthDomainMapping_ds"
  timeframes = ["allthetime"]
  profile    = "user"
}
resource "wallix-bastion_authdomain_mapping" "testacc_dataAuthDomainMapping_ds" {
  domain_id      = wallix-bastion_authdomain_ldap.testacc_dataAuthDomainMapping_ds.id
  user_group     = wallix-bastion_usergroup.testacc_dataAuthDomainMapping_ds.group_name
  external_group = "CN=testaccds,OU=FR,DC=test,DC=com"
}
`
}

// Datasource configuration to retrieve the created resource.
func testAccDataSourceAuthDomainMappingConfigData() string {
	return `
resource "wallix-bastion_authdomain_ldap" "testacc_dataAuthDomainMapping_ds" {
  domain_name          = "testacc-domain-mapping-ds"
  auth_domain_name     = "testacc-auth-domain-mapping-ds"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_dataAuthDomainMapping_ds.authentication_name]
  default_language     = "fr"
  default_email_domain = "testacc-auth-domain-mapping-ds"
}
resource "wallix-bastion_externalauth_ldap" "testacc_dataAuthDomainMapping_ds" {
  authentication_name = "testacc_dataAuthDomainMapping_ds"
  cn_attribute        = "sAMAccountName"
  host                = "192.168.100.20"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
resource "wallix-bastion_usergroup" "testacc_dataAuthDomainMapping_ds" {
  group_name = "testacc_dataAuthDomainMapping_ds"
  timeframes = ["allthetime"]
  profile    = "user"
}
resource "wallix-bastion_authdomain_mapping" "testacc_dataAuthDomainMapping_ds" {
  domain_id      = wallix-bastion_authdomain_ldap.testacc_dataAuthDomainMapping_ds.id
  user_group     = wallix-bastion_usergroup.testacc_dataAuthDomainMapping_ds.group_name
  external_group = "CN=testaccds,OU=FR,DC=test,DC=com"
}

data "wallix-bastion_authdomain_mapping" "testacc_dataAuthDomainMapping" {
  domain_id  = wallix-bastion_authdomain_ldap.testacc_dataAuthDomainMapping_ds.id
  user_group = wallix-bastion_authdomain_mapping.testacc_dataAuthDomainMapping_ds.user_group
}
`
}
