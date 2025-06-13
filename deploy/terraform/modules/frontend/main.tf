resource "railway_service" "frontend" {
  project_id = var.project_id
  name       = "frontend-${var.environment}"
}

resource "railway_deployment" "frontend" {
  service_id = railway_service.frontend.id
  
  # Use the pre-built Docker image
  image = var.frontend_image
  
  variables = {
    REACT_APP_API_URL = var.backend_url
  }
}

resource "railway_variable" "api_url" {
  service_id = railway_service.frontend.id
  name       = "REACT_APP_API_URL"
  value      = var.backend_url
}