package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDomain_basic(t *testing.T) {
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
				Config: testAccResourceDomainCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_domain.testacc_Domain",
						"id"),
				),
			},
			{
				Config: testAccResourceDomainUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_domain.testacc_Domain",
				ImportState:   true,
				ImportStateId: "testacc_Domain",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceDomainCreate() string {
	return `
resource wallix-bastion_domain testacc_Domain {
  domain_name = "testacc_Domain"
  domain_real_name = "testacc.domain"
  ca_private_key = "generate:RSA_4096"
}
resource wallix-bastion_domain testacc_Domain2 {
  domain_name = "testacc_Domain2"
}
resource wallix-bastion_domain_account testacc_Domain {
  domain_id = wallix-bastion_domain.testacc_Domain.id
  account_name = "testacc_Domain_Admin"
  account_login = "admin"
}
`
}

func testAccResourceDomainUpdate() string {
	return `
resource wallix-bastion_domain testacc_Domain {
  domain_name = "testacc_Domain"
  domain_real_name = "testacc.domain"
  ca_private_key = tls_private_key.testacc_Domain.private_key_pem
  passphrase = random_password.testacc_Domain.result
  admin_account = "testacc_Domain_Admin"
  description = "testacc Domain"
  enable_password_change = true
  password_change_policy = "default"
  password_change_plugin = "Unix"
  password_change_plugin_parameters = jsonencode({
    host: "192.0.2.1"
  })
}
resource wallix-bastion_domain testacc_Domain2 {
  domain_name = "testacc_Domain2"
}
resource "tls_private_key" "testacc_Domain" {
  algorithm = "RSA"
  rsa_bits  = 4096
}
resource "random_password" "testacc_Domain" {
  length           = 12
  special          = true
  override_special = "_%@"
  min_upper        = 1
  min_numeric      = 1
  min_special      = 1
}
resource wallix-bastion_domain_account testacc_Domain {
  domain_id = wallix-bastion_domain.testacc_Domain.id
  account_name = "testacc_Domain_Admin"
  account_login = "admin"
}
`
}
