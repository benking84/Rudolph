data "google_project" "current" {
  project_id = var.project_id
}

resource "google_service_account" "api" {
  project      = var.project_id
  account_id   = substr("${var.prefix}-rudolph-api", 0, 30)
  display_name = "Rudolph API (${var.prefix})"
  description  = "Runtime identity for the Rudolph Cloud Run service."

  depends_on = [google_project_service.services]
}

resource "google_project_iam_member" "api_firestore" {
  project = var.project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.api.email}"
}

resource "google_project_iam_member" "api_pubsub" {
  count   = var.enable_event_export ? 1 : 0
  project = var.project_id
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.api.email}"
}

resource "google_project_iam_member" "api_logging" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.api.email}"
}

resource "google_project_iam_member" "api_metrics" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.api.email}"
}

resource "google_secret_manager_secret_iam_member" "api_client_ca" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.client_ca.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.api.email}"
}
