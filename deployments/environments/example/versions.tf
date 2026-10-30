terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }

  # The example environment keeps state on local disk so `terraform init`
  # does not need a pre-existing bucket. Production environments should use
  # a GCS backend. Create the bucket first, then replace this block:
  #
  # backend "gcs" {
  #   bucket = "PROJECT-rudolph-tfstate"
  #   prefix = "rudolph/example"
  # }
  backend "local" {
    path = "terraform.tfstate"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}
