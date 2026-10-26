variable "bastion_ip" {
  description = "IP address or hostname of the WALLIX Bastion"
  type        = string
}

variable "bastion_user" {
  description = "Username for WALLIX Bastion authentication"
  type        = string
}

variable "bastion_token" {
  description = "API token for WALLIX Bastion authentication"
  type        = string
  sensitive   = true
}

variable "api_version" {
  description = "WALLIX Bastion API version"
  type        = string
  default     = "v3.12"
}

variable "active_quorum" {
  description = "Number of approvers needed when they are active"
  type        = number
  default     = 1
}

variable "inactive_quorum" {
  description = "Number of approvers needed when they are inactive"
  type        = number
  default     = 2
}

variable "approval_timeout" {
  description = "Approval timeout in minutes (0: no timeout)"
  type        = number
  default     = 60
}

variable "mandatory_comment" {
  description = "Whether comments are mandatory"
  type        = bool
  default     = true
}

variable "mandatory_ticket" {
  description = "Whether tickets are mandatory"
  type        = bool
  default     = false
}

variable "single_connection" {
  description = "Whether to allow only single connection"
  type        = bool
  default     = false
}