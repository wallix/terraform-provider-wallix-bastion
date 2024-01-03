package bastion_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/wallix/terraform-provider-wallix-bastion/bastion"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceAuthDomainMapping_basic(t *testing.T) {
	resourceName := "wallix-bastion_authdomain_mapping.testacc_AuthDomainMapping"
	if v := os.Getenv("WALLIX_BASTION_API_VERSION"); v != "" &&
		v != bastion.VersionWallixAPI33 &&
		v != bastion.VersionWallixAPI36 {
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: testAccResourceAuthDomainMappingCreate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceName,
							"id"),
						resource.TestCheckResourceAttrSet(resourceName,
							"domain"),
					),
				},
				{
					Config: testAccResourceAuthDomainMappingUpdate(),
				},
				{
					ResourceName: resourceName,
					ImportState:  true,
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return "", fmt.Errorf("Resource %s not found", resourceName)
						}
						devID := rs.Primary.Attributes["domain_id"]
						if devID == "" {
							return "", fmt.Errorf("Attribute %s not found:\n%+v", "domain_id", rs.Primary.Attributes)
						}

						return devID + "/testacc_AuthDomainMapping2", nil
					},
				},
			},
			PreventPostDestroyRefresh: true,
		})
	}
}

func testAccResourceAuthDomainMappingCreate() string {
	return `
resource "wallix-bastion_authdomain_ldap" "testacc_AuthDomainMapping" {
  domain_name          = "testacc_AuthDomainMapping"
  auth_domain_name     = "test.com"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_AuthDomainMapping.authentication_name]
  default_language     = "fr"
  default_email_domain = "test.com"
}
resource "wallix-bastion_externalauth_ldap" "testacc_AuthDomainMapping" {
  authentication_name = "testacc_AuthDomainMapping"
  cn_attribute        = "sAMAccountName"
  host                = "server1"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
resource "wallix-bastion_usergroup" "testacc_AuthDomainMapping" {
  group_name = "testacc_AuthDomainMapping"
  timeframes = ["allthetime"]
  profile    = "user"
}
resource "wallix-bastion_authdomain_mapping" "testacc_AuthDomainMapping" {
  domain_id      = wallix-bastion_authdomain_ldap.testacc_AuthDomainMapping.id
  user_group     = wallix-bastion_usergroup.testacc_AuthDomainMapping.group_name
  external_group = "CN=testacc,OU=FR,DC=test,DC=com"
}
`
}

func testAccResourceAuthDomainMappingUpdate() string {
	return `
resource "wallix-bastion_authdomain_ldap" "testacc_AuthDomainMapping" {
  domain_name          = "testacc_AuthDomainMapping"
  auth_domain_name     = "test.com"
  external_auths       = [wallix-bastion_externalauth_ldap.testacc_AuthDomainMapping.authentication_name]
  default_language     = "fr"
  default_email_domain = "test.com"
}
resource "wallix-bastion_externalauth_ldap" "testacc_AuthDomainMapping" {
  authentication_name = "testacc_AuthDomainMapping"
  cn_attribute        = "sAMAccountName"
  host                = "server1"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
resource "wallix-bastion_usergroup" "testacc_AuthDomainMapping" {
  group_name = "testacc_AuthDomainMapping"
  timeframes = ["allthetime"]
  profile    = "user"
}
resource "wallix-bastion_usergroup" "testacc_AuthDomainMapping2" {
  group_name = "testacc_AuthDomainMapping2"
  timeframes = ["allthetime"]
  profile    = "user"
}
resource "wallix-bastion_authdomain_mapping" "testacc_AuthDomainMapping" {
  domain_id      = wallix-bastion_authdomain_ldap.testacc_AuthDomainMapping.id
  user_group     = wallix-bastion_usergroup.testacc_AuthDomainMapping2.group_name
  external_group = "CN=testacc2,OU=FR,DC=test,DC=com"
}
`
}
