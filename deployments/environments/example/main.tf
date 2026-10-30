module "santa_api" {
  source = "../../terraform_modules/santa_api"

  project_id                  = var.project_id
  region                      = var.region
  prefix                      = var.prefix
  domain_name                 = var.domain_name
  container_image             = var.container_image
  firestore_database_id       = var.firestore_database_id
  firestore_location          = var.firestore_location
  client_ca_pem               = var.client_ca_pem
  client_intermediate_pems    = var.client_intermediate_pems
  dns_managed_zone            = var.dns_managed_zone
  allowed_inbound_cidrs       = var.allowed_inbound_cidrs
  enable_event_export         = var.enable_event_export
  event_retention_days        = var.event_retention_days
  min_instances               = var.min_instances
  max_instances               = var.max_instances
  firestore_delete_protection = var.firestore_delete_protection
  deletion_protection         = var.deletion_protection
  force_destroy_events_bucket = var.force_destroy_events_bucket
  require_client_cert_header  = var.require_client_cert_header
}
