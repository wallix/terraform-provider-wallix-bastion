output "user_group_id" {
  description = "ID of the created user group"
  value       = wallix-bastion_usergroup.basic_users.id
}

output "target_group_id" {
  description = "ID of the created target group"
  value       = wallix-bastion_targetgroup.basic_targets.id
}

output "authorization_id" {
  description = "ID of the created authorization"
  value       = wallix-bastion_authorization.basic_auth.id
}