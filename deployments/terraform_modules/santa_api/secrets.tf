resource "google_secret_manager_secret" "client_ca" {
  project   = var.project_id
  secret_id = "${local.name}-client-ca"

  replication {
    auto {}
  }

  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "client_ca" {
  secret      = google_secret_manager_secret.client_ca.id
  secret_data = var.client_ca_pem
}
