variable "project_id" {
  type        = string
  description = "GCP project id that hosts Rudolph."
}

variable "region" {
  type        = string
  description = "Region for Cloud Run, Firestore, Artifact Registry, KMS, and the serverless NEG. Example: us-central1."
}

variable "prefix" {
  type        = string
  description = "Short lowercase prefix for resource names. Keep it under 20 characters."

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{0,18}$", var.prefix))
    error_message = "prefix must be 1-19 characters, start with a letter, and contain only lowercase letters, numbers, and hyphens."
  }
}

variable "domain_name" {
  type        = string
  description = "Public hostname Santa clients use, for example sync.example.com. A Google-managed certificate is requested for this name."
}

variable "container_image" {
  type        = string
  description = "Fully qualified Artifact Registry image for the Rudolph API, including the tag."
}

variable "firestore_database_id" {
  type        = string
  description = "Firestore Native database id. Use a named database so Rudolph does not take over (default)."
  default     = "rudolph"
}

variable "firestore_location" {
  type        = string
  description = "Firestore location id. Must be compatible with var.region, for example us-central1 or nam5."
  default     = ""
}

variable "client_ca_pem" {
  type        = string
  description = "PEM-encoded CA certificate that signs Santa client certificates. This is the trust anchor for the external HTTPS load balancer."
  sensitive   = true

  validation {
    condition     = can(regex("BEGIN CERTIFICATE", var.client_ca_pem))
    error_message = "client_ca_pem must be a PEM encoded X.509 certificate."
  }
}

variable "client_intermediate_pems" {
  type        = list(string)
  description = "Optional PEM-encoded intermediate certificates between the client leaf and client_ca_pem."
  default     = []
}

variable "dns_managed_zone" {
  type        = string
  description = "Existing Cloud DNS managed zone name (not the DNS name). Leave empty to create the A record and certificate DNS challenge yourself."
  default     = ""
}

variable "allowed_inbound_cidrs" {
  type        = list(string)
  description = "Optional source CIDRs allowed to reach the load balancer. Empty allows any source; mTLS is still required."
  default     = []
}

variable "enable_event_export" {
  type        = bool
  description = "Publish Santa eventupload payloads to Pub/Sub and archive them in Cloud Storage."
  default     = true
}

variable "event_retention_days" {
  type        = number
  description = "Days to retain archived Santa events in Cloud Storage."
  default     = 365
}

variable "min_instances" {
  type        = number
  description = "Minimum Cloud Run instances. Use 1 or more to avoid cold starts on the sync path."
  default     = 0
}

variable "max_instances" {
  type        = number
  description = "Maximum Cloud Run instances."
  default     = 20
}

variable "firestore_delete_protection" {
  type        = bool
  description = "Enable Firestore delete protection."
  default     = true
}

variable "deletion_protection" {
  type        = bool
  description = "Enable Cloud Run deletion protection."
  default     = true
}

variable "force_destroy_events_bucket" {
  type        = bool
  description = "Allow Terraform to delete the events bucket even when it still holds objects."
  default     = false
}

variable "require_client_cert_header" {
  type        = bool
  description = "When true, the API rejects non-health requests that the load balancer did not mark as a verified client certificate."
  default     = true
}
