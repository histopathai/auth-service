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
    prefix = "platform/prod"
  }
}

locals {
    project_id = data.terraform_remote_state.platform.outputs.project_id
    project_number = data.terraform_remote_state.platform.outputs.project_number
    region     = data.terraform_remote_state.platform.outputs.region
    service_account = data.terraform_remote_state.platform.outputs.auth_service_account_email
    artifact_repository_id = data.terraform_remote_state.platform.outputs.artifact_repository_id
    service_name = var.environment == "prod" ? "auth-service" : "auth-service-${var.environment}"
    image_name = "${local.region}-docker.pkg.dev/${local.project_id}/${local.artifact_repository_id}/${local.service_name}:${var.image_tag}"
    main_service_name = var.environment == "prod" ? "main-service" : "main-service-${var.environment}"
    main_service_url  = "https://${local.main_service_name}-${local.project_number}.${local.region}.run.app"
}

provider "google" {
  project = local.project_id
  region  = local.region
}

# ---------------------------------
# CLOUD RUN SERVICE
# ---------------------------------

resource "google_cloud_run_v2_service" "auth_service" {
    name     = local.service_name
    location = local.region
    ingress  = var.allow_public_access ? "INGRESS_TRAFFIC_ALL" : "INGRESS_TRAFFIC_INTERNAL_ONLY"

    template {
        service_account = local.service_account
    
        scaling {
        min_instance_count = var.min_instances
        max_instance_count = var.max_instances
    }

    containers {
        image = var.image_tag      
        resources {
            limits = {
            cpu    = var.cpu_limit
                memory = var.memory_limit
            }
            cpu_idle = true
        }
        ports {
            container_port = 8080
        }

        env {
            name = "PROJECT_ID"
            value = local.project_id
        }

        env {
            name = "REGION"
            value = local.region
        }

        env {
            name = "PROJECT_NUMBER"
            value = local.project_number
        }

        env {
            name = "ENV"
            value = var.environment
        }

        env {
            name = "GIN_MODE"
            value = var.environment == "prod" ? "release" : "debug"
        }

        env {
            name  = "LOG_LEVEL"
            value = var.log_levels
        }

        env {
            name = "LOG_FORMAT"
            value = "json"
        }

        env {
            name  = "READ_TIMEOUT"
            value = "15"
        }

        env {
            name  = "WRITE_TIMEOUT"
            value = "15"
        }

        env {
            name  = "IDLE_TIMEOUT"
            value = "60"
        }

        env {
            name  = "MAIN_SERVICE_URL"
            value = local.main_service_url
        }

        env {
            name  = "MAIN_SERVICE_NAME"
            value = local.main_service_name
        }
    }
    }
    
    traffic {
        type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
        percent = 100
    }

    labels = {
        environment = var.environment
        service     = "auth-service"
        managed_by  = "terraform"
    }

}

# ---------------------------------
# IAM for Public Access
# ---------------------------------
resource "google_cloud_run_v2_service_iam_member" "public_access" {
  count = var.allow_public_access ? 1 : 0

  project  = google_cloud_run_v2_service.auth_service.project
  location = google_cloud_run_v2_service.auth_service.location
  name     = google_cloud_run_v2_service.auth_service.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
