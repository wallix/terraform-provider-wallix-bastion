package bastion_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceConnectionPolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConnectionPolicyCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"wallix-bastion_connection_policy.testacc_ConnectionPolicy",
						"id"),
				),
			},
			{
				Config: testAccResourceConnectionPolicyUpdate(),
			},
			{
				ResourceName:  "wallix-bastion_connection_policy.testacc_ConnectionPolicy2",
				ImportState:   true,
				ImportStateId: "testacc_ConnectionPolicy2",
			},
		},
		PreventPostDestroyRefresh: true,
	})
}

func testAccResourceConnectionPolicyCreate() string {
	return `
resource wallix-bastion_connection_policy testacc_ConnectionPolicy {
  connection_policy_name = "testacc_ConnectionPolicy"
  protocol               = "RAWTCPIP"
  options = jsonencode({
    general = {}
  })
}
resource wallix-bastion_connection_policy testacc_ConnectionPolicy2 {
  connection_policy_name = "testacc_ConnectionPolicy2"
  description            = "testacc ConnectionPolicy2"
  protocol               = "SSH"
  authentication_methods = ["PASSWORD_VAULT"]
  options = jsonencode(
    {
      algorithms = {
        cipher_algos      = ""
        compression_algos = ""
        hostkey_algos     = ""
        integrity_algos   = ""
        kex_algos         = ""
      }
      file_storage = {
        store_file = "never"
      }
      file_verification = {
        enable_down = false
        enable_up   = false
      }
      general = {
        transformation_rule       = ""
        vault_transformation_rule = ""
      }
      proxy = {
        enable     = false
        host       = ""
        login      = ""
        password   = ""
        port       = 0
        proxy_type = "None"
      }
      restriction = {
        cmds_compatibility = "cisco"
      }
      server_pubkey = {
        server_access_allowed_message = "0"
        server_pubkey_check           = "1"
        server_pubkey_create_message  = "1"
        server_pubkey_failure_message = "1"
        server_pubkey_store           = true
        server_pubkey_success_message = "0"
      }
      session = {
        allow_multi_channels      = false
        force_shell_disconnection = false
        inactivity_timeout        = 0
        server_keepalive_interval = 0
        server_keepalive_type     = "none"
      }
      startup_scenario = {
        ask_startup = false
        enable      = false
        scenario    = ""
        show_output = true
        timeout     = 10
      }
      trace = {
        log_all_kbd          = false
        log_group_membership = false
      }
    }
  )
}
`
}

func testAccResourceConnectionPolicyUpdate() string {
	return `
resource wallix-bastion_connection_policy testacc_ConnectionPolicy2 {
  connection_policy_name = "testacc_ConnectionPolicy2"
  description            = "testacc ConnectionPolicy2"
  protocol               = "SSH"
  authentication_methods = ["PASSWORD_VAULT", "PASSWORD_MAPPING"]
  options = jsonencode(
    {
      algorithms = {
        cipher_algos      = ""
        compression_algos = ""
        hostkey_algos     = ""
        integrity_algos   = ""
        kex_algos         = ""
      }
      file_storage = {
        store_file = "never"
      }
      file_verification = {
        enable_down = true
        enable_up   = true
      }
      general = {
        transformation_rule       = ""
        vault_transformation_rule = ""
      }
      proxy = {
        enable     = false
        host       = ""
        login      = ""
        password   = ""
        port       = 0
        proxy_type = "None"
      }
      restriction = {
        cmds_compatibility = "cisco"
      }
      server_pubkey = {
        server_access_allowed_message = "0"
        server_pubkey_check           = "1"
        server_pubkey_create_message  = "1"
        server_pubkey_failure_message = "1"
        server_pubkey_store           = true
        server_pubkey_success_message = "0"
      }
      session = {
        allow_multi_channels      = true
        force_shell_disconnection = false
        inactivity_timeout        = 0
        server_keepalive_interval = 0
        server_keepalive_type     = "none"
      }
      startup_scenario = {
        ask_startup = false
        enable      = false
        scenario    = ""
        show_output = true
        timeout     = 10
      }
      trace = {
        log_all_kbd          = false
        log_group_membership = false
      }
    }
  )
}
`
}
