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

