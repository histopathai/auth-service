#!/bin/bash
set -e

echo "Checking if Cloud Run service exists and needs to be imported..."

# Check if Cloud Run service is in state
if ! terraform state show google_cloud_run_v2_service.auth_service >/dev/null 2>&1; then
  echo "Cloud Run service not in state, attempting import..."
  
  # Import the existing Cloud Run service
  terraform import \
    -var-file="${TF_VAR_environment}.tfvars" \
    -var="image_tag=${TF_VAR_image_tag}" \
    -var="tf_state_bucket=${TF_VAR_tf_state_bucket}" \
    google_cloud_run_v2_service.auth_service \
    "projects/${TF_VAR_project_id}/locations/${TF_VAR_region}/services/auth-service-${TF_VAR_environment}" \
    2>&1 || echo "  - Import failed or service doesn't exist (will be created)"
else
  echo "  - Cloud Run service already in state"
fi

echo "State cleanup complete!"
