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

# Regular users who need approval
resource "wallix-bastion_usergroup" "regular_users" {
  group_name  = "regular-users"
  description = "Regular users requiring approval for critical access"
  timeframes  = ["allthetime"]
}

# Approvers group
resource "wallix-bastion_usergroup" "approvers" {
  group_name  = "security-approvers"
  description = "Security team members who can approve access"
  timeframes  = ["allthetime"]
}

# Critical systems target group
resource "wallix-bastion_targetgroup" "critical_systems" {
  group_name  = "critical-systems"
  description = "Production and critical infrastructure systems"
}

# Authorization requiring approval
resource "wallix-bastion_authorization" "critical_access_approval" {
  authorization_name = "critical-access-with-approval"
  description        = "Critical system access requiring approval"
  user_group         = wallix-bastion_usergroup.regular_users.group_name
  target_group       = wallix-bastion_targetgroup.critical_systems.group_name

  # Session permissions
  authorize_sessions           = true
  authorize_password_retrieval = true
  authorize_session_sharing    = true
  session_sharing_mode         = "view_control"

  subprotocols = [
    "SSH_SHELL_SESSION",
    "SSH_SCP_UP",
    "SSH_SCP_DOWN",
    "RDP",
    "SFTP_SESSION"
  ]

  # Security settings
  is_critical = true
  is_recorded = true

  # Approval workflow configuration
  approval_required = true
  approvers         = [wallix-bastion_usergroup.approvers.group_name]
  active_quorum     = var.active_quorum
  inactive_quorum   = var.inactive_quorum
  approval_timeout  = var.approval_timeout

  # Comment and ticket requirements
  has_comment       = true
  has_ticket        = true
  mandatory_comment = var.mandatory_comment
  mandatory_ticket  = var.mandatory_ticket

  single_connection = var.single_connection
}

# Emergency access without approval (for emergencies)
resource "wallix-bastion_authorization" "emergency_access" {
  authorization_name = "emergency-access"
  description        = "Emergency access for critical situations"
  user_group         = wallix-bastion_usergroup.approvers.group_name
  target_group       = wallix-bastion_targetgroup.critical_systems.group_name

  authorize_sessions = true
  subprotocols = [
    "SSH_SHELL_SESSION",
    "RDP"
  ]

  is_critical       = true
  is_recorded       = true
  approval_required = false # No approval needed for emergency access

  # Still require comments for audit trail
  has_comment       = true
  mandatory_comment = true
}