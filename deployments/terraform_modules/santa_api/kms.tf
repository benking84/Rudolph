resource "google_kms_key_ring" "rudolph" {
  project  = var.project_id
  name     = local.name
  location = var.region

  depends_on = [google_project_service.services]
}

resource "google_kms_crypto_key" "events" {
  name            = "events"
  key_ring        = google_kms_key_ring.rudolph.id
  purpose         = "ENCRYPT_DECRYPT"
  rotation_period = "7776000s"

  lifecycle {
    prevent_destroy = false
  }
}

data "google_storage_project_service_account" "gcs" {
  project = var.project_id
}

resource "google_kms_crypto_key_iam_member" "gcs" {
  crypto_key_id = google_kms_crypto_key.events.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:${data.google_storage_project_service_account.gcs.email_address}"
}

resource "google_kms_crypto_key_iam_member" "pubsub" {
  count         = var.enable_event_export ? 1 : 0
  crypto_key_id = google_kms_crypto_key.events.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.current.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}
