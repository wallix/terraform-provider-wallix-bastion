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

# User groups
resource "wallix-bastion_usergroup" "developers" {
  group_name  = "developers"
  description = "Development team members"
  timeframes  = ["business_hours"]
}

resource "wallix-bastion_usergroup" "sysadmins" {
  group_name  = "sysadmins"
  description = "System administrators"
  timeframes  = ["allthetime"]
}

resource "wallix-bastion_usergroup" "security_team" {
  group_name  = "security-team"
  description = "Security team members"
  timeframes  = ["allthetime"]
}

# Target groups
resource "wallix-bastion_targetgroup" "development_servers" {
  group_name  = "development-servers"
  description = "Development environment servers"
}

resource "wallix-bastion_targetgroup" "production_servers" {
  group_name  = "production-servers"
  description = "Production environment servers"
}

resource "wallix-bastion_targetgroup" "database_servers" {
  group_name  = "database-servers"
  description = "Database servers (all environments)"
}

# Developer access to development servers
resource "wallix-bastion_authorization" "dev_to_dev_servers" {
  authorization_name = "developers-to-dev-servers"
  description        = "Developer access to development servers"
  user_group         = wallix-bastion_usergroup.developers.group_name
  target_group       = wallix-bastion_targetgroup.development_servers.group_name

  authorize_sessions = true
  subprotocols = [
    "SSH_SHELL_SESSION",
    "SSH_SCP_UP",
    "SSH_SCP_DOWN",
    "SFTP_SESSION"
  ]

  is_recorded = true
}

# Sysadmin access to production (with approval)
resource "wallix-bastion_authorization" "sysadmin_to_production" {
  authorization_name = "sysadmins-to-production"
  description        = "Sysadmin access to production servers"
  user_group         = wallix-bastion_usergroup.sysadmins.group_name
  target_group       = wallix-bastion_targetgroup.production_servers.group_name

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

  is_critical       = true
  is_recorded       = true
  approval_required = true
  approvers         = [wallix-bastion_usergroup.security_team.group_name]
  active_quorum     = 1
  inactive_quorum   = 2
  approval_timeout  = 1800 # 30 minutes

  has_comment       = true
  mandatory_comment = true
}

# Database access (highly restricted)
resource "wallix-bastion_authorization" "database_access" {
  authorization_name = "database-access"
  description        = "Restricted database access"
  user_group         = wallix-bastion_usergroup.sysadmins.group_name
  target_group       = wallix-bastion_targetgroup.database_servers.group_name

  authorize_sessions = true
  subprotocols = [
    "SSH_SHELL_SESSION"
  ]

  is_critical       = true
  is_recorded       = true
  approval_required = true
  approvers         = [wallix-bastion_usergroup.security_team.group_name]
  active_quorum     = 2
  inactive_quorum   = 3
  approval_timeout  = 3600 # 1 hour

  has_comment       = true
  has_ticket        = true
  mandatory_comment = true
  mandatory_ticket  = true
  single_connection = true
}

# Security team emergency access
resource "wallix-bastion_authorization" "security_emergency_access" {
  authorization_name = "security-emergency-access"
  description        = "Security team emergency access to all systems"
  user_group         = wallix-bastion_usergroup.security_team.group_name
  target_group       = wallix-bastion_targetgroup.production_servers.group_name

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

  is_critical = true
  is_recorded = true
  # No approval required for security team emergency access

  has_comment       = true
  mandatory_comment = true
}