resource "google_firestore_database" "rudolph" {
  project                           = var.project_id
  name                              = var.firestore_database_id
  location_id                       = local.firestore_location
  type                              = "FIRESTORE_NATIVE"
  delete_protection_state           = var.firestore_delete_protection ? "DELETE_PROTECTION_ENABLED" : "DELETE_PROTECTION_DISABLED"
  point_in_time_recovery_enablement = "POINT_IN_TIME_RECOVERY_ENABLED"
  deletion_policy                   = "ABANDON"

  depends_on = [google_project_service.services]
}

resource "google_firestore_field" "expire_at" {
  project    = var.project_id
  database   = google_firestore_database.rudolph.name
  collection = "items"
  field      = "expireAt"

  ttl_config {}
}

resource "google_firestore_index" "serial" {
  project     = var.project_id
  database    = google_firestore_database.rudolph.name
  collection  = "items"
  query_scope = "COLLECTION_GROUP"

  fields {
    field_path = "SerialNum"
    order      = "ASCENDING"
  }
  fields {
    field_path = "DataType"
    order      = "ASCENDING"
  }
}

resource "google_firestore_index" "machine_id" {
  project     = var.project_id
  database    = google_firestore_database.rudolph.name
  collection  = "items"
  query_scope = "COLLECTION_GROUP"

  fields {
    field_path = "DataType"
    order      = "ASCENDING"
  }
  fields {
    field_path = "MachineID"
    order      = "ASCENDING"
  }
}

resource "google_firestore_index" "scan" {
  project     = var.project_id
  database    = google_firestore_database.rudolph.name
  collection  = "items"
  query_scope = "COLLECTION_GROUP"

  fields {
    field_path = "PK"
    order      = "ASCENDING"
  }
  fields {
    field_path = "SK"
    order      = "ASCENDING"
  }
}

resource "google_firestore_index" "machine_rule_deletion" {
  project     = var.project_id
  database    = google_firestore_database.rudolph.name
  collection  = "items"
  query_scope = "COLLECTION"

  fields {
    field_path = "DeleteOnNextSync"
    order      = "ASCENDING"
  }
  fields {
    field_path = "SK"
    order      = "ASCENDING"
  }
}
