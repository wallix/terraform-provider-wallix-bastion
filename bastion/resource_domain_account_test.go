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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
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

// Emptying resources was never sent: the account kept its access to the resources.
func TestAccResourceDomainAccount_clearResources(t *testing.T) {
	resourceName := "wallix-bastion_domain_account.testacc_DomAccClearRes"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDomainAccountClearResources(`["testacc_DomAccClearRes:SSH"]`),
				Check:  resource.TestCheckResourceAttr(resourceName, "resources.#", "1"),
			},
			{
				Config: testAccResourceDomainAccountClearResources(`[]`),
				Check:  resource.TestCheckResourceAttr(resourceName, "resources.#", "0"),
			},
		},
	})
}

func testAccResourceDomainAccountClearResources(resources string) string {
	return fmt.Sprintf(`
resource "wallix-bastion_domain" "testacc_DomAccClearRes" {
  domain_name = "testacc_DomAccClearRes"
}

resource "wallix-bastion_device" "testacc_DomAccClearRes" {
  device_name = "testacc_DomAccClearRes"
  host        = "192.168.100.14"
}

resource "wallix-bastion_device_service" "testacc_DomAccClearRes" {
  device_id         = wallix-bastion_device.testacc_DomAccClearRes.id
  service_name      = "SSH"
  connection_policy = "SSH"
  port              = 22
  protocol          = "SSH"
  subprotocols      = ["SSH_SHELL_SESSION"]
  global_domains    = [wallix-bastion_domain.testacc_DomAccClearRes.domain_name]
}

resource "wallix-bastion_domain_account" "testacc_DomAccClearRes" {
  domain_id     = wallix-bastion_domain.testacc_DomAccClearRes.id
  account_name  = "testacc_DomAccClearRes"
  account_login = "testacc_DomAccClearRes"
  resources     = %s
  depends_on    = [wallix-bastion_device_service.testacc_DomAccClearRes]
}
`, resources)
}
