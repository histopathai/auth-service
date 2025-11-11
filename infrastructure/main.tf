terraform {
  required_version = ">=1.5.0"
  
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
  
  backend "gcs" {
    bucket = "tf-state-histopathai-platform"
    prefix = "services/auth-service"
  }
}

# Platform-infra outputs'unu kullan
data "terraform_remote_state" "platform" {
  backend = "gcs"
  
  config = {
    bucket = "tf-state-histopathai-platform"
    prefix = "platform-infra"
  }
}

locals {
    project_id = data.terraform_remote_state.platform.outputs.project_id
    region     = data.terraform_remote_state.platform.outputs.region
    service_account = data.terraform_remote_state.platform.outputs.auth_service_account_email
}