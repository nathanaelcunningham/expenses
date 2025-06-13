terraform {
  required_providers {
    railway = {
      source = "terraform-community-providers/railway"
    }
  }
}

resource "railway_service" "frontend" {
  project_id = var.project_id
  name       = "frontend-${var.environment}"
}


resource "railway_variable" "api_url" {
  environment_id = var.environment_id
  service_id     = railway_service.frontend.id
  name           = "REACT_APP_API_URL"
  value          = var.backend_url
}