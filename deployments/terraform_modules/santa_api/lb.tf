resource "google_certificate_manager_trust_config" "client_ca" {
  project  = var.project_id
  name     = "${local.name}-client-ca"
  location = "global"

  trust_stores {
    trust_anchors {
      pem_certificate = var.client_ca_pem
    }

    dynamic "intermediate_cas" {
      for_each = { for i, pem in var.client_intermediate_pems : tostring(i) => pem }
      content {
        pem_certificate = intermediate_cas.value
      }
    }
  }

  depends_on = [google_project_service.services]
}

resource "google_network_security_server_tls_policy" "mtls" {
  project    = var.project_id
  name       = "${local.name}-mtls"
  location   = "global"
  allow_open = false

  mtls_policy {
    client_validation_mode         = "REJECT_INVALID"
    client_validation_trust_config = google_certificate_manager_trust_config.client_ca.id
  }

  depends_on = [google_project_service.services]
}

resource "google_certificate_manager_dns_authorization" "server" {
  project  = var.project_id
  name     = "${local.name}-dns"
  location = "global"
  domain   = var.domain_name

  depends_on = [google_project_service.services]
}

resource "google_certificate_manager_certificate" "server" {
  project  = var.project_id
  name     = "${local.name}-server"
  location = "global"

  managed {
    domains            = [var.domain_name]
    dns_authorizations = [google_certificate_manager_dns_authorization.server.id]
  }
}

resource "google_certificate_manager_certificate_map" "server" {
  project    = var.project_id
  name       = "${local.name}-map"
  depends_on = [google_project_service.services]
}

resource "google_certificate_manager_certificate_map_entry" "server" {
  project      = var.project_id
  name         = "${local.name}-entry"
  map          = google_certificate_manager_certificate_map.server.name
  certificates = [google_certificate_manager_certificate.server.id]
  hostname     = var.domain_name
}

resource "google_compute_global_address" "api" {
  project = var.project_id
  name    = local.name

  depends_on = [google_project_service.services]
}

resource "google_compute_region_network_endpoint_group" "api" {
  project               = var.project_id
  name                  = local.name
  region                = var.region
  network_endpoint_type = "SERVERLESS"

  cloud_run {
    service = google_cloud_run_v2_service.api.name
  }
}

resource "google_compute_security_policy" "api" {
  count   = length(var.allowed_inbound_cidrs) > 0 ? 1 : 0
  project = var.project_id
  name    = local.name
  type    = "CLOUD_ARMOR"

  rule {
    action      = "allow"
    priority    = 1000
    description = "Allowed Santa client networks"
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = var.allowed_inbound_cidrs
      }
    }
  }

  rule {
    action      = "deny(403)"
    priority    = 2147483647
    description = "Default deny"
    match {
      versioned_expr = "SRC_IPS_V1"
      config {
        src_ip_ranges = ["*"]
      }
    }
  }
}

resource "google_compute_backend_service" "api" {
  project               = var.project_id
  name                  = local.name
  load_balancing_scheme = "EXTERNAL_MANAGED"
  protocol              = "HTTP"
  timeout_sec           = 30
  security_policy       = length(var.allowed_inbound_cidrs) > 0 ? google_compute_security_policy.api[0].id : null

  backend {
    group = google_compute_region_network_endpoint_group.api.id
  }

  log_config {
    enable      = true
    sample_rate = 1
  }

  custom_request_headers = [
    "X-Client-Cert-Present:{client_cert_present}",
    "X-Client-Cert-Chain-Verified:{client_cert_chain_verified}",
    "X-Client-Cert-Sha256-Fingerprint:{client_cert_sha256_fingerprint}",
    "X-Client-Cert-Serial-Number:{client_cert_serial_number}",
    "X-Client-Cert-Subject-Dn:{client_cert_subject_dn}",
    "X-Client-Cert-Issuer-Dn:{client_cert_issuer_dn}",
  ]
}

resource "google_compute_url_map" "https" {
  project         = var.project_id
  name            = local.name
  default_service = google_compute_backend_service.api.id
}

resource "google_compute_target_https_proxy" "api" {
  project           = var.project_id
  name              = local.name
  url_map           = google_compute_url_map.https.id
  certificate_map   = "//certificatemanager.googleapis.com/${google_certificate_manager_certificate_map.server.id}"
  server_tls_policy = google_network_security_server_tls_policy.mtls.id
}

resource "google_compute_global_forwarding_rule" "https" {
  project               = var.project_id
  name                  = "${local.name}-https"
  ip_address            = google_compute_global_address.api.id
  port_range            = "443"
  target                = google_compute_target_https_proxy.api.id
  load_balancing_scheme = "EXTERNAL_MANAGED"
}

resource "google_compute_url_map" "http_redirect" {
  project = var.project_id
  name    = "${local.name}-http"

  default_url_redirect {
    https_redirect         = true
    strip_query            = false
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
  }
}

resource "google_compute_target_http_proxy" "http" {
  project = var.project_id
  name    = "${local.name}-http"
  url_map = google_compute_url_map.http_redirect.id
}

resource "google_compute_global_forwarding_rule" "http" {
  project               = var.project_id
  name                  = "${local.name}-http"
  ip_address            = google_compute_global_address.api.id
  port_range            = "80"
  target                = google_compute_target_http_proxy.http.id
  load_balancing_scheme = "EXTERNAL_MANAGED"
}

data "google_dns_managed_zone" "zone" {
  count   = var.dns_managed_zone == "" ? 0 : 1
  project = var.project_id
  name    = var.dns_managed_zone
}

resource "google_dns_record_set" "api" {
  count        = var.dns_managed_zone == "" ? 0 : 1
  project      = var.project_id
  managed_zone = data.google_dns_managed_zone.zone[0].name
  name         = "${var.domain_name}."
  type         = "A"
  ttl          = 300
  rrdatas      = [google_compute_global_address.api.address]
}

resource "google_dns_record_set" "cert_auth" {
  count        = var.dns_managed_zone == "" ? 0 : 1
  project      = var.project_id
  managed_zone = data.google_dns_managed_zone.zone[0].name
  name         = google_certificate_manager_dns_authorization.server.dns_resource_record[0].name
  type         = google_certificate_manager_dns_authorization.server.dns_resource_record[0].type
  ttl          = 300
  rrdatas      = [google_certificate_manager_dns_authorization.server.dns_resource_record[0].data]
}
