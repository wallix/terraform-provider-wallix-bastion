package bastion_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// config_x509 is a singleton resource on the Bastion (a single, static
// configuration identified by the hardcoded ID "x509Config"). The data
// source therefore has no Required lookup field: it always reads the single
// existing x509 configuration, so the "data" block below takes no arguments.
func TestAccDataSourceConfigX509_basic(t *testing.T) {
	dataSourceName := "data.wallix-bastion_config_x509.current"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Create the resource to be fetched by the datasource.
				Config: testAccDataSourceConfigX509ConfigCreate(),
			},
			{
				// Validate that the datasource correctly retrieves the resource.
				Config: testAccDataSourceConfigX509ConfigData(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ca_certificate"),
					resource.TestCheckResourceAttrSet(dataSourceName, "server_public_key"),
					resource.TestCheckResourceAttrSet(dataSourceName, "server_private_key"),
					resource.TestCheckResourceAttr(dataSourceName, "enable", "true"),
				),
			},
		},
	})
}

// Resource creation configuration.
// getBastionHostname() is defined in resource_config_x509_test.go (same package).
func testAccDataSourceConfigX509ConfigCreate() string {
	hostname := getBastionHostname()

	return fmt.Sprintf(`
resource "tls_private_key" "ca_ds" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "ca_ds" {
  private_key_pem = tls_private_key.ca_ds.private_key_pem

  subject {
    common_name  = "WALLIX Bastion Test CA DS"
    organization = "WALLIX Test"
    country      = "FR"
  }

  validity_period_hours = 8760 # 1 year

  is_ca_certificate = true

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "cert_signing",
    "crl_signing",
  ]
}

resource "tls_private_key" "server_ds" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_cert_request" "server_ds" {
  private_key_pem = tls_private_key.server_ds.private_key_pem

  subject {
    common_name  = "%s"
    organization = "WALLIX Test"
    country      = "FR"
  }

  dns_names = [
    "%s",
    "localhost",
  ]

  ip_addresses = [
    "127.0.0.1",
    "192.168.1.100",
  ]
}

resource "tls_locally_signed_cert" "server_ds" {
  cert_request_pem   = tls_cert_request.server_ds.cert_request_pem
  ca_private_key_pem = tls_private_key.ca_ds.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca_ds.cert_pem

  validity_period_hours = 720 # 30 days

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}

resource "wallix-bastion_config_x509" "test_ds" {
  ca_certificate     = tls_self_signed_cert.ca_ds.cert_pem
  server_public_key  = tls_locally_signed_cert.server_ds.cert_pem
  server_private_key = tls_private_key.server_ds.private_key_pem
  enable             = true
}
`, hostname, hostname)
}

// Datasource configuration to retrieve the created (singleton) resource.
func testAccDataSourceConfigX509ConfigData() string {
	hostname := getBastionHostname()

	return fmt.Sprintf(`
resource "tls_private_key" "ca_ds" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "ca_ds" {
  private_key_pem = tls_private_key.ca_ds.private_key_pem

  subject {
    common_name  = "WALLIX Bastion Test CA DS"
    organization = "WALLIX Test"
    country      = "FR"
  }

  validity_period_hours = 8760 # 1 year

  is_ca_certificate = true

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "cert_signing",
    "crl_signing",
  ]
}

resource "tls_private_key" "server_ds" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_cert_request" "server_ds" {
  private_key_pem = tls_private_key.server_ds.private_key_pem

  subject {
    common_name  = "%s"
    organization = "WALLIX Test"
    country      = "FR"
  }

  dns_names = [
    "%s",
    "localhost",
  ]

  ip_addresses = [
    "127.0.0.1",
    "192.168.1.100",
  ]
}

resource "tls_locally_signed_cert" "server_ds" {
  cert_request_pem   = tls_cert_request.server_ds.cert_request_pem
  ca_private_key_pem = tls_private_key.ca_ds.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca_ds.cert_pem

  validity_period_hours = 720 # 30 days

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}

resource "wallix-bastion_config_x509" "test_ds" {
  ca_certificate     = tls_self_signed_cert.ca_ds.cert_pem
  server_public_key  = tls_locally_signed_cert.server_ds.cert_pem
  server_private_key = tls_private_key.server_ds.private_key_pem
  enable             = true
}

data "wallix-bastion_config_x509" "current" {}
`, hostname, hostname)
}
