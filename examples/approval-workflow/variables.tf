variable "bastion_ip" {
  description = "IP address or hostname of the Wallix Bastion"
  type        = string
}

variable "bastion_token" {
  description = "API token for Wallix Bastion authentication"
  type        = string
  sensitive   = true
}

variable "api_version" {
  description = "Wallix Bastion API version"
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
  description = "Approval timeout in seconds"
  type        = number
  default     = 3600 # 1 hour
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