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

// nolint: lll, nolintlint
func testAccResourceConnectionPolicyCreate() string {
	return `
locals {
  optionsv8 = {
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
  optionsv9 = {
    algorithms = {
      allow_rsa_sha2_cert = true
      cipher_algos        = ""
      compression_algos   = ""
      hostkey_algos       = ""
      integrity_algos     = ""
      kex_algos           = ""
    }
    file_storage = {
      store_file = "never"
    }
    file_verification = {
      abort_on_block          = false
      block_invalid_file_down = false
      block_invalid_file_up   = false
      block_show_message      = true
      enable_down             = false
      enable_up               = false
      max_file_size_rejected  = 500
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
    tcp = {
      enable_tcpkeepalive    = false
      tcpkeepalive_interval  = 0
      tcpkeepalive_max_count = 0
    }
    trace = {
      log_all_kbd          = false
      log_group_membership = false
    }
  }
  optionsv12 = {
    general = {
      transformation_rule       = ""
      vault_transformation_rule = ""
    }
    authentication = {
      show_issue_banner = false
    }
    session = {
      inactivity_timeout        = 0
      allow_multi_channels      = false
      force_shell_disconnection = false
      server_keepalive_type     = "none"
      server_keepalive_interval = 0
    }
    trace = {
      log_all_kbd          = false
      log_group_membership = false
    }
    restriction = {
      cmds_compatibility = "cisco"
    }
    algorithms = {
      kex_algos           = "diffie-hellman-group-exchange-sha256,diffie-hellman-group16-sha512,ecdh-sha2-nistp256,ecdh-sha2-nistp384,ecdh-sha2-nistp521,curve25519-sha256,curve25519-sha256@libssh.org,diffie-hellman-group18-sha512"
      cipher_algos        = "chacha20-poly1305@openssh.com,aes128-gcm@openssh.com,aes256-gcm@openssh.com,aes256-ctr,aes192-ctr,aes128-ctr"
      integrity_algos     = "hmac-sha2-256,hmac-sha2-512,hmac-sha2-256-etm@openssh.com,hmac-sha2-512-etm@openssh.com"
      compression_algos   = ""
      hostkey_algos       = "ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,ssh-ed25519,rsa-sha2-256,rsa-sha2-512"
      allow_rsa_sha2_cert = true
      dh_modulus_min_size = 3072
    }
    server_pubkey = {
      server_pubkey_store           = true
      server_pubkey_check           = "1"
      server_access_allowed_message = "0"
      server_pubkey_create_message  = "1"
      server_pubkey_success_message = "0"
      server_pubkey_failure_message = "1"
    }
    startup_scenario = {
      enable      = false
      scenario    = ""
      show_output = true
      timeout     = 10
      ask_startup = false
    }
    tcp = {
      enable_tcpkeepalive    = false
      tcpkeepalive_interval  = 0
      tcpkeepalive_max_count = 0
    }
    proxy = {
      enable     = false
      proxy_type = "None"
      host       = ""
      port       = 0
      login      = ""
      password   = ""
    }
    file_verification = {
      enable_up               = false
      enable_down             = false
      block_invalid_file_up   = false
      block_invalid_file_down = false
      max_file_size_rejected  = 500
      abort_on_block          = false
      block_show_message      = true
    }
    file_storage = {
      store_file = "never"
    }
  }
  optionsRAWTCPIP = {
    nat_redirection = {
      enable = false
      host   = ""
      port   = 0
    }
  }

  options = {
    "8"  = local.optionsv8
    "9"  = local.optionsv9
    "10" = local.optionsv9
    "12" = local.optionsv12
  }
}

data "wallix-bastion_version" "v" {}
resource "wallix-bastion_connection_policy" "testacc_ConnectionPolicy2" {
  connection_policy_name = "testacc_ConnectionPolicy2"
  description            = "testacc ConnectionPolicy2"
  protocol               = "SSH"
  authentication_methods = ["PASSWORD_VAULT"]
  options                = jsonencode(local.options[split(".", data.wallix-bastion_version.v.wab_version)[0]])
}
resource "wallix-bastion_connection_policy" "testacc_ConnectionPolicy" {
  connection_policy_name = "testacc_ConnectionPolicy"
  protocol               = "RAWTCPIP"
  options                = jsonencode(local.optionsRAWTCPIP)
}
`
}

