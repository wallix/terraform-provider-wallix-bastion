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

# User group for session sharing
resource "wallix-bastion_usergroup" "session_users" {
  group_name  = "session-sharing-users"
  description = "Users with session sharing capabilities"
  timeframes  = ["allthetime"]
}

# Target group for session sharing
resource "wallix-bastion_targetgroup" "session_targets" {
  group_name  = "session-sharing-targets"
  description = "Targets that support session sharing"
}

# Authorization with view-only session sharing
resource "wallix-bastion_authorization" "session_sharing_view_only" {
  authorization_name = "session-sharing-view-only"
  description        = "Authorization with view-only session sharing"
  user_group         = wallix-bastion_usergroup.session_users.group_name
  target_group       = wallix-bastion_targetgroup.session_targets.group_name

  authorize_sessions        = true
  authorize_session_sharing = true
  session_sharing_mode      = "view_only"

  subprotocols = [
    "RDP",
    "SSH_SHELL_SESSION",
    "VNC"
  ]

  is_recorded = true
  is_critical = false
}

# Authorization with view-control session sharing
resource "wallix-bastion_authorization" "session_sharing_view_control" {
  authorization_name = "session-sharing-view-control"
  description        = "Authorization with view-control session sharing"
  user_group         = wallix-bastion_usergroup.session_users.group_name
  target_group       = wallix-bastion_targetgroup.session_targets.group_name

  authorize_sessions        = true
  authorize_session_sharing = true
  session_sharing_mode      = "view_control"

  subprotocols = [
    "RDP",
    "SSH_SHELL_SESSION"
  ]

  is_recorded = true
  is_critical = true
}