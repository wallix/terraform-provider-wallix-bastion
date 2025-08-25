terraform {
  required_version = ">= 1.0"
  required_providers {
    wallix-bastion = {
      source  = "wallix/wallix-bastion"
      version = "~> 0.14.0"
    }
  }
}

provider "wallix-bastion" {
  ip          = var.bastion_ip
  token       = var.bastion_token
  api_version = var.api_version
}

# Create a basic user group
resource "wallix-bastion_usergroup" "basic_users" {
  group_name  = "basic-users"
  description = "Basic user group for demonstration"
  timeframes  = ["allthetime"]
}

# Create a basic target group
resource "wallix-bastion_targetgroup" "basic_targets" {
  group_name  = "basic-targets"
  description = "Basic target group for demonstration"
}

# Create a basic authorization
resource "wallix-bastion_authorization" "basic_auth" {
  authorization_name = "basic-authorization"
  description        = "Basic authorization for SSH access"
  user_group         = wallix-bastion_usergroup.basic_users.group_name
  target_group       = wallix-bastion_targetgroup.basic_targets.group_name

  authorize_sessions = true
  subprotocols = [
    "SSH_SHELL_SESSION",
    "SSH_SCP_UP",
    "SSH_SCP_DOWN"
  ]

  is_recorded = true
}