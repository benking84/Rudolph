resource "google_artifact_registry_repository" "rudolph" {
  project       = var.project_id
  location      = var.region
  repository_id = local.name
  format        = "DOCKER"
  description   = "Rudolph API container images."

  depends_on = [google_project_service.services]
}

resource "google_cloud_run_v2_service" "api" {
  project             = var.project_id
  name                = local.name
  location            = var.region
  ingress             = "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"
  deletion_protection = var.deletion_protection
  # The load balancer does not present an identity token. Ingress already
  # blocks every caller except the external Application Load Balancer, and
  # that load balancer requires a client certificate.
  invoker_iam_disabled = true

  template {
    service_account                  = google_service_account.api.email
    timeout                          = "60s"
    max_instance_request_concurrency = 80

    scaling {
      min_instance_count = var.min_instances
      max_instance_count = var.max_instances
    }

    containers {
      image = var.container_image

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      env {
        name  = "GCP_PROJECT"
        value = var.project_id
      }
      env {
        name  = "FIRESTORE_DATABASE"
        value = google_firestore_database.rudolph.name
      }
      env {
        name  = "EVENT_TOPIC"
        value = var.enable_event_export ? google_pubsub_topic.events[0].name : ""
      }
      env {
        name  = "REQUIRE_CLIENT_CERT"
        value = var.require_client_cert_header ? "true" : "false"
      }
    }
  }

  depends_on = [google_project_service.services]
}
