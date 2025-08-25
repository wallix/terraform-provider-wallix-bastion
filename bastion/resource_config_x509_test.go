package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccResourceConfigX509_basic tests creating, updating the x509 configuration.
func TestAccResourceConfigX509_basic(t *testing.T) {
	resourceName := "wallix-bastion_config_x509.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigX509Basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "server_public_key"),
					resource.TestCheckResourceAttrSet(resourceName, "server_private_key"),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
				),
				ExpectNonEmptyPlan: true,
			},
			// Test updating the resource
			{
				Config: testAccResourceConfigX509Update(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "server_public_key"),
					resource.TestCheckResourceAttrSet(resourceName, "server_private_key"),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
				),
				ExpectNonEmptyPlan: true,
			},
			// Test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "x509_config",
			},
		},
		PreventPostDestroyRefresh: true, // Prevent deletion
	})
}

// TestAccResourceConfigX509_enableToggle tests enabling and disabling X509 configuration.
func TestAccResourceConfigX509_enableToggle(t *testing.T) {
	resourceName := "wallix-bastion_config_x509.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		ExternalProviders: map[string]resource.ExternalProvider{
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		Steps: []resource.TestStep{
			// Create with enable=false
			{
				Config: testAccResourceConfigX509Disabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "enable", "false"),
				),
				ExpectNonEmptyPlan: true,
			},
			// Update to enable=true
			{
				Config: testAccResourceConfigX509Enabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
		PreventPostDestroyRefresh: true, // Prevent deletion
	})
}

// Test configuration for creating the resource with TLS-generated certificates.
func testAccResourceConfigX509Basic() string {
	return `
# Generate CA private key
resource "tls_private_key" "ca" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Generate CA certificate
resource "tls_self_signed_cert" "ca" {
  private_key_pem = tls_private_key.ca.private_key_pem

  subject {
    common_name  = "Wallix Bastion Test CA"
    organization = "Wallix Test"
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

# Generate server private key
resource "tls_private_key" "server" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Generate server certificate signed by CA
resource "tls_cert_request" "server" {
  private_key_pem = tls_private_key.server.private_key_pem

  subject {
    common_name  = "bastion.test.local"
    organization = "Wallix Test"
    country      = "FR"
  }

  dns_names = [
    "bastion.test.local",
    "localhost",
  ]

  ip_addresses = [
    "127.0.0.1",
    "192.168.1.100",
  ]
}

resource "tls_locally_signed_cert" "server" {
  cert_request_pem   = tls_cert_request.server.cert_request_pem
  ca_private_key_pem = tls_private_key.ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca.cert_pem

  validity_period_hours = 720 # 30 days

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}

# Wallix Bastion X509 configuration
resource "wallix-bastion_config_x509" "test" {
  ca_certificate     = tls_self_signed_cert.ca.cert_pem
  server_public_key  = tls_locally_signed_cert.server.cert_pem
  server_private_key = tls_private_key.server.private_key_pem
  enable             = true
}
`
}

// Test configuration for updating the resource with new certificates.
func testAccResourceConfigX509Update() string {
	return `
# Generate new CA private key for update test
resource "tls_private_key" "ca_updated" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Generate new CA certificate
resource "tls_self_signed_cert" "ca_updated" {
  private_key_pem = tls_private_key.ca_updated.private_key_pem

  subject {
    common_name  = "Wallix Bastion Updated Test CA"
    organization = "Wallix Test Updated"
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

# Generate new server private key
resource "tls_private_key" "server_updated" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Generate new server certificate signed by updated CA
resource "tls_cert_request" "server_updated" {
  private_key_pem = tls_private_key.server_updated.private_key_pem

  subject {
    common_name  = "bastion-updated.test.local"
    organization = "Wallix Test Updated"
    country      = "FR"
  }

  dns_names = [
    "bastion-updated.test.local",
    "localhost",
  ]

  ip_addresses = [
    "127.0.0.1",
    "192.168.1.101",
  ]
}

resource "tls_locally_signed_cert" "server_updated" {
  cert_request_pem   = tls_cert_request.server_updated.cert_request_pem
  ca_private_key_pem = tls_private_key.ca_updated.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca_updated.cert_pem

  validity_period_hours = 720 # 30 days

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}

# Updated Wallix Bastion X509 configuration
resource "wallix-bastion_config_x509" "test" {
  ca_certificate     = tls_self_signed_cert.ca_updated.cert_pem
  server_public_key  = tls_locally_signed_cert.server_updated.cert_pem
  server_private_key = tls_private_key.server_updated.private_key_pem
  enable             = false
}
`
}

// Test configuration with enable=false.
func testAccResourceConfigX509Disabled() string {
	return `
# Generate CA private key
resource "tls_private_key" "ca" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Generate CA certificate
resource "tls_self_signed_cert" "ca" {
  private_key_pem = tls_private_key.ca.private_key_pem

  subject {
    common_name  = "Wallix Bastion Test CA"
    organization = "Wallix Test"
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

# Generate server private key
resource "tls_private_key" "server" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Generate server certificate signed by CA
resource "tls_cert_request" "server" {
  private_key_pem = tls_private_key.server.private_key_pem

  subject {
    common_name  = "bastion.test.local"
    organization = "Wallix Test"
    country      = "FR"
  }

  dns_names = [
    "bastion.test.local",
    "localhost",
  ]

  ip_addresses = [
    "127.0.0.1",
    "192.168.1.100",
  ]
}

resource "tls_locally_signed_cert" "server" {
  cert_request_pem   = tls_cert_request.server.cert_request_pem
  ca_private_key_pem = tls_private_key.ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca.cert_pem

  validity_period_hours = 720 # 30 days

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}

# Wallix Bastion X509 configuration (disabled)
resource "wallix-bastion_config_x509" "test" {
  ca_certificate     = tls_self_signed_cert.ca.cert_pem
  server_public_key  = tls_locally_signed_cert.server.cert_pem
  server_private_key = tls_private_key.server.private_key_pem
  enable             = false
}
`
}

// Test configuration with enable=true.
func testAccResourceConfigX509Enabled() string {
	return `
# Generate CA private key
resource "tls_private_key" "ca" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Generate CA certificate
resource "tls_self_signed_cert" "ca" {
  private_key_pem = tls_private_key.ca.private_key_pem

  subject {
    common_name  = "Wallix Bastion Test CA"
    organization = "Wallix Test"
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

# Generate server private key
resource "tls_private_key" "server" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Generate server certificate signed by CA
resource "tls_cert_request" "server" {
  private_key_pem = tls_private_key.server.private_key_pem

  subject {
    common_name  = "bastion.test.local"
    organization = "Wallix Test"
    country      = "FR"
  }

  dns_names = [
    "bastion.test.local",
    "localhost",
  ]

  ip_addresses = [
    "127.0.0.1",
    "192.168.1.100",
  ]
}

resource "tls_locally_signed_cert" "server" {
  cert_request_pem   = tls_cert_request.server.cert_request_pem
  ca_private_key_pem = tls_private_key.ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca.cert_pem

  validity_period_hours = 720 # 30 days

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}

# Wallix Bastion X509 configuration (enabled)
resource "wallix-bastion_config_x509" "test" {
  ca_certificate     = tls_self_signed_cert.ca.cert_pem
  server_public_key  = tls_locally_signed_cert.server.cert_pem
  server_private_key = tls_private_key.server.private_key_pem
  enable             = true
}
`
}
