locals {
  services = [
    "run.googleapis.com",
    "firestore.googleapis.com",
    "compute.googleapis.com",
    "certificatemanager.googleapis.com",
    "networksecurity.googleapis.com",
    "secretmanager.googleapis.com",
    "pubsub.googleapis.com",
    "storage.googleapis.com",
    "artifactregistry.googleapis.com",
    "cloudkms.googleapis.com",
    "iam.googleapis.com",
    "dns.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com",
  ]

  firestore_location = var.firestore_location != "" ? var.firestore_location : var.region
  name               = "${var.prefix}-rudolph"
}

resource "google_project_service" "services" {
  for_each = toset(local.services)

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}
