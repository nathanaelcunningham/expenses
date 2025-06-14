output "service_id" {
  description = "Backend service ID"
  value       = railway_service.backend.id
}

output "service_url" {
  description = "Backend service URL"
  value       = "$${{${railway_service.backend.name}.RAILWAY_PRIVATE_DOMAIN}}"
}
