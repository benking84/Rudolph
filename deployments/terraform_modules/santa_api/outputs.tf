output "sync_base_url" {
  description = "SyncBaseURL value for the Santa configuration profile. Include the trailing slash."
  value       = "https://${var.domain_name}/"
}

output "load_balancer_ip" {
  description = "Global anycast address of the external HTTPS load balancer."
  value       = google_compute_global_address.api.address
}

output "cloud_run_service" {
  description = "Cloud Run service name."
  value       = google_cloud_run_v2_service.api.name
}

output "cloud_run_uri" {
  description = "Cloud Run URI. Direct ingress is disabled; clients must use the load balancer."
  value       = google_cloud_run_v2_service.api.uri
}

output "service_account" {
  description = "Runtime service account email."
  value       = google_service_account.api.email
}

output "firestore_database" {
  description = "Firestore Native database id."
  value       = google_firestore_database.rudolph.name
}

output "artifact_registry" {
  description = "Docker repository that should receive the Rudolph image."
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.rudolph.repository_id}"
}

output "events_topic" {
  description = "Pub/Sub topic that receives Santa execution events. Empty when event export is disabled."
  value       = var.enable_event_export ? google_pubsub_topic.events[0].name : ""
}

output "events_bucket" {
  description = "Cloud Storage bucket that archives Santa execution events."
  value       = var.enable_event_export ? google_storage_bucket.events[0].name : ""
}

output "client_ca_secret" {
  description = "Secret Manager secret that stores the client CA certificate."
  value       = google_secret_manager_secret.client_ca.id
}

output "certificate_dns_record_name" {
  description = "DNS name to publish so Certificate Manager can validate the server certificate."
  value       = google_certificate_manager_dns_authorization.server.dns_resource_record[0].name
}

output "certificate_dns_record_type" {
  description = "DNS record type for the Certificate Manager challenge."
  value       = google_certificate_manager_dns_authorization.server.dns_resource_record[0].type
}

output "certificate_dns_record_data" {
  description = "DNS record data for the Certificate Manager challenge."
  value       = google_certificate_manager_dns_authorization.server.dns_resource_record[0].data
}
