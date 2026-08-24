variable "bastion_ip" {
  description = "IP address or hostname of the WALLIX Bastion"
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