// nolint: lll, nolintlint
func testAccResourceConnectionPolicyUpdate() string {
	return `
locals {
  optionsv8 = {
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
  optionsv9 = {
    algorithms = {
      allow_rsa_sha2_cert = true
      cipher_algos        = ""
      compression_algos   = ""
      hostkey_algos       = ""
      integrity_algos     = ""
      kex_algos           = ""
    }
    file_storage = {
      store_file = "never"
    }
    file_verification = {
      abort_on_block          = false
      block_invalid_file_down = false
      block_invalid_file_up   = false
      block_show_message      = true
      enable_down             = false
      enable_up               = false
      max_file_size_rejected  = 500
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
    tcp = {
      enable_tcpkeepalive    = true
      tcpkeepalive_interval  = 10
      tcpkeepalive_max_count = 10
    }
    trace = {
      log_all_kbd          = false
      log_group_membership = false
    }
  }
  optionsv12 = {
    general = {
      transformation_rule       = ""
      vault_transformation_rule = ""
    }
    authentication = {
      show_issue_banner = false
    }
    session = {
      inactivity_timeout        = 0
      allow_multi_channels      = false
      force_shell_disconnection = false
      server_keepalive_type     = "none"
      server_keepalive_interval = 0
    }
    trace = {
      log_all_kbd          = false
      log_group_membership = false
    }
    restriction = {
      cmds_compatibility = "cisco"
    }
    algorithms = {
      kex_algos           = "diffie-hellman-group-exchange-sha256,diffie-hellman-group16-sha512,ecdh-sha2-nistp256,ecdh-sha2-nistp384,ecdh-sha2-nistp521,curve25519-sha256,curve25519-sha256@libssh.org,diffie-hellman-group18-sha512"
      cipher_algos        = "chacha20-poly1305@openssh.com,aes128-gcm@openssh.com,aes256-gcm@openssh.com,aes256-ctr,aes192-ctr,aes128-ctr"
      integrity_algos     = "hmac-sha2-256,hmac-sha2-512,hmac-sha2-256-etm@openssh.com,hmac-sha2-512-etm@openssh.com"
      compression_algos   = ""
      hostkey_algos       = "ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,ssh-ed25519,rsa-sha2-256,rsa-sha2-512"
      allow_rsa_sha2_cert = true
      dh_modulus_min_size = 3072
    }
    server_pubkey = {
      server_pubkey_store           = true
      server_pubkey_check           = "1"
      server_access_allowed_message = "0"
      server_pubkey_create_message  = "1"
      server_pubkey_success_message = "0"
      server_pubkey_failure_message = "1"
    }
    startup_scenario = {
      enable      = false
      scenario    = ""
      show_output = true
      timeout     = 10
      ask_startup = false
    }
    tcp = {
      enable_tcpkeepalive    = false
      tcpkeepalive_interval  = 0
      tcpkeepalive_max_count = 0
    }
    proxy = {
      enable     = false
      proxy_type = "None"
      host       = ""
      port       = 0
      login      = ""
      password   = ""
    }
    file_verification = {
      enable_up               = false
      enable_down             = false
      block_invalid_file_up   = false
      block_invalid_file_down = false
      max_file_size_rejected  = 500
      abort_on_block          = false
      block_show_message      = true
    }
    file_storage = {
      store_file = "never"
    }
  }
  options = {
    "8"  = local.optionsv8
    "9"  = local.optionsv9
    "10" = local.optionsv9
    "12" = local.optionsv12
  }
}

data "wallix-bastion_version" "v" {}
resource "wallix-bastion_connection_policy" "testacc_ConnectionPolicy2" {
  connection_policy_name = "testacc_ConnectionPolicy2"
  description            = "testacc ConnectionPolicy2"
  protocol               = "SSH"
  authentication_methods = ["PASSWORD_VAULT", "PASSWORD_MAPPING"]
  options                = jsonencode(local.options[split(".", data.wallix-bastion_version.v.wab_version)[0]])
}
`
}
