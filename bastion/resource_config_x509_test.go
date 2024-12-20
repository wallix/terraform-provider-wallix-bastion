package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccResourceConfigX509_basic tests creating, updating the x509 configuration.
func TestAccResourceConfigX509_basic(t *testing.T) {
	resourceName := "bastion_x509_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) }, // Ensures necessary environment variables are set
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigX509Basic(),
				Check: resource.ComposeTestCheckFunc(
					// Verify that the resource exists
					resource.TestCheckResourceAttr(resourceName, "ca_certificate", "test-ca-cert"),
					resource.TestCheckResourceAttr(resourceName, "server_public_key", "test-public-key"),
					resource.TestCheckResourceAttr(resourceName, "server_private_key", "test-private-key"),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
				),
			},
			// Test updating the resource
			{
				Config: testAccResourceConfigX509Update(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ca_certificate", "updated-ca-cert"),
					resource.TestCheckResourceAttr(resourceName, "server_public_key", "updated-public-key"),
					resource.TestCheckResourceAttr(resourceName, "server_private_key", "updated-private-key"),
					resource.TestCheckResourceAttr(resourceName, "enable", "false"),
				),
			},
		},
	})
}

// Test configuration for creating the resource
func testAccResourceConfigX509Basic() string {
	return (`
resource "bastion_x509_config" "test" {
  ca_certificate    = "test-ca-cert"
  server_public_key = "test-public-key"
  server_private_key = "test-private-key"
  enable            = true
}
`)
}

// Test configuration for updating the resource
func testAccResourceConfigX509Update() string {
	return (`
resource "bastion_x509_config" "test" {
  ca_certificate    = "updated-ca-cert"
  server_public_key = "updated-public-key"
  server_private_key = "updated-private-key"
  enable            = false
}
`)
}
