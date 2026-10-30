variable "project_id" {
  type        = string
  description = "GCP project id."
}

variable "region" {
  type        = string
  description = "Primary GCP region."
}

variable "prefix" {
  type        = string
  description = "Short resource name prefix."
}

variable "domain_name" {
  type        = string
  description = "Public Santa sync hostname."
}

variable "container_image" {
  type        = string
  description = "Rudolph API container image."
}

variable "client_ca_pem" {
  type        = string
  description = "PEM client CA trusted by the load balancer."
  sensitive   = true
}

variable "firestore_database_id" {
  type    = string
  default = "rudolph"
}

variable "firestore_location" {
  type    = string
  default = ""
}

variable "client_intermediate_pems" {
  type    = list(string)
  default = []
}

variable "dns_managed_zone" {
  type    = string
  default = ""
}

variable "allowed_inbound_cidrs" {
  type    = list(string)
  default = []
}

variable "enable_event_export" {
  type    = bool
  default = true
}

variable "event_retention_days" {
  type    = number
  default = 365
}

variable "min_instances" {
  type    = number
  default = 0
}

variable "max_instances" {
  type    = number
  default = 20
}

variable "firestore_delete_protection" {
  type    = bool
  default = true
}

variable "deletion_protection" {
  type    = bool
  default = true
}

variable "force_destroy_events_bucket" {
  type    = bool
  default = false
}

variable "require_client_cert_header" {
  type    = bool
  default = true
}
