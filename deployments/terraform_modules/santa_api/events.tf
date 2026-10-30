resource "google_pubsub_topic" "events" {
  count   = var.enable_event_export ? 1 : 0
  project = var.project_id
  name    = "${local.name}-events"

  depends_on = [google_project_service.services]
}

resource "google_storage_bucket" "events" {
  count                       = var.enable_event_export ? 1 : 0
  project                     = var.project_id
  name                        = "${var.project_id}-${local.name}-events"
  location                    = var.region
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
  force_destroy               = var.force_destroy_events_bucket

  encryption {
    default_kms_key_name = google_kms_crypto_key.events.id
  }

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = var.event_retention_days
    }
    action {
      type = "Delete"
    }
  }

  depends_on = [
    google_kms_crypto_key_iam_member.gcs,
    google_project_service.services,
  ]
}

resource "google_storage_bucket_iam_member" "pubsub_writer" {
  count  = var.enable_event_export ? 1 : 0
  bucket = google_storage_bucket.events[0].name
  role   = "roles/storage.admin"
  member = "serviceAccount:service-${data.google_project.current.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_pubsub_subscription" "events_gcs" {
  count   = var.enable_event_export ? 1 : 0
  project = var.project_id
  name    = "${local.name}-events-gcs"
  topic   = google_pubsub_topic.events[0].id

  enable_message_ordering = true
  ack_deadline_seconds    = 60

  cloud_storage_config {
    bucket          = google_storage_bucket.events[0].name
    filename_prefix = "santa-events/"
    max_bytes       = 10000000
    max_duration    = "300s"
  }

  depends_on = [
    google_storage_bucket_iam_member.pubsub_writer,
    google_kms_crypto_key_iam_member.pubsub,
  ]
}
