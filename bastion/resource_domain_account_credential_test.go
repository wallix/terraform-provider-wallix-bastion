package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDomainAccountCred_basic(t *testing.T) {
	resourceName := "wallix-bastion_domain_account_credential.testacc_DomainAccountCred"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source: "hashicorp/random",
			},
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDomainAccountCredCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						resourceName,
						"id"),
				),
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
					accountID := rs.Primary.Attributes["account_id"]
					if accountID == "" {
						return "", fmt.Errorf("Attribute %s not found:\n%+v", "account_id", rs.Primary.Attributes)
					}

					return domID + "/" + accountID + "/password", nil
				},
			},
			{
				Config: testAccResourceDomainAccountCredUpdate(),
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceDomainAccountCredCreate() string {
	return `
resource wallix-bastion_domain testacc_DomainAccountCred {
  domain_name = "testacc_DomainAccountCred"
}
resource wallix-bastion_domain_account testacc_DomainAccountCred {
  domain_id     = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_name  = "testacc_DomainAccountCred_Admin"
  account_login = "admin"
}
resource wallix-bastion_domain_account_credential testacc_DomainAccountCred {
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
resource wallix-bastion_domain_account_credential testacc_DomainAccountCred2 {
  domain_id   = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_id  = wallix-bastion_domain_account.testacc_DomainAccountCred.id
  type        = "ssh_key"
  private_key = "generate:RSA_4096"
}

`
}

func testAccResourceDomainAccountCredUpdate() string {
	return `
resource wallix-bastion_domain testacc_DomainAccountCred {
  domain_name = "testacc_DomainAccountCred"
}
resource wallix-bastion_domain_account testacc_DomainAccountCred {
  domain_id     = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_name  = "testacc_DomainAccountCred_Admin"
  account_login = "admin"
}

resource wallix-bastion_domain_account_credential testacc_DomainAccountCred2 {
  domain_id   = wallix-bastion_domain.testacc_DomainAccountCred.id
  account_id  = wallix-bastion_domain_account.testacc_DomainAccountCred.id
  type        = "ssh_key"
  private_key = tls_private_key.testacc_DomainAccountCred.private_key_pem
  passphrase  = random_password.testacc_DomainAccountCred.result
}
resource "random_password" "testacc_DomainAccountCred" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
resource "tls_private_key" "testacc_DomainAccountCred" {
  algorithm = "RSA"
  rsa_bits  = 4096
}
`
}
