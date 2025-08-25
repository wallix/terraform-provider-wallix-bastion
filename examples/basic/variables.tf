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