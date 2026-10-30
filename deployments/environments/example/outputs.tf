output "sync_base_url" {
  value = module.santa_api.sync_base_url
}

output "load_balancer_ip" {
  value = module.santa_api.load_balancer_ip
}

output "artifact_registry" {
  value = module.santa_api.artifact_registry
}

output "firestore_database" {
  value = module.santa_api.firestore_database
}

output "events_topic" {
  value = module.santa_api.events_topic
}

output "events_bucket" {
  value = module.santa_api.events_bucket
}

output "service_account" {
  value = module.santa_api.service_account
}

output "certificate_dns_record_name" {
  value = module.santa_api.certificate_dns_record_name
}

output "certificate_dns_record_type" {
  value = module.santa_api.certificate_dns_record_type
}

output "certificate_dns_record_data" {
  value = module.santa_api.certificate_dns_record_data
}
