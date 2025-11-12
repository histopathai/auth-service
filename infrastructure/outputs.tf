output "service_url" {
  description = "The URL of the deployed auth-service"
  value       = google_cloud_run_v2_service.auth_service.uri
}