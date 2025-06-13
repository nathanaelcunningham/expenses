variable "railway_project_name" {
  description = "Name of the Railway project"
  type        = string
  default     = "expenses-app"
}

variable "environment" {
  description = "Environment name (dev, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "prod"], var.environment)
    error_message = "Environment must be either 'dev' or 'prod'."
  }
}

variable "backend_image" {
  description = "Backend Docker image"
  type        = string
  default     = "ghcr.io/nathanaelcunningham/expenses/backend:latest"
}

variable "frontend_image" {
  description = "Frontend Docker image"
  type        = string
  default     = "ghcr.io/nathanaelcunningham/expenses/frontend:latest"
}

variable "db_user" {
  description = "Database username"
  type        = string
  default     = "expenses"
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "expenses"
}