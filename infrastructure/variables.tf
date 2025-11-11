variable "project_id" {
    description = "Google Cloud Project ID"
    type        = string
}

variable "project_number" {
    description = "Google Cloud Project Number"
    type        = string
}

variable "region" {
    description = "Google Cloud Region"
    type        = string
    default     = "europe-west1"
}

variable "environment" {
    description = "Environment name (prod, dev)"
    type        = string

    validation {
        condition     = contains(["prod", "dev"], var.environment)
        error_message = "Environment must be either 'prod' or 'dev'."
    }
}

# --------------------------------
# Scaling Configuration
# --------------------------------
variable "min_instances" {
    description = "Minimum number of instances for scaling"
    type        = number
    default     = 1
}

variable "max_instances" {
    description = "Maximum number of instances for scaling"
    type        = number
    default     = 2
}

# --------------------------------
# Resource Configuration
# --------------------------------
variable "cpu_limit" {
    description = "CPU limit for each instance"
    type        = string
    default     = "1"
}

variable "memory_limit" {
    description = "Memory limit for each instance"
    type        = string
    default     = "512Mi"
}

# --------------------------------
# Access Configuration
# --------------------------------
variable "allow_public_access" {
  description = "Allow public access to the service"
  type        = bool
  default     = true
}

variable "log_levels" {
    description = "Log level (debug, info, warn, error)"
    type        = string
    default     = "info"

    validation {
        condition     = contains(["debug", "info", "warn", "error"], var.log_levels)
        error_message = "Log level must be one of 'debug', 'info', 'warn', or 'error'."
    }
}


