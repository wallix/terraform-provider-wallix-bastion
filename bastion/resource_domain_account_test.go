package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDomainAccount_basic(t *testing.T) {
	resourceName := "wallix-bastion_domain_account.testacc_DomainAccount"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDomainAccountCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						resourceName,
						"id"),
				),
			},
			{
				Config: testAccResourceDomainAccountUpdate(),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Resource %s not found", resourceName)
					}
					domID := rs.Primary.Attributes["domain_id"]
					if domID == "" {
						return "", fmt.Errorf("Attribute %s not found:\n%+v", "domain_id", rs.Primary.Attributes)
					}

					return domID + "/testacc_DomainAccount_Admin", nil
				},
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceDomainAccountCreate() string {
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

func testAccResourceDomainAccountUpdate() string {
	return `
resource "wallix-bastion_domain" "testacc_DomainAccount" {
  domain_name = "testacc_DomainAccount"
}
resource "wallix-bastion_domain_account" "testacc_DomainAccount" {
  domain_id            = wallix-bastion_domain.testacc_DomainAccount.id
  account_name         = "testacc_DomainAccount_Admin"
  account_login        = "admin"
  auto_change_password = true
  auto_change_ssh_key  = true
  certificate_validity = "+2h"
  description          = "testacc domain"
}
`
}